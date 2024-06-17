package health

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/mattmattox/k8s-deletion-inspector/pkg/logging"
	"github.com/mattmattox/k8s-deletion-inspector/pkg/version"
)

var logger = logging.SetupLogging()

// VersionInfo represents the structure of version information.
type VersionInfo struct {
	Version   string `json:"version"`
	GitCommit string `json:"gitCommit"`
	BuildTime string `json:"buildTime"`
}

var (
	processingMu sync.RWMutex
	processing   bool

	connectionMu sync.RWMutex
	connected    bool
)

// SetProcessing sets the processing status.
func SetProcessing(status bool) {
	processingMu.Lock()
	defer processingMu.Unlock()
	processing = status
}

// IsProcessing returns the current processing status.
func IsProcessing() bool {
	processingMu.RLock()
	defer processingMu.RUnlock()
	return processing
}

// SetConnected sets the connection status.
func SetConnected(status bool) {
	connectionMu.Lock()
	defer connectionMu.Unlock()
	connected = status
}

// IsConnected returns the current connection status.
func IsConnected() bool {
	connectionMu.RLock()
	defer connectionMu.RUnlock()
	return connected
}

// HealthzHandler checks the health status of the application.
func HealthzHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if IsProcessing() {
			fmt.Fprint(w, "ok")
		} else {
			http.Error(w, "not processing", http.StatusServiceUnavailable)
		}
	}
}

// ReadyzHandler checks the readiness of the application.
func ReadyzHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if IsConnected() {
			fmt.Fprint(w, "ok")
		} else {
			http.Error(w, "not connected", http.StatusServiceUnavailable)
		}
	}
}

// VersionHandler returns version information as JSON.
func VersionHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		versionInfo := VersionInfo{
			Version:   version.Version,
			GitCommit: version.GitCommit,
			BuildTime: version.BuildTime,
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(versionInfo); err != nil {
			logger.Error("Failed to encode version info to JSON", err)
			http.Error(w, "Failed to encode version info", http.StatusInternalServerError)
		}
		logger.Debug("Version info is successfully returned")
	}
}
