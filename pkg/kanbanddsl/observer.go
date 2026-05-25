package kanbanddsl

import "time"

// Observer receives bounded Kanban runtime measurements. Implementations must be
// quick and must not panic; Mount calls observers from request handlers.
type Observer interface {
	ObserveFragment(board string, duration time.Duration, err error)
	ObserveAction(board, action string, refresh bool, duration time.Duration, err error)
	ObserveDispatch(board, action string, duration time.Duration, err error)
	ObserveRender(board, reason string, duration time.Duration, htmlBytes int, err error)
}
