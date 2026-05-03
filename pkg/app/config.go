package app

// Config describes one goja-site server process.
type Config struct {
	Addr       string
	DBPath     string
	ScriptsDir string
	Dev        bool
}
