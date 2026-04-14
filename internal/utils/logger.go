package utils

import (
	"fmt"
	"os"
	"time"
)

func LogDebug(format string, args ...interface{}) {
	f, err := os.OpenFile("debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	msg := fmt.Sprintf(format, args...)
	if len(msg) > 0 && msg[len(msg)-1] != '\n' {
		msg += "\n"
	}
	fmt.Fprintf(f, "[%s] %s", timestamp, msg)
}
