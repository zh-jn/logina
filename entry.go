package logina

import (
	"fmt"
	"time"
)

// logrus like Entry
type Entry struct {
	Logger *Logger

	Time time.Time

	Level Level

	Message string
}

func NewEntry(logger *Logger) *Entry {
	return &Entry{
		Logger: logger,
	}
}

func (entry *Entry) Logf(level Level, format string, args ...interface{}) {
	entry.Level = level
	entry.Message = fmt.Sprintf(format, args...)
	entry.Time = time.Now()
}

func (entry *Entry) Clean() {

}
