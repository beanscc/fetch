package fetch

import (
	"context"
	"log"
	"os"
	"sync"
)

type Log interface {
	WithContext(ctx context.Context) Log

	Errorf(format string, args ...interface{})

	Infof(format string, args ...interface{})
}

var loggerMut sync.Mutex
var logger Log = &stdLog{logger: log.New(os.Stdout, "", log.LstdFlags)}

func SetLogger(l Log) {
	loggerMut.Lock()
	defer loggerMut.Unlock()
	logger = l
}

type stdLog struct {
	ctx    context.Context
	logger *log.Logger
}

func (l *stdLog) WithContext(ctx context.Context) Log {
	nl := new(stdLog)
	*nl = *l
	nl.ctx = ctx
	return nl
}

func (l stdLog) Infof(format string, args ...interface{}) {
	l.logger.Printf("[info] "+format, args...)
}

func (l stdLog) Errorf(format string, args ...interface{}) {
	l.logger.Printf("[error] "+format, args...)
}
