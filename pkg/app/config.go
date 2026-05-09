package app

// Config describes one goja-site server process.
type Config struct {
	Addr       string
	DBPath     string
	ScriptDirs []string
	Dev        bool
}
