package observability

import (
	"net/http"
	"strconv"
	"strings"
)

func SiteLabel(site string) string {
	site = strings.TrimSpace(site)
	if site == "" {
		return "default"
	}
	return site
}

func StatusClass(status int) string {
	if status <= 0 {
		return "unknown"
	}
	return strconv.Itoa(status/100) + "xx"
}

// CoarseRoute returns a bounded route label. It intentionally does not expose
// raw paths because paths may contain user-controlled IDs or search terms.
func CoarseRoute(path string) string {
	switch {
	case path == "":
		return "/"
	case path == "/":
		return "/"
	case path == "/healthz" || path == "/readyz":
		return path
	case path == "/_kanban/client.js":
		return "/_kanban/client.js"
	case strings.HasPrefix(path, "/_kanban/") && strings.Contains(path, "/fragment"):
		return "/_kanban/:board/fragment"
	case strings.HasPrefix(path, "/_kanban/") && strings.Contains(path, "/action/"):
		return "/_kanban/:board/action/:action"
	case strings.HasPrefix(path, "/assets/"):
		return "/assets/*"
	default:
		return "other"
	}
}

func MethodLabel(method string) string {
	method = strings.ToUpper(strings.TrimSpace(method))
	if method == "" {
		return http.MethodGet
	}
	return method
}
