package log

// logger log interface
type logger interface {
	Printf(format string, args ...interface{})
}
