package dbguard

import "time"

type Options struct {
	Path              string
	MaxBytes          int64
	SoftMaxBytes      int64
	HardMaxBytes      int64
	Cooldown          time.Duration
	CheckEveryWrites  int64
	IncludeWAL        bool
	FailOverHardLimit bool
}

type Stats struct {
	Path               string    `json:"path"`
	FileBytes          int64     `json:"fileBytes"`
	WALBytes           int64     `json:"walBytes"`
	SHMBytes           int64     `json:"shmBytes"`
	TotalBytes         int64     `json:"totalBytes"`
	MaxBytes           int64     `json:"maxBytes"`
	SoftMaxBytes       int64     `json:"softMaxBytes"`
	HardMaxBytes       int64     `json:"hardMaxBytes"`
	OverByBytes        int64     `json:"overByBytes"`
	PageSize           int64     `json:"pageSize"`
	PageCount          int64     `json:"pageCount"`
	FreeListCount      int64     `json:"freelistCount"`
	EstimatedLiveBytes int64     `json:"estimatedLiveBytes"`
	CheckedAt          time.Time `json:"checkedAt"`
}

type CheckResult struct {
	Reason           string `json:"reason"`
	Triggered        bool   `json:"triggered"`
	CallbackCalled   bool   `json:"callbackCalled"`
	CallbackError    string `json:"callbackError"`
	CallbackResult   any    `json:"callbackResult"`
	Before           Stats  `json:"before"`
	After            *Stats `json:"after,omitempty"`
	ReducedByBytes   int64  `json:"reducedByBytes"`
	StillOverLimit   bool   `json:"stillOverLimit"`
	SkippedReason    string `json:"skippedReason"`
	CleanupAttempt   int64  `json:"cleanupAttempt"`
	OriginalQuery    string `json:"originalQuery,omitempty"`
	SQLKind          string `json:"sqlKind,omitempty"`
	FailHardLimitHit bool   `json:"failHardLimitHit"`
}

func (s Stats) Map() map[string]any {
	return map[string]any{
		"path":               s.Path,
		"fileBytes":          s.FileBytes,
		"walBytes":           s.WALBytes,
		"shmBytes":           s.SHMBytes,
		"totalBytes":         s.TotalBytes,
		"maxBytes":           s.MaxBytes,
		"softMaxBytes":       s.SoftMaxBytes,
		"hardMaxBytes":       s.HardMaxBytes,
		"overByBytes":        s.OverByBytes,
		"pageSize":           s.PageSize,
		"pageCount":          s.PageCount,
		"freelistCount":      s.FreeListCount,
		"estimatedLiveBytes": s.EstimatedLiveBytes,
		"checkedAt":          s.CheckedAt.Format(time.RFC3339Nano),
	}
}

func (r CheckResult) Map() map[string]any {
	m := map[string]any{
		"reason":           r.Reason,
		"triggered":        r.Triggered,
		"callbackCalled":   r.CallbackCalled,
		"callbackError":    r.CallbackError,
		"callbackResult":   r.CallbackResult,
		"before":           r.Before.Map(),
		"reducedByBytes":   r.ReducedByBytes,
		"stillOverLimit":   r.StillOverLimit,
		"skippedReason":    r.SkippedReason,
		"cleanupAttempt":   r.CleanupAttempt,
		"originalQuery":    r.OriginalQuery,
		"sqlKind":          r.SQLKind,
		"failHardLimitHit": r.FailHardLimitHit,
	}
	if r.After != nil {
		m["after"] = r.After.Map()
	}
	return m
}
