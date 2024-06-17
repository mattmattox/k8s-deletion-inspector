package config

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/mattmattox/k8s-deletion-inspector/pkg/version"
)

// AppConfig structure for environment-based configurations.
type AppConfig struct {
	Debug        bool   `json:"debug"`
	MetricsPort  int    `json:"metricsPort"`
	Kubeconfig   string `json:"kubeconfig"`
	DeleteAfter  int    `json:"deleteAfter"`
	ScanInterval int    `json:"scanInterval"`
	Version      bool   `json:"version"`
}

// CFG is the global configuration instance populated by LoadConfiguration.
var CFG AppConfig

// LoadConfiguration loads the configuration from the environment variables and command line flags.
func LoadConfiguration() {
	debug := flag.Bool("debug", parseEnvBool("DEBUG", true), "Enable debug mode")
	metricsPort := flag.Int("metricsPort", parseEnvInt("METRICS_PORT", 9000), "Port for metrics server")
	Kubeconfig := flag.String("kubeconfig", getEnvOrDefault("KUBECONFIG", ""), "Path to the kubeconfig file")
	DeleteAfter := flag.Int("deleteAfter", parseEnvInt("DELETE_AFTER", 72), "Number of hours to wait before deleting stuck objects")
	ScanInterval := flag.Int("scanInterval", parseEnvInt("SCAN_INTERVAL", 24), "Number of hours to wait between scans")
	showVersion := flag.Bool("version", false, "Show version and exit")

	flag.Parse()

	CFG.Debug = *debug
	CFG.MetricsPort = *metricsPort
	CFG.Kubeconfig = *Kubeconfig
	CFG.DeleteAfter = *DeleteAfter
	CFG.ScanInterval = *ScanInterval
	CFG.Version = *showVersion

	if CFG.Version {
		fmt.Printf("Version: %s\nGit Commit: %s\nBuild Time: %s\n", version.Version, version.GitCommit, version.BuildTime)
		os.Exit(0)
	}
}

// getEnvOrDefault returns the value of the environment variable with the given key or the default value if the key is not set.
func getEnvOrDefault(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// parseEnvInt parses the environment variable with the given key and returns its integer representation or the default value if the key is not set.
func parseEnvInt(key string, defaultValue int) int {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}
	intValue, err := strconv.Atoi(value)
	if err != nil {
		log.Printf("Error parsing %s as int: %v. Using default value: %d", key, err, defaultValue)
		return defaultValue
	}
	return intValue
}

// parseEnvBool parses the environment variable with the given key and returns its boolean representation or the default value if the key is not set.
func parseEnvBool(key string, defaultValue bool) bool {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}
	boolValue, err := strconv.ParseBool(value)
	if err != nil {
		log.Printf("Error parsing %s as bool: %v. Using default value: %t", key, err, defaultValue)
		return defaultValue
	}
	return boolValue
}
