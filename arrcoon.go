package main

import (
	"arrcoon/arrs"
	"arrcoon/clients"
	"io"
	"os"
	"path/filepath"

	"github.com/natefinch/lumberjack"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Sonarr struct {
		Host  string `yaml:"host"`
		Token string `yaml:"token"`
	} `yaml:"sonarr"`
	Radarr struct {
		Host  string `yaml:"host"`
		Token string `yaml:"token"`
	} `yaml:"radarr"`
	Clients map[string]clients.ClientConfig `yaml:"clients"`
	Log     struct {
		Level string `yaml:"level"`
	} `yaml:"log"`
}

func main() {
	binDir := getArcoonDir()
	initLog(binDir)

	configFile, err := os.Open(filepath.Join(binDir, "config.yml"))
	if err != nil {
		log.WithError(err).Error("Couldn't read config")
		return
	}
	defer configFile.Close()

	// Create a decoder and a Config instance
	var config Config
	decoder := yaml.NewDecoder(configFile)

	// Decode the YAML file into the struct
	err = decoder.Decode(&config)
	if err != nil {
		log.WithError(err).Error("Couldn't decode YAML config")
		return
	}

	// Set log level from config or fallback to info
	level, err := log.ParseLevel(config.Log.Level)
	if err != nil {
		level = log.InfoLevel
	}
	log.SetLevel(level)

	if len(config.Clients) == 0 {
		log.Error("No torrent clients defined")
		os.Exit(1)
	}
	if len(config.Clients) > 1 {
		log.WithFields(log.Fields{
			"Clients": config.Clients,
		}).Error("Multiple torrent clients not supported")
		os.Exit(1)
	}

	var clientType string
	var clientConfig clients.ClientConfig
	for k, v := range config.Clients {
		clientType = k
		clientConfig = v
		break
	}

	// Get Torrent Client Instance
	constructor, clientExists := clients.Instances[clientType]

	if !clientExists {
		log.WithFields(log.Fields{
			"Type": clientType,
		}).Error("Couldn't map torrent client")
	}

	torrentClient := constructor(clientConfig)

	// Get Sonarr event type
	sonarrEventType := os.Getenv("sonarr_eventtype")
	radarrEventType := os.Getenv("radarr_eventtype")
	arrcoonEventType := os.Getenv("arrcoon_eventtype")

	if arrcoonEventType != "" {
		switch arrcoonEventType {
		case "test_torrent_client":
			torrentClient.Test()
		default:
			log.WithFields(log.Fields{
				"EventType": arrcoonEventType,
			}).Info("Unknown arrcoon event type")
		}
		return
	}
	if sonarrEventType == "Test" || radarrEventType == "Test" {
		log.WithFields(log.Fields{
			"Sonarr URL":     config.Sonarr.Host,
			"Radarr URL":     config.Radarr.Host,
			"Torrent Client": clientType,
		}).Info()
		if !torrentClient.Test() {
			os.Exit(1)
		}
	}

	switch {
	case sonarrEventType != "":
		log.WithFields(log.Fields{
			"Sonarr EventType": sonarrEventType,
		}).Debug()
		sonarr := arrs.NewSonarr(binDir, config.Sonarr.Host, config.Sonarr.Token, torrentClient)
		sonarr.HandleEvent(sonarrEventType)
	case radarrEventType != "":
		log.WithFields(log.Fields{
			"Radarr EventType": radarrEventType,
		}).Debug()
		radarr := arrs.NewRadarr(binDir, config.Radarr.Host, config.Radarr.Token, torrentClient)
		radarr.HandleEvent(radarrEventType)
	default:
		log.Warn("Neither Sonarr nor Radarr events found")
		os.Exit(1)
	}
}

func getArcoonDir() string {
	binPath, err := os.Executable()
	if err != nil {
		log.WithError(err).Error("Couldn't get arrcoon binary path")
	}
	binDir := filepath.Dir(binPath)
	log.SetOutput(io.MultiWriter(os.Stdout))
	log.WithFields(log.Fields{"Arcoon binary directory": binDir}).Info()
	return binDir
}

func initLog(dir string) {
	logPath := filepath.Join(dir, "logs/arrcoon.log")

	// Ensure the log directory exists
	logDir := filepath.Dir(logPath)
	err := os.MkdirAll(logDir, os.ModePerm)

	if err != nil {
		log.Error("Failed to create a directory for log file")
		return
	}

	// Configure lumberjack for log rotation
	logFile := &lumberjack.Logger{
		Filename:   logPath, // Log file location
		MaxSize:    2,       // Max size in MB before rotation (e.g., 10MB)
		MaxBackups: 2,       // Maximum number of old log files to keep
		Compress:   true,    // Enable compression for rotated logs
	}
	log.SetOutput(io.MultiWriter(os.Stdout, logFile))
}
