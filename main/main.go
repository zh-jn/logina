package main

import (
	"fmt"
	"time"

	"github.com/zh-jn/logina"
	"github.com/zh-jn/logina/sinks"
)

func main() {
	fileSink, err := sinks.NewRotateFile("/data/log/logina")
	if err != nil {
		fmt.Println(err)
		return
	}
	logger := logina.New(logina.DebugLevel, []logina.Sink{fileSink}, logina.LogDegradationStrategy_Hybrid, 0.3)
	for i := 0; i < 1000000; i ++ {
		logger.Infof("willtest %v", i)
	}
	time.Sleep(1 * time.Second)
	return
}