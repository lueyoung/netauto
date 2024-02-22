package core

import (
	"log"
	"os"
)

var (
	debug = createDebug("/tmp/debug.log")
)

func createDebug(path string) *log.Logger {
	f, err := os.Create(path)
	if err != nil {
		log.Fatal(err)
	}
	return log.New(f, "", log.LstdFlags)
}

/**
func prepareDebug(path string) {
	f, err := os.Create(path)
	if err != nil {
		log.Fatal(err)
	}
	logger := log.New(f, "", log.LstdFlags)
	for {
		select {
		case msg := <-debug:
			logger.Println(msg)
		}
	}
}**/
