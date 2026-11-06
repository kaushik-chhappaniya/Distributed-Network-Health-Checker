package logger

import (
	"io"
	"log"
	"os"
)

const DEFAULT_LOGFILE = "application.log"

type Service struct {
	Logger *log.Logger
}

func New(logFile string, writeToConsole bool) (*log.Logger, func(), error) {
	if logFile == "" {
		logFile = DEFAULT_LOGFILE
	}

	// Open or create the log file
	logFile = "./logs/" + logFile
	loggingFile, err := os.OpenFile(logFile,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644,
	)

	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}

	// log.Printf("Creating logger instance with file : %s.", logFile)
	writer := io.MultiWriter(loggingFile)
	if writeToConsole {
		log.SetOutput(os.Stdout)
		writer = io.MultiWriter(os.Stdout, loggingFile)
	}
	logger := log.New(
		writer,
		"",
		log.Ldate|log.Ltime|log.Lshortfile,
	)

	closeFn := func() {
		logger.Println("----------------------- Logging Completed -----------------------")
		loggingFile.Close()
	}

	logger.Println("----------------------- Logging Started -----------------------")
	return logger, closeFn, nil
}
