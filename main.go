package main

import (
	"os"
	"log"
	"flag"
	"fmt"
	"time"
)

var (
	verbose bool
)

type notification struct {
	startTime time.Time
	endTime time.Time
}

func main() {
	var (
		 conf conf
		 notification notification
	)

	verbose := flag.Bool("v", false, "Verbose output")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s <filename>\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "<filename> is YAML config file (see docs)\n")
		flag.PrintDefaults()
	}
	flag.Parse()
	args := flag.Args()
	if len(args) != 1 {
		flag.Usage()
		os.Exit(1)
	}

	var filename = args[0]
	if *verbose {
		log.Printf("Reading config from file: %s", filename)
	}
	conf.parseYamlConf(filename)

	for _, dest := range conf.Spec.Destinations {
		log.Printf("Initialising destination '%s'", dest.Description)
		// TODO
	}

	log.Printf("Initialising source '%s'", conf.Spec.Source.Description)
	notification.startTime = time.Now()
  // TODO stuff
	notification.endTime = time.Now()
	elapsed := notification.endTime.Sub(notification.startTime)
	log.Printf("Process run in %d milliseconds.", elapsed)

	for _, notifier := range conf.Spec.Notifiers {
		log.Printf("Sending notification via '%s'", notifier.Description)
		// TODO
	}

	// Happy ending
	log.Printf("Completed successfully.")
}
