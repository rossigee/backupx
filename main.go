package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/rossigee/backupx/config"
	"github.com/rossigee/backupx/destinations/s3"
	"github.com/rossigee/backupx/internal/jsonconfig"
	"github.com/rossigee/backupx/internal/pipeline"
	"github.com/rossigee/backupx/internal/yamlconfig"
	"github.com/rossigee/backupx/notifications"
	"github.com/rossigee/backupx/notifications/backupregistry"
	"github.com/rossigee/backupx/notifications/flagfile"
	"github.com/rossigee/backupx/notifications/logging"
	"github.com/rossigee/backupx/notifications/slack"
	"github.com/rossigee/backupx/sources/folder"
	"github.com/rossigee/backupx/sources/mysql"
	"github.com/rossigee/backupx/sources/postgresql"
)

var (
	conf       config.IBackupConfig
	verbose    bool
	debug      bool
)

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [-v] [-d] <filename>\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "<filename> is YAML or JSON config file (see docs)\n")
	flag.PrintDefaults()
}

func parseArgs() (string, bool) {
	flag.Usage = usage
	flag.BoolVar(&verbose, "v", false, "Verbose output")
	flag.BoolVar(&debug, "d", false, "Debug output")
	flag.Parse()

	args := flag.Args()
	if len(args) != 1 {
		return "", true
	}

	return args[0], false
}

func newDestination(config interface {
	GetType() string
	GetOtherAttributes() map[string]string
}) (interface{}, error) {
	switch config.GetType() {
	case "s3":
		hostname, _ := os.Hostname()
		return initS3Destination(config, hostname)
	default:
		return nil, fmt.Errorf("unknown destination type: %s", config.GetType())
	}
}

func initS3Destination(config interface {
	GetType() string
	GetOtherAttributes() map[string]string
}, hostname string) (interface{}, error) {
	s3config, ok := config.(interface {
		GetId() string
		GetType() string
		GetName() string
		GetOtherAttributes() map[string]string
	})
	if !ok {
		return nil, fmt.Errorf("invalid S3 destination config")
	}
	return createS3Destination(s3config, hostname)
}

func createS3Destination(cfg interface {
	GetId() string
	GetType() string
	GetName() string
	GetOtherAttributes() map[string]string
}, hostname string) (interface{}, error) {
	return s3.NewS3Destination(cfg, hostname)
}

func newSource(config interface {
	GetId() string
	GetType() string
	GetName() string
	GetOtherAttributes() map[string]string
}) (interface{}, error) {
	switch config.GetType() {
	case "postgresql", "pgsql":
		return initPostgreSQLSource(config)
	case "mysql", "mariadb":
		return initMySQLSource(config)
	case "folder":
		return initFolderSource(config)
	default:
		return nil, fmt.Errorf("unknown source type: %s", config.GetType())
	}
}

func initFolderSource(cfg interface {
	GetId() string
	GetType() string
	GetName() string
	GetOtherAttributes() map[string]string
}) (interface{}, error) {
	return folder.NewFolderSource(cfg)
}

func initPostgreSQLSource(cfg interface {
	GetId() string
	GetType() string
	GetName() string
	GetOtherAttributes() map[string]string
}) (interface{}, error) {
	return postgresql.NewPostgreSQLSource(cfg)
}

func initMySQLSource(cfg interface {
	GetId() string
	GetType() string
	GetName() string
	GetOtherAttributes() map[string]string
}) (interface{}, error) {
	return mysql.NewMySQLSource(cfg)
}

func newNotification(config interface {
	GetId() string
	GetType() string
	GetName() string
	GetOtherAttributes() map[string]string
}) (interface{}, error) {
	switch config.GetType() {
	case "flagfile":
		return initFlagFileNotification(config)
	case "logging", "stdout":
		return initLoggingNotification(config)
	case "slack":
		return initSlackNotification(config)
	case "backup-registry":
		return initBackupRegistryNotification(config)
	default:
		return nil, fmt.Errorf("unknown notification type: %s", config.GetType())
	}
}

func initFlagFileNotification(cfg interface {
	GetId() string
	GetType() string
	GetName() string
	GetOtherAttributes() map[string]string
}) (interface{}, error) {
	return flagfile.NewFlagFileNotification(cfg)
}

func initLoggingNotification(cfg interface {
	GetId() string
	GetType() string
	GetName() string
	GetOtherAttributes() map[string]string
}) (interface{}, error) {
	return logging.NewLoggingNotification(cfg)
}

func initSlackNotification(cfg interface {
	GetId() string
	GetType() string
	GetName() string
	GetOtherAttributes() map[string]string
}) (interface{}, error) {
	return slack.NewSlackNotification(cfg)
}

func initBackupRegistryNotification(cfg interface {
	GetId() string
	GetType() string
	GetName() string
	GetOtherAttributes() map[string]string
}) (interface{}, error) {
	return backupregistry.NewBackupRegistry(cfg)
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
	destMap := make(map[string]interface{})
	for _, destCfg := range conf.GetDestinationConfigs() {
		log.Printf("Initialising destination '%s'", destCfg.GetId())
		dest, err := newDestination(destCfg)
		if err != nil {
			log.Fatalf("Failed to initialize destination '%s': %v", destCfg.GetId(), err)
		}
		destMap[destCfg.GetId()] = dest
		log.Printf("Successfully initialised destination '%s' (type: %s)", destCfg.GetId(), destCfg.GetType())
	}

	// Initialise notifications
	notifMap := make(map[string]interface{})
	for _, notifCfg := range conf.GetNotificationConfigs() {
		log.Printf("Initialising notification '%s'", notifCfg.GetId())
		notif, err := newNotification(notifCfg)
		if err != nil {
			log.Fatalf("Failed to initialize notification '%s': %v", notifCfg.GetId(), err)
		}
		notifMap[notifCfg.GetId()] = notif
		log.Printf("Successfully initialised notification '%s' (type: %s)", notifCfg.GetId(), notifCfg.GetType())
	}

	// Loop for each source
	sources := conf.GetSourceConfigs()
	for _, srcCfg := range sources {
		log.Printf("Processing source '%s' (type: %s)", srcCfg.GetId(), srcCfg.GetType())

		// Initialize source
		src, err := newSource(srcCfg)
		if err != nil {
			log.Printf("Failed to initialize source '%s': %v", srcCfg.GetId(), err)
			continue
		}
		log.Printf("Successfully initialised source '%s'", srcCfg.GetId())

		// Get source reader
		var sourceReader interface {
			GetReader() (interface{}, error)
		}
		switch s := src.(type) {
		case interface {
			GetReader() (interface{}, error)
		}:
			sourceReader = s
		default:
			log.Printf("Source '%s' does not implement GetReader", srcCfg.GetId())
			continue
		}

		reader, err := sourceReader.GetReader()
		if err != nil {
			log.Printf("Failed to get reader from source '%s': %v", srcCfg.GetId(), err)
			continue
		}

		readCloser, ok := reader.(interface {
			Close() error
			Read([]byte) (int, error)
		})
		if !ok {
			log.Printf("Source reader is not an io.ReadCloser")
			continue
		}
		defer func() {
			_ = readCloser.Close()
		}()

		// Collect destinations for this source
		var destWriters []pipeline.DestinationWriter
		for _, destObj := range destMap {
			s3Dest, ok := destObj.(*s3.S3Destination)
			if ok {
				destWriters = append(destWriters, s3Dest)
			}
		}

		if len(destWriters) == 0 {
			log.Printf("No suitable destinations found for source '%s'", srcCfg.GetId())
			continue
		}

		// Execute backup pipeline
		pipelineConfig := pipeline.PipelineConfig{
			SourceId:     srcCfg.GetId(),
			SourceName:   srcCfg.GetName(),
			Passphrase:   srcCfg.GetOtherAttributes()["passphrase"],
			CompressOnly: srcCfg.GetOtherAttributes()["compress_only"] == "1",
			Destinations: destWriters,
		}

		bp := pipeline.NewBackupPipeline(pipelineConfig)
		backupResult := bp.Execute(context.Background(), readCloser, srcCfg.GetId())

		log.Printf("Backup result: success=%v, duration=%dms", backupResult.Success, backupResult.Duration)

		// Send notifications
		ctx := context.Background()
		for notifID, notifObj := range notifMap {
			switch n := notifObj.(type) {
			case interface {
				Notify(context.Context, notifications.BackupResult) error
			}:
				if err := n.Notify(ctx, backupResult); err != nil {
					log.Printf("Notification '%s' failed: %v", notifID, err)
				} else {
					log.Printf("Notification '%s' sent successfully", notifID)
				}
			}
		}
	}

	// Happy ending
	log.Printf("Completed successfully.")
}
