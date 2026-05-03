package dbguard

import (
	"database/sql"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/dop251/goja"
)

type Guard struct {
	mu       sync.Mutex
	db       *sql.DB
	path     string
	options  Options
	vm       *goja.Runtime
	callback goja.Callable

	writeCount     int64
	cleanupAttempt int64
	lastCleanup    time.Time
	inCleanup      bool
	lastStats      Stats
	lastResult     CheckResult
}

func New(db *sql.DB, path string) *Guard {
	opts := Options{Path: path, Cooldown: 30 * time.Second, CheckEveryWrites: 10, IncludeWAL: true}
	return &Guard{db: db, path: path, options: opts}
}

func (g *Guard) Configure(opts Options) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if opts.Path != "" {
		g.path = opts.Path
		g.options.Path = opts.Path
	}
	if opts.MaxBytes != 0 {
		g.options.MaxBytes = opts.MaxBytes
	}
	if opts.SoftMaxBytes != 0 {
		g.options.SoftMaxBytes = opts.SoftMaxBytes
	}
	if opts.HardMaxBytes != 0 {
		g.options.HardMaxBytes = opts.HardMaxBytes
	}
	if opts.Cooldown != 0 {
		g.options.Cooldown = opts.Cooldown
	}
	if opts.CheckEveryWrites != 0 {
		g.options.CheckEveryWrites = opts.CheckEveryWrites
	}
	g.options.IncludeWAL = opts.IncludeWAL
	g.options.FailOverHardLimit = opts.FailOverHardLimit
}

func (g *Guard) SetRuntime(vm *goja.Runtime) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.vm = vm
}

func (g *Guard) SetCallback(fn goja.Callable) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.callback = fn
}

func (g *Guard) Stats() (Stats, error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	stats, err := g.measureLocked()
	if err == nil {
		g.lastStats = stats
	}
	return stats, err
}

func (g *Guard) IsOverLimit() bool {
	stats, err := g.Stats()
	if err != nil {
		return false
	}
	g.mu.Lock()
	limit := g.limitBytesLocked()
	g.mu.Unlock()
	return limit > 0 && stats.TotalBytes > limit
}

func (g *Guard) BeforeExec(query string) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.hardLimitErrorLocked("before exec", query, classifySQL(query), false)
}

func (g *Guard) ErrorAfterExec(query string, result CheckResult) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.hardLimitErrorLocked("after exec", query, classifySQL(query), true)
}

func (g *Guard) AfterExec(query string) (CheckResult, error) {
	g.mu.Lock()
	g.writeCount++
	if g.inCleanup {
		res := CheckResult{Reason: "afterExec", SkippedReason: "cleanup already running", OriginalQuery: query, SQLKind: string(classifySQL(query))}
		g.mu.Unlock()
		return res, nil
	}
	if g.limitBytesLocked() <= 0 {
		res := CheckResult{Reason: "afterExec", SkippedReason: "no limit configured", OriginalQuery: query, SQLKind: string(classifySQL(query))}
		g.mu.Unlock()
		return res, nil
	}
	if g.options.CheckEveryWrites > 0 && g.writeCount%g.options.CheckEveryWrites != 0 {
		res := CheckResult{Reason: "afterExec", SkippedReason: "write counter throttle", OriginalQuery: query, SQLKind: string(classifySQL(query))}
		g.mu.Unlock()
		return res, nil
	}
	if !g.lastCleanup.IsZero() && g.options.Cooldown > 0 && time.Since(g.lastCleanup) < g.options.Cooldown {
		res := CheckResult{Reason: "afterExec", SkippedReason: "cooldown", OriginalQuery: query, SQLKind: string(classifySQL(query))}
		g.mu.Unlock()
		return res, nil
	}
	g.mu.Unlock()
	return g.CheckNow("afterExec", query)
}

func (g *Guard) CheckNow(reason string, originalQuery ...string) (CheckResult, error) {
	if reason == "" {
		reason = "manual"
	}
	query := ""
	if len(originalQuery) > 0 {
		query = originalQuery[0]
	}

	g.mu.Lock()
	before, err := g.measureLocked()
	if err != nil {
		g.mu.Unlock()
		return CheckResult{Reason: reason, OriginalQuery: query}, err
	}
	g.lastStats = before
	limit := g.limitBytesLocked()
	res := CheckResult{Reason: reason, Before: before, OriginalQuery: query, SQLKind: string(classifySQL(query))}
	if limit <= 0 {
		res.SkippedReason = "no limit configured"
		g.lastResult = res
		g.mu.Unlock()
		return res, nil
	}
	if before.TotalBytes <= limit {
		res.SkippedReason = "under limit"
		g.lastResult = res
		g.mu.Unlock()
		return res, nil
	}
	res.Triggered = true
	res.FailHardLimitHit = g.options.HardMaxBytes > 0 && before.TotalBytes > g.options.HardMaxBytes
	if g.callback == nil || g.vm == nil {
		res.SkippedReason = "no callback registered"
		res.StillOverLimit = true
		g.lastResult = res
		g.mu.Unlock()
		return res, nil
	}
	if g.inCleanup {
		res.SkippedReason = "cleanup already running"
		res.StillOverLimit = true
		g.lastResult = res
		g.mu.Unlock()
		return res, nil
	}
	g.inCleanup = true
	g.cleanupAttempt++
	res.CleanupAttempt = g.cleanupAttempt
	callback := g.callback
	vm := g.vm
	event := g.eventMapLocked(reason, query, before, res.CleanupAttempt)
	g.mu.Unlock()

	result, callErr := callback(goja.Undefined(), vm.ToValue(event))
	if callErr != nil {
		res.CallbackError = callErr.Error()
	} else {
		res.CallbackCalled = true
		if !goja.IsUndefined(result) && !goja.IsNull(result) {
			res.CallbackResult = result.Export()
		}
	}

	g.mu.Lock()
	after, measureErr := g.measureLocked()
	if measureErr == nil {
		res.After = &after
		res.ReducedByBytes = before.TotalBytes - after.TotalBytes
		res.StillOverLimit = after.TotalBytes > limit
		g.lastStats = after
	}
	g.inCleanup = false
	g.lastCleanup = time.Now()
	g.lastResult = res
	g.mu.Unlock()
	if measureErr != nil {
		return res, measureErr
	}
	return res, nil
}

func (g *Guard) LastResult() CheckResult {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.lastResult
}

func (g *Guard) eventMapLocked(reason, query string, stats Stats, attempt int64) map[string]any {
	limit := g.limitBytesLocked()
	return map[string]any{
		"reason":         reason,
		"query":          query,
		"stats":          stats.Map(),
		"thresholds":     map[string]any{"maxBytes": limit, "softMaxBytes": g.options.SoftMaxBytes, "hardMaxBytes": g.options.HardMaxBytes},
		"overByBytes":    stats.TotalBytes - limit,
		"cleanupAttempt": attempt,
		"inCallback":     true,
	}
}

func (g *Guard) measureLocked() (Stats, error) {
	stats := Stats{Path: g.path, MaxBytes: g.options.MaxBytes, SoftMaxBytes: g.options.SoftMaxBytes, HardMaxBytes: g.options.HardMaxBytes, CheckedAt: time.Now()}
	var err error
	stats.FileBytes, err = fileSize(g.path)
	if err != nil {
		return stats, err
	}
	if g.options.IncludeWAL {
		stats.WALBytes, _ = fileSize(g.path + "-wal")
		stats.SHMBytes, _ = fileSize(g.path + "-shm")
	}
	stats.TotalBytes = stats.FileBytes + stats.WALBytes + stats.SHMBytes
	limit := g.limitBytesLocked()
	stats.MaxBytes = limit
	if limit > 0 && stats.TotalBytes > limit {
		stats.OverByBytes = stats.TotalBytes - limit
	}
	if g.db != nil {
		_ = g.db.QueryRow("PRAGMA page_size").Scan(&stats.PageSize)
		_ = g.db.QueryRow("PRAGMA page_count").Scan(&stats.PageCount)
		_ = g.db.QueryRow("PRAGMA freelist_count").Scan(&stats.FreeListCount)
		if stats.PageSize > 0 && stats.PageCount >= stats.FreeListCount {
			stats.EstimatedLiveBytes = stats.PageSize * (stats.PageCount - stats.FreeListCount)
		}
	}
	return stats, nil
}

func (g *Guard) hardLimitErrorLocked(phase, query string, kind SQLKind, afterExec bool) error {
	if !g.options.FailOverHardLimit || g.options.HardMaxBytes <= 0 {
		return nil
	}
	if g.inCleanup {
		return nil
	}
	if !growthBlockedKind(kind) {
		return nil
	}
	stats, err := g.measureLocked()
	if err != nil {
		return err
	}
	g.lastStats = stats
	if stats.TotalBytes <= g.options.HardMaxBytes {
		return nil
	}
	return &HardLimitError{Phase: phase, Path: g.path, Query: query, Kind: kind, TotalBytes: stats.TotalBytes, HardMaxBytes: g.options.HardMaxBytes, AfterExec: afterExec}
}

func (g *Guard) limitBytesLocked() int64 {
	if g.options.SoftMaxBytes > 0 {
		return g.options.SoftMaxBytes
	}
	return g.options.MaxBytes
}

func fileSize(path string) (int64, error) {
	if path == "" {
		return 0, fmt.Errorf("database path is required")
	}
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, err
	}
	return info.Size(), nil
}
