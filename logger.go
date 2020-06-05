package logina

import (
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

type Logger struct {
	Targets []Sink

	// Flag for whether to log caller info (off by default)
	ReportCaller bool

	Level Level

	queue *EntryQueue

	entryPool sync.Pool

	sendChan chan *Entry

	strategy LogDegradationStrategy

	strategyParam []float64
}

func New(level Level, targets []Sink, strategy LogDegradationStrategy, strategyParam ...float64) *Logger {
	var queueStrategy OverflowStrategy
	switch strategy {
	case LogDegradationStrategy_DiscardNewest:
		queueStrategy = OverflowStrategy_DiscardNewest
	case LogDegradationStrategy_DiscardOldest:
		queueStrategy = OverflowStrategy_DiscardOldest
	default:
		queueStrategy = OverflowStrategy_Unknown
	}
	ret := &Logger{
		Targets:       targets,
		Level:         level,
		queue:         NewQueue(queueStrategy),
		sendChan:      make(chan *Entry, 100),
		strategy:      strategy,
		strategyParam: strategyParam,
	}
	if ret.strategy == LogDegradationStrategy_Hybrid && len(strategyParam) == 0 {
		ret.strategyParam = append(ret.strategyParam, 0.5)
	}
	go ret.logProducer()
	go ret.logConsumer()
	return ret
}

func (logger *Logger) Infof(format string, args ...interface{}) {
	logger.Logf(InfoLevel, format, args...)
}

func (logger *Logger) Logf(level Level, format string, args ...interface{}) {
	if logger.IsLevelEnabled(level) {
		entry := logger.newEntry()
		entry.Logf(level, format, args...)
		res := logger.queue.Push(entry)
		switch logger.strategy {
		case LogDegradationStrategy_DiscardNewest:
		case LogDegradationStrategy_DiscardOldest:
		case LogDegradationStrategy_Synchronize:
			if res == PushResult_Fail {
				logger.sendChan <- entry
			}
		case LogDegradationStrategy_Hybrid:
			r := rand.Float64()
			if r > logger.strategyParam[0] {
				logger.sendChan <- entry
			}
		}
	}
}

// IsLevelEnabled checks if the log level of the logger is greater than the level param
func (logger *Logger) IsLevelEnabled(level Level) bool {
	return logger.level() >= level
}

func (logger *Logger) level() Level {
	return Level(atomic.LoadUint32((*uint32)(&logger.Level)))
}

func (logger *Logger) newEntry() *Entry {
	entry, ok := logger.entryPool.Get().(*Entry)
	if ok {
		return entry
	}
	return NewEntry(logger)
}

func (logger *Logger) releaseEntry(entry *Entry) {
	entry.Clean()
	logger.entryPool.Put(entry)
}

func (logger *Logger) logProducer() {
	for {
		entry := logger.queue.Pop()
		if entry != nil {
			logger.sendChan <- entry
			continue
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func (logger *Logger) logConsumer() {
	for entry := range logger.sendChan {
		for _, s := range logger.Targets {
			s.Fire(entry)
		}
		logger.releaseEntry(entry)
	}
}
