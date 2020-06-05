package sinks

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/zh-jn/logina"
)

var (
	LogMaxFileSize = 50 * 1024 * 1024 //
	LogMaxFileNum  = 20               // 1 ~ 99
)

type RotateFile struct {
	writer io.Writer
	name   string
	size   int64
	file   *os.File
	mu     sync.Mutex
}

func NewRotateFile(logFile string) (*RotateFile, error) {
	log, err := os.OpenFile(logFile, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		fmt.Println("OpenFile fail:", err)
		return nil, err
	}

	fs, err := log.Stat()
	if err != nil {
		fmt.Println("Stat fail:", err)
		return nil, err
	}

	return &RotateFile{
		writer: log,
		name:   logFile,
		size:   fs.Size(),
		file:   log,
	}, nil
}

func (hook *RotateFile) Rotate() bool {
	lastName := fmt.Sprintf("%s.%02d", hook.name, LogMaxFileNum)
	os.Remove(lastName)
	var fileName string
	for index := LogMaxFileNum - 1; index > 0; index-- {
		fileName = fmt.Sprintf("%s.%02d", hook.name, index)
		os.Rename(fileName, lastName)
		lastName = fileName
	}
	os.Rename(hook.name, fileName)
	log, err := os.OpenFile(hook.name, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		fmt.Println("OpenFile fail:", err)
		return false
	}
	hook.file.Close()
	hook.writer = log
	hook.file = log
	hook.size = 0
	return true
}

func (hook *RotateFile) Fire(entry *logina.Entry) error {
	writer := hook.writer
	msg := fmt.Sprintf("[%s] time=%v message=%v\n", strings.ToUpper(entry.Level.String()), entry.Time, entry.Message)
	n, err := writer.Write([]byte(msg))
	if err != nil {
		return err
	}
	hook.mu.Lock()
	defer hook.mu.Unlock()
	if hook.size += int64(n); hook.size >= int64(LogMaxFileSize) {
		hook.Rotate()
	}
	return nil
}
