package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/rossigee/backupx/config"
	"github.com/rossigee/backupx/internal/jsonconfig"
	"github.com/rossigee/backupx/internal/yamlconfig"
)

var (
	conf    config.IBackupConfig
	verbose bool
	debug   bool
	timings Timings
)

type Timings struct {
	startTime time.Time
	endTime   time.Time
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [-v] [-d] <filename>\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "<filename> is YAML or JSON config file (see docs)\n")
	flag.PrintDefaults()
}

func parseArgs() (string, bool) {
	flag.Usage = usage
	verbose_opt := flag.Bool("v", false, "Verbose output")
	debug_opt := flag.Bool("d", false, "Debug output")
	flag.Parse()
	verbose = *verbose_opt
	debug = *debug_opt

	args := flag.Args()
	if len(args) != 1 {
		return "", true
	}

	return args[0], false
}

func main() {
	// Determine where to find our configuration
	filename, usage := parseArgs()
	if usage {
		flag.Usage()
		os.Exit(1)
	}

	// Parse JSON or YAML file specified
	var err error
	if strings.HasSuffix(filename, ".json") {
		conf, err = jsonconfig.ParseJSONConfigFile(filename)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing JSON config file: %s\n", err)
			os.Exit(2)
		}
	} else if strings.HasSuffix(filename, ".yaml") {
		conf, err = yamlconfig.ParseYAMLConfigFile(filename)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing YAML config file: %s\n", err)
			os.Exit(2)
		}
	} else {
		log.Fatal("Unknown config file type")
	}

	// Initialise destinations
	for _, dest := range conf.GetDestinationConfigs() {
		log.Printf("Initialising destination '%s'", dest.GetId())
		// TODO
	}

	// Loop for each source
	sources := conf.GetSourceConfigs()
	for _, src := range sources {
		log.Printf("Initialising source '%s'", src.GetId())
		timings.startTime = time.Now()

		// TODO stuff
		log.Printf("DOING STUFF...")

		timings.endTime = time.Now()
		elapsed := timings.endTime.Sub(timings.startTime)
		log.Printf("Process run in %d milliseconds.", elapsed.Milliseconds())

		for _, notification := range conf.GetNotificationConfigs() {
			log.Printf("Sending timings via '%s'", notification.GetName())
			// TODO
		}
	}

	// Happy ending
	log.Printf("Completed successfully.")
}
