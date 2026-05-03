package dbguard

import "fmt"

type HardLimitError struct {
	Phase        string
	Path         string
	Query        string
	Kind         SQLKind
	TotalBytes   int64
	HardMaxBytes int64
	AfterExec    bool
}

func (e *HardLimitError) Error() string {
	return fmt.Sprintf("sqlite hard limit exceeded %s: totalBytes=%d hardMaxBytes=%d kind=%s path=%s", e.Phase, e.TotalBytes, e.HardMaxBytes, e.Kind, e.Path)
}
