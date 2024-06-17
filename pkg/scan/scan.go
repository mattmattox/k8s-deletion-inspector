package scan

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/mattmattox/k8s-deletion-inspector/pkg/k8s"
	"github.com/mattmattox/k8s-deletion-inspector/pkg/logging"
	"github.com/mattmattox/k8s-deletion-inspector/pkg/metrics"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var logger = logging.SetupLogging()

// StartScan initiates a scan of the Kubernetes cluster to find resources that are stuck in a deletion state.
func StartScan(clientset *kubernetes.Clientset, restConfig *rest.Config) (bool, int, int, error) {
	start := time.Now()  // Start time for the scan
	var totalObjects int // Counter for total objects scanned

	logger.Infoln("Starting scan...")

	logger.Debugln("Verifying access to cluster")
	if err := k8s.VerifyAccessToCluster(clientset); err != nil {
		logger.Fatalf("Error verifying access to cluster: %v", err)
		return false, 0, 0, fmt.Errorf("error verifying access to cluster: %v", err)
	}

	logger.Infoln("Fetching core namespaced resources...")
	coreResources, err := GetCoreResources(clientset)
	if err != nil {
		logger.Fatalf("Error fetching core resources: %v", err)
		return false, 0, 0, err
	}

	logger.Infof("Found %d core namespaced resources: %v", len(coreResources), coreResources)

	logger.Infoln("Fetching custom namespaced resources...")
	namespacedResources, err := k8s.GetNamespacedObjects(clientset)
	if err != nil {
		logger.Fatalf("Error fetching namespaced resources: %v", err)
		return false, 0, 0, err
	}

	logger.Infof("Found %d namespaced custom resources: %v", len(namespacedResources), namespacedResources)

	logger.Infoln("Fetching namespaces...")
	namespaces, err := k8s.GetNamespaces(clientset)
	if err != nil {
		logger.Errorf("Error fetching namespaces: %v", err)
		return false, 0, 0, err
	}

	logger.Infof("Found %d namespaces", len(namespaces))

	// Update the number of namespaces metric
	metrics.WriteNamespaceCount(len(namespaces))

	for _, ns := range namespaces {
		logger.Debugf("Processing core resources in namespace %s", ns)
		coreObjects, err := processNamespace(restConfig, ns, coreResources)
		if err != nil {
			logger.Errorf("Error processing core resources in namespace %s: %v", ns, err)
			continue
		}
		totalObjects += coreObjects

		logger.Debugf("Processing custom resources in namespace %s", ns)
		customObjects, err := processNamespace(restConfig, ns, namespacedResources)
		if err != nil {
			logger.Errorf("Error processing custom resources in namespace %s: %v", ns, err)
			continue
		}
		totalObjects += customObjects
	}

	// Record the scan metrics
	metrics.RecordScanMetrics(start, len(namespaces), totalObjects)

	return true, len(namespaces), totalObjects, nil
}

// GetCoreResources fetches the core namespaced resources available in the cluster.
func GetCoreResources(clientset *kubernetes.Clientset) ([]schema.GroupVersionResource, error) {
	discoveryClient := clientset.Discovery()
	resourceList, err := discoveryClient.ServerPreferredResources()
	if err != nil {
		logger.Errorf("Error fetching server resources: %v", err)
		return nil, fmt.Errorf("error fetching server resources: %v", err)
	}

	var coreResources []schema.GroupVersionResource
	for _, resourceGroup := range resourceList {
		if shouldIgnoreGroup(resourceGroup.GroupVersion) {
			logger.Debugf("Ignoring group version: %s", resourceGroup.GroupVersion)
			continue
		}
		for _, resource := range resourceGroup.APIResources {
			if resource.Namespaced && (resourceGroup.GroupVersion == "v1" || resourceGroup.GroupVersion == "core") {
				gv, err := schema.ParseGroupVersion(resourceGroup.GroupVersion)
				if err != nil {
					logger.Errorf("Error parsing group version: %v", err)
					return nil, fmt.Errorf("error parsing group version: %v", err)
				}
				coreResources = append(coreResources, gv.WithResource(resource.Name))
				logger.Debugf("Added core resource: %s", gv.WithResource(resource.Name))
			}
		}
	}

	return coreResources, nil
}

// processNamespace processes all resources in a given namespace.
func processNamespace(restConfig *rest.Config, ns string, resources []schema.GroupVersionResource) (int, error) {
	logger.Infof("Processing namespace %s", ns)

	totalObjects := 0

	for _, resource := range resources {
		logger.Debugf("Processing resource %s in namespace %s", resource.Resource, ns)
		objects, err := processResource(restConfig, ns, resource)
		if err != nil {
			logger.Errorf("Error processing resource %s in namespace %s: %v", resource.Resource, ns, err)
			continue
		}
		totalObjects += objects
	}

	return totalObjects, nil
}

// processResource processes all objects of a given resource type in a namespace.
func processResource(restConfig *rest.Config, ns string, resource schema.GroupVersionResource) (int, error) {
	logger.Infof("Processing resource %s", resource.Resource)

	objects, err := k8s.GetNamespaceObjects(restConfig, ns, resource)
	if err != nil {
		if isResourceNotFoundError(err) {
			logger.Warnf("Resource %s not found in namespace %s", resource.Resource, ns)
			return 0, nil
		}
		logger.Errorf("Error fetching objects for resource %s in namespace %s: %v", resource.Resource, ns, err)
		return 0, err
	}

	logger.Infof("Found %d objects for resource %s in namespace %s", len(objects), resource.Resource, ns)
	for _, object := range objects {
		logger.Debugf("Processing object %s of resource %s in namespace %s", object, resource.Resource, ns)
		processObject(restConfig, ns, resource, object)
	}

	return len(objects), nil
}

// processObject processes a single object, checking if it is deleted and recording it if it is stuck.
func processObject(restConfig *rest.Config, ns string, resource schema.GroupVersionResource, object string) {
	logger.Infof("Processing object %s", object)
	isDeleted, deletionTimestamp, err := k8s.IsObjectDeleted(restConfig, ns, resource, object)
	if err != nil {
		logger.Errorf("Error checking if object %s is deleted: %v", object, err)
		return
	}

	if isDeleted {
		logger.Infof("Object %s is deleted", object)
		metrics.AddStuckObject(ns, resource, object, deletionTimestamp)
	} else {
		logger.Infof("Object %s is not deleted", object)
	}
}

// isResourceNotFoundError checks if the error returned is a "resource not found" error.
func isResourceNotFoundError(err error) bool {
	statusErr, ok := err.(*errors.StatusError)
	return ok && statusErr.ErrStatus.Code == http.StatusNotFound
}

// shouldIgnoreGroup determines if a group should be ignored during discovery.
func shouldIgnoreGroup(groupVersion string) bool {
	return strings.Contains(groupVersion, "metrics.k8s.io")
}
