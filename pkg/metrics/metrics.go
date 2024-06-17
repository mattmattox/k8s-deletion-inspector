package metrics

import (
	"encoding/json"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/mattmattox/k8s-deletion-inspector/pkg/config"
	"github.com/mattmattox/k8s-deletion-inspector/pkg/health"
	"github.com/mattmattox/k8s-deletion-inspector/pkg/logging"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var logger = logging.SetupLogging()

var (
	stuckObjects      []StuckObject
	stuckObjectsMutex sync.Mutex
)

var mu sync.Mutex

// Prometheus metrics
var (
	namespaceCount = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "k8s_deletion_inspector_namespace_count",
		Help: "Number of namespaces",
	})

	scanDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "k8s_deletion_inspector_scan_duration_seconds",
		Help:    "Duration of the scan in seconds",
		Buckets: prometheus.DefBuckets,
	})

	totalObjectsScanned = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "k8s_deletion_inspector_total_objects_scanned",
		Help: "Total number of objects scanned",
	})

	numberStuckObjects = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "k8s_deletion_inspector_stuck_resources_total",
		Help: "Number of stuck objects",
	})
)

// StuckObject represents a stuck object in the Kubernetes cluster
type StuckObject struct {
	Namespace            string                      `json:"namespace"`
	Resource             string                      `json:"resource"`
	Name                 string                      `json:"name"`
	DeleteTimestamp      time.Time                   `json:"deleteTimestamp"`
	GroupVersionResource schema.GroupVersionResource `json:"groupVersionResource"`
}

// Set up Prometheus metrics
func init() {
	logger.Debug("Initializing Prometheus metrics")
	prometheus.MustRegister(namespaceCount, scanDuration, totalObjectsScanned, numberStuckObjects)
}

// GetStuckObjectsHandler handles requests for stuck objects in the cluster
func GetStuckObjectsHandler(w http.ResponseWriter, r *http.Request) {
	logger.Debug("Handling request for stuck objects")
	mu.Lock()
	defer mu.Unlock()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(stuckObjects); err != nil {
		logger.Errorf("Failed to encode stuck objects: %v", err)
		http.Error(w, "Failed to encode stuck objects", http.StatusInternalServerError)
		return
	}
	logger.Debug("Successfully encoded stuck objects")
}

// AddStuckObject adds a stuck object to the list of stuck objects in memory and Prometheus metrics
func AddStuckObject(namespace string, gvr schema.GroupVersionResource, object string, deletionTimestamp time.Time) {
	logger.Debugf("Adding stuck object: namespace=%s, resource=%s, object=%s, deletionTimestamp=%s", namespace, gvr.Resource, object, deletionTimestamp)
	stuckObjectsMutex.Lock()
	defer stuckObjectsMutex.Unlock()

	stuckObject := StuckObject{
		Namespace:            namespace,
		Resource:             gvr.Resource,
		Name:                 object,
		DeleteTimestamp:      deletionTimestamp,
		GroupVersionResource: gvr,
	}

	stuckObjects = append(stuckObjects, stuckObject)
	numberStuckObjects.Set(float64(len(stuckObjects)))
	logger.Debugf("Stuck object added: %+v", stuckObject)
}

// GetStuckObjects returns the list of stuck objects in the cluster
func GetStuckObjects() []StuckObject {
	logger.Debug("Fetching stuck objects")
	stuckObjectsMutex.Lock()
	defer stuckObjectsMutex.Unlock()

	logger.Debugf("Returning %d stuck objects", len(stuckObjects))
	return stuckObjects
}

// WriteNamespaceCount sets namespace count for Prometheus metrics
func WriteNamespaceCount(count int) {
	logger.Debugf("Setting namespace count to %d", count)
	namespaceCount.Set(float64(count))
}

// RecordScanMetrics records scan metrics for Prometheus metrics
func RecordScanMetrics(start time.Time, namespaces, objects int) {
	duration := time.Since(start).Seconds()
	logger.Debugf("Recording scan metrics: duration=%.2f seconds, namespaces=%d, objects=%d", duration, namespaces, objects)
	scanDuration.Observe(duration)
	totalObjectsScanned.Add(float64(objects))
}

// StartMetricsServer starts the metrics server
func StartMetricsServer() {
	logger.Debug("Starting metrics server setup")
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/healthz", health.HealthzHandler())
	mux.HandleFunc("/readyz", health.ReadyzHandler())
	mux.HandleFunc("/version", health.VersionHandler())
	mux.HandleFunc("/stuck-objects", GetStuckObjectsHandler)

	serverPortStr := strconv.Itoa(config.CFG.MetricsPort)
	logger.Printf("Metrics server starting on port %s\n", serverPortStr)

	srv := &http.Server{
		Addr:         ":" + serverPortStr,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	if err := srv.ListenAndServe(); err != nil {
		logger.Fatalf("Metrics server failed to start: %v", err)
	}
}
