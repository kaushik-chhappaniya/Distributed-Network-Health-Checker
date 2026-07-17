package logger

import (
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
)

/*
runLogProcessor acts as the centralized I/O manager. It consumes messages from the logChan
and routes them: status logs go to the primary logger, while raw ping output is written
to the output file. This function is the "Exclusive Writer," preventing race conditions
by ensuring no other goroutine touches the output file.
*/
func LoggerGoroutine(logChan <-chan string, logWg *sync.WaitGroup, log *log.Logger, outputFileName string) {

	defer logWg.Done()
	log.Println("<WORKER> Inside loggerGoroutine. worker started.")
	// Open the ping output file for writing
	// if _, err := file.Write(fmt.Appendf(nil, "%s : %s \n\n", node.Name, output)); err != nil {
	// if _, err := fmt.Fprintf(file, fmt.Sprintf("%s : %s \n\n", node.Name, output)); err != nil {
	file, err := os.OpenFile(outputFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println("<ERROR> Could not open ping output file")
		return
	}
	defer file.Close()

	for msg := range logChan {
		if strings.HasPrefix(msg, "<ERROR>") || strings.HasPrefix(msg, "<INFO>") || strings.HasPrefix(msg, "<WORKER>") {
			log.Printf("%s\n", msg)
			// fmt.Fprintf(file, "%s\n", msg)
		} else {
			fmt.Fprintln(file, msg)
		}
	}
}
