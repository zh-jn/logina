package logina

type LogDegradationStrategy int

const (
	LogDegradationStrategy_DiscardNewest LogDegradationStrategy = iota
	LogDegradationStrategy_DiscardOldest
	LogDegradationStrategy_Synchronize
	LogDegradationStrategy_Hybrid
)
