package logina

// target where the log entry sent to
type Sink interface {
	Fire(*Entry) error
}
