package logger

import (
	"fmt"
	"log"
	"os"
)

const DEFAULT_LOGFILE = "application.log"

func New(logFile string) (*log.Logger, func(), error) {
	log.Println("logfile name received:", logFile)
	if logFile == "" {
		logFile = DEFAULT_LOGFILE
		fmt.Println("Inside if:")
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
	logger := log.New(
		loggingFile,
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
