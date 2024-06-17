package config

import (
	"os"
	"testing"
)

func TestLoadConfiguration(t *testing.T) {
	os.Setenv("DEBUG", "true")
	os.Setenv("METRICS_PORT", "9182")
	os.Setenv("KUBECONFIG", "/path/to/kubeconfig")

	LoadConfiguration()

	if !CFG.Debug {
		t.Errorf("Expected Debug to be true, got %v", CFG.Debug)
	}
	if CFG.MetricsPort != 9182 {
		t.Errorf("Expected MetricsPort to be 9182, got %d", CFG.MetricsPort)
	}
	if CFG.Kubeconfig != "/path/to/kubeconfig" {
		t.Errorf("Expected KUBECONFIG to be '/path/to/kubeconfig', got '%s'", CFG.Kubeconfig)
	}
}

func TestGetEnvOrDefault(t *testing.T) {
	key := "TEST_KEY"
	defaultValue := "default"

	value := getEnvOrDefault(key, defaultValue)
	if value != defaultValue {
		t.Errorf("Expected '%s', got '%s'", defaultValue, value)
	}

	expectedValue := "test-value"
	os.Setenv(key, expectedValue)
	value = getEnvOrDefault(key, defaultValue)
	if value != expectedValue {
		t.Errorf("Expected '%s', got '%s'", expectedValue, value)
	}
}
