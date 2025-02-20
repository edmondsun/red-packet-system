package logger

import (
	"log"
	"os"
	"sync"
)

var (
	instance *log.Logger
	once     sync.Once
)

func GetLogger() *log.Logger {
	once.Do(func() {
		instance = log.New(os.Stdout, "[LOG] ", log.Ldate|log.Ltime|log.Lshortfile)
	})
	return instance
}
