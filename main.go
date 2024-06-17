package main

import (
	"time"

	"github.com/mattmattox/k8s-deletion-inspector/pkg/config"
	"github.com/mattmattox/k8s-deletion-inspector/pkg/k8s"
	"github.com/mattmattox/k8s-deletion-inspector/pkg/logging"
	"github.com/mattmattox/k8s-deletion-inspector/pkg/metrics"
	"github.com/mattmattox/k8s-deletion-inspector/pkg/scan"
)

var logger = logging.SetupLogging()

func main() {
	config.LoadConfiguration()

	logger.Infoln("Starting k8s-deletion-inspector")

	clientset, restConfig, err := k8s.ConnectToCluster(config.CFG.Kubeconfig)
	if err != nil {
		logger.Fatalf("Error connecting to cluster: %v", err)
	}

	go func() {
		for {
			success, namespaces, totalObjects, err := scan.StartScan(clientset, restConfig)
			if err != nil {
				logger.Fatalf("Error starting scan: %v", err)
			}

			if success {
				logger.Infof("Scan completed successfully: %d namespaces, %d objects", namespaces, totalObjects)
			} else {
				logger.Infoln("Scan did not complete successfully")
			}

			// Perform cleanup of old resources
			stuckObjects := metrics.GetStuckObjects()
			for _, obj := range stuckObjects {
				if time.Since(obj.DeleteTimestamp) > time.Duration(config.CFG.DeleteAfter)*time.Hour {
					err := k8s.ForceDeleteOldResource(restConfig, obj.Namespace, obj.GroupVersionResource, obj.Name)
					if err != nil {
						logger.Errorf("Error force deleting old resource %s in namespace %s: %v", obj.Name, obj.Namespace, err)
					} else {
						logger.Infof("Successfully force deleted old resource %s in namespace %s", obj.Name, obj.Namespace)
					}
				}
			}

			// Sleep between scans
			time.Sleep(time.Duration(config.CFG.ScanInterval) * time.Hour)
		}
	}()

	metrics.StartMetricsServer()
}
