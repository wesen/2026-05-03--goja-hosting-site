package dbguard

import "strings"

type SQLKind string

const (
	SQLKindRead        SQLKind = "read"
	SQLKindGrowth      SQLKind = "growth"
	SQLKindCleanup     SQLKind = "cleanup"
	SQLKindMaintenance SQLKind = "maintenance"
	SQLKindUnknown     SQLKind = "unknown"
)

func classifySQL(query string) SQLKind {
	token := firstSQLToken(query)
	switch token {
	case "SELECT", "WITH":
		return SQLKindRead
	case "INSERT", "UPDATE", "CREATE", "ALTER", "REPLACE":
		return SQLKindGrowth
	case "DELETE", "DROP", "TRUNCATE":
		return SQLKindCleanup
	case "VACUUM", "PRAGMA", "ANALYZE", "REINDEX":
		return SQLKindMaintenance
	case "":
		return SQLKindUnknown
	default:
		return SQLKindUnknown
	}
}

func growthBlockedKind(kind SQLKind) bool {
	return kind == SQLKindGrowth || kind == SQLKindUnknown
}

func firstSQLToken(query string) string {
	s := strings.TrimSpace(query)
	for {
		if strings.HasPrefix(s, "--") {
			if idx := strings.IndexByte(s, '\n'); idx >= 0 {
				s = strings.TrimSpace(s[idx+1:])
				continue
			}
			return ""
		}
		if strings.HasPrefix(s, "/*") {
			if idx := strings.Index(s, "*/"); idx >= 0 {
				s = strings.TrimSpace(s[idx+2:])
				continue
			}
			return ""
		}
		break
	}
	if s == "" {
		return ""
	}
	for i, r := range s {
		if !(r == '_' || r == '-' || r >= '0' && r <= '9' || r >= 'A' && r <= 'Z' || r >= 'a' && r <= 'z') {
			return strings.ToUpper(s[:i])
		}
	}
	return strings.ToUpper(s)
}
