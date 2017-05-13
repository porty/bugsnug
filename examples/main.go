package main

import (
	"log"
	"os"

	"github.com/pkg/errors"

	"github.com/porty/bugsnug"
)

func doStuffToFile() error {
	f, err := os.Open("fart")
	if err != nil {
		return errors.Wrap(err, "failed top open file")
	}
	return f.Close()
}

func thing1() error {
	if err := doStuffToFile(); err != nil {
		return errors.Wrap(err, "failed to doStuffToFile")
	}
	return nil
}

func thing2() error {
	if err := thing1(); err != nil {
		return errors.Wrap(err, "failed to thing1")
	}
	return nil
}

func thing3() error {
	if err := thing2(); err != nil {
		return errors.Wrap(err, "failed to thing2")
	}
	return nil
}

func main() {
	apiKey := os.Getenv("BUGSNAG_API_KEY")
	if apiKey == "" {
		log.Fatal("Env var BUGSNAG_API_KEY required")
	}
	err := thing3()
	if err != nil {
		if err = bugsnug.Notify(err, apiKey); err != nil {
			log.Print("Failed to notify bugsnag: " + err.Error())
		}
	}
}
