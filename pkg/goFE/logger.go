package goFE

// Logger in the goFE package is a simple logger that uses println instead of fmt for
// a small wasm binary size.
// The level is hierarchical, with the highest level being the most verbose.
type Logger struct {
	Level int
}

const (
	ERROR = iota
	WARNING
	INFO
	DEBUG
)

func (l *Logger) Log(level int, message string) {
	if l.Level >= level {
		println(message)
	}
}
