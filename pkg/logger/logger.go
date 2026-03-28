package logger

import (
	"io"
	"log"
	"os"

	"gopkg.in/natefinch/lumberjack.v2"
)

func SetupLogger(filename string) (*log.Logger, io.WriteCloser) {
	logRotator := &lumberjack.Logger{
		Filename:   filename,
		MaxSize:    10,
		MaxBackups: 5,
		MaxAge:     30,
		Compress:   true,
		LocalTime:  true,
	}

	multiWriter := io.MultiWriter(os.Stdout, logRotator)

	logger := log.New(multiWriter, "["+os.Getenv("NAME")+"] ", log.LstdFlags|log.Lshortfile)

	return logger, logRotator
}
