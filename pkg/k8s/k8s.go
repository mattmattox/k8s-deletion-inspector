package k8s

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/mattmattox/k8s-deletion-inspector/pkg/logging"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var logger = logging.SetupLogging()

// ClientsetInterface defines the interface for Kubernetes clientsets.
type ClientsetInterface interface {
	CoreV1() v1.CoreV1Interface
	Discovery() discovery.DiscoveryInterface
}

// ConnectToCluster connects to the Kubernetes cluster using the provided kubeconfig file.
// If the environment variables KUBERNETES_SERVICE_HOST and KUBERNETES_SERVICE_PORT are set,
// it assumes the application is running inside a Kubernetes cluster and uses the in-cluster config.
func ConnectToCluster(kubeconfig string) (*kubernetes.Clientset, *rest.Config, error) {
	logger.Debugln("Connecting to Kubernetes cluster...")

	// Check if a kubeconfig file is provided.
	logger.Debugln("Checking if a kubeconfig file is provided...")
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		logger.Errorf("Error creating client config: %v", err)
		return nil, nil, fmt.Errorf("error creating client config: %v", err)
	}
	logger.Debugf("Using kubeconfig file %s to connect to Kubernetes cluster...", kubeconfig)

	// Check if the application is running inside a Kubernetes cluster.
	logger.Debugln("Checking if the application is running inside a Kubernetes cluster...")
	if os.Getenv("KUBERNETES_SERVICE_HOST") != "" && os.Getenv("KUBERNETES_SERVICE_PORT") != "" {
		logger.Debugln("Application is running inside a Kubernetes cluster...")
		config, err = rest.InClusterConfig()
		if err != nil {
			logger.Errorf("Error creating in-cluster client config: %v", err)
			return nil, nil, fmt.Errorf("error creating in-cluster client config: %v", err)
		}
		logger.Debugln("Using in-cluster config to connect to Kubernetes cluster...")
	}

	// Otherwise, it assumes that the application is running outside a Kubernetes cluster and uses the provided kubeconfig file.
	logger.Debugln("Application is running outside a Kubernetes cluster...")
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		logger.Errorf("Error creating clientset: %v", err)
		return nil, nil, fmt.Errorf("error creating clientset: %v", err)
	}

	logger.Debugln("Successfully connected to Kubernetes cluster...")
	return clientset, config, nil
}

// VerifyAccessToCluster verifies if the application has access to the Kubernetes cluster
// by attempting to list the nodes in the cluster.
func VerifyAccessToCluster(clientset ClientsetInterface) error {
	logger.Debugln("Verifying access to Kubernetes cluster...")
	ctx := context.TODO()
	listOptions := metav1.ListOptions{}

	// Attempt to list the nodes in the cluster to verify access.
	logger.Debugln("Listing nodes in the cluster...")
	_, err := clientset.CoreV1().Nodes().List(ctx, listOptions)
	if err != nil {
		logger.Errorf("Error listing nodes: %v", err)
		return fmt.Errorf("error listing nodes: %v", err)
	}

	logger.Debugln("Successfully verified access to Kubernetes cluster...")
	return nil
}

// GetNamespaces retrieves the list of namespaces in the Kubernetes cluster.
func GetNamespaces(clientset ClientsetInterface) ([]string, error) {
	logger.Debugln("Fetching namespaces...")

	// List all namespaces in the cluster.
	logger.Debugln("Listing namespaces in the cluster...")
	namespaceList, err := clientset.CoreV1().Namespaces().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		logger.Errorf("Error fetching namespaces: %v", err)
		return nil, err
	}

	// Extract the names of the namespaces.
	logger.Debugln("Extracting names of namespaces...")
	namespaces := make([]string, len(namespaceList.Items))
	for i, namespace := range namespaceList.Items {
		logger.Debugf("Namespace: %s", namespace.GetName())
		namespaces[i] = namespace.GetName()
	}

	logger.Debugln("Successfully fetched namespaces...")
	return namespaces, nil
}

// GetNamespacedObjects retrieves the list of namespaced objects available in the cluster.
func GetNamespacedObjects(clientset ClientsetInterface) ([]schema.GroupVersionResource, error) {
	logger.Debugln("Fetching namespaced API resources...")

	// List all namespaced API resources in the cluster.
	apiResourceList, err := clientset.Discovery().ServerPreferredResources()
	if err != nil {
		logger.Errorf("Error fetching namespaced API resources: %v", err)
		return nil, err
	}
	logger.Debugf("API Resources: %v", apiResourceList)

	// Extract the namespaced objects from the API resources.
	var gv schema.GroupVersion
	objects := make([]schema.GroupVersionResource, 0)
	for _, apiResources := range apiResourceList {
		// Skip if apiResources contains metrics.k8s.io
		logger.Debugln("Checking if API resources contain metrics.k8s.io...")
		if strings.Contains(apiResources.GroupVersion, "metrics.k8s.io") {
			logger.Debugf("Ignoring group version: %s", apiResources.GroupVersion)
			continue
		}
		logger.Debugf("Found group version: %s", apiResources.GroupVersion)
		gv, _ = schema.ParseGroupVersion(apiResources.GroupVersion)
		for _, apiResource := range apiResources.APIResources {
			if apiResource.Namespaced {
				logger.Debugf("Found namespaced API resource: %s", apiResource.Name)
				object := gv.WithResource(apiResource.Name)
				objects = append(objects, object)
			}
		}
	}

	logger.Debugln("Successfully fetched namespaced API resources...")
	return objects, nil
}

// GetNamespaceObjects retrieves the objects in a namespace for a given resource.
func GetNamespaceObjects(restConfig *rest.Config, ns string, resource schema.GroupVersionResource) ([]string, error) {
	logger.Debugf("Fetching objects for resource %s in namespace %s with GroupVersion %s", resource.Resource, ns, resource.GroupVersion())

	// Create a dynamic client to interact with the Kubernetes API.
	logger.Debugln("Creating dynamic client...")
	dynamicClient, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		logger.Errorf("Error creating dynamic client: %v", err)
		return nil, err
	}

	// List all objects in the namespace for the given resource.
	logger.Debugln("Listing objects in the namespace...")
	resourceClient := dynamicClient.Resource(resource).Namespace(ns)
	objectList, err := resourceClient.List(context.Background(), metav1.ListOptions{})
	if err != nil {
		logger.Errorf("Error fetching objects for resource %s in namespace %s: %v", resource.Resource, ns, err)
		return nil, err
	}

	if len(objectList.Items) == 0 {
		logger.Warnf("No objects found for resource %s in namespace %s", resource.Resource, ns)
	} else {
		logger.Debugf("Found %d objects for resource %s in namespace %s: %v", len(objectList.Items), resource.Resource, ns, objectList.Items)
	}

	objects := make([]string, len(objectList.Items))
	for i, object := range objectList.Items {
		objects[i] = object.GetName()
	}
	logger.Debugf("Object names for resource %s in namespace %s: %v", resource.Resource, ns, objects)
	return objects, nil
}

// GetAPIVersionForResource retrieves the API version for a given resource.
func GetAPIVersionForResource(clientset ClientsetInterface, resource schema.GroupVersionResource) (string, error) {
	apiResource, err := clientset.Discovery().ServerResourcesForGroupVersion(resource.GroupVersion().String())
	if err != nil {
		return "", err
	}
	apiVersion := apiResource.APIResources[0].Version
	return apiVersion, nil
}

// IsObjectDeleted checks if an object is marked for deletion and returns the deletion timestamp if it exists.
func IsObjectDeleted(restConfig *rest.Config, ns string, resource schema.GroupVersionResource, name string) (bool, time.Time, error) {
	dynamicClient, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		return false, time.Time{}, fmt.Errorf("error creating dynamic client: %v", err)
	}

	res := dynamicClient.Resource(resource).Namespace(ns)
	obj, err := res.Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		logger.Errorf("Error fetching object %s in namespace %s: %v", name, ns, err)
		return false, time.Time{}, err
	}

	deletionTimestamp := obj.GetDeletionTimestamp()
	if deletionTimestamp != nil {
		logger.Infof("Object %s in namespace %s is marked for deletion", name, ns)
		return true, deletionTimestamp.Time, nil
	}

	logger.Infof("Object %s in namespace %s is not marked for deletion", name, ns)
	return false, time.Time{}, nil
}

// ShouldIgnoreGroup determines if a group should be ignored during discovery.
func ShouldIgnoreGroup(groupVersion string) bool {
	return strings.Contains(groupVersion, "metrics.k8s.io")
}

// ForceDeleteOldResource forcefully deletes a specific resource that has been in the deletion state for more than DeleteAfter hours.
func ForceDeleteOldResource(restConfig *rest.Config, ns string, resource schema.GroupVersionResource, name string) error {
	logger.Infof("Force deleting old resource %s for resource %s in namespace %s", name, resource.Resource, ns)

	dynamicClient, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		return fmt.Errorf("error creating dynamic client: %v", err)
	}

	resourceClient := dynamicClient.Resource(resource).Namespace(ns)
	obj, err := resourceClient.Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("error fetching object %s in namespace %s: %v", name, ns, err)
	}

	obj.SetFinalizers(nil)
	_, err = resourceClient.Update(context.Background(), obj, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("error removing finalizers for object %s in namespace %s: %v", name, ns, err)
	}

	err = resourceClient.Delete(context.Background(), name, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("error deleting object %s in namespace %s: %v", name, ns, err)
	}

	logger.Infof("Successfully removed finalizers and deleted object %s in namespace %s", name, ns)
	return nil
}
