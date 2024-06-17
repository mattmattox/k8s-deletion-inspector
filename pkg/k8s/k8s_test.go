package k8s_test

import (
	"os"
	"testing"

	"github.com/mattmattox/k8s-deletion-inspector/pkg/k8s"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	kubernetesfake "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"
)

func TestConnectToCluster(t *testing.T) {
	os.Setenv("KUBERNETES_SERVICE_HOST", "dummy-host")
	os.Setenv("KUBERNETES_SERVICE_PORT", "dummy-port")
	_, _, err := k8s.ConnectToCluster("")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestVerifyAccessToCluster(t *testing.T) {
	clientset := kubernetesfake.NewSimpleClientset()
	err := k8s.VerifyAccessToCluster(clientset)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestGetNamespaces(t *testing.T) {
	clientset := kubernetesfake.NewSimpleClientset(&v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{Name: "default"},
	})
	namespaces, err := k8s.GetNamespaces(clientset)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if len(namespaces) != 1 {
		t.Errorf("Expected 1 namespace, got %d", len(namespaces))
	}
	if namespaces[0] != "default" {
		t.Errorf("Expected namespace 'default', got %s", namespaces[0])
	}
}

func TestGetNamespacedObjects(t *testing.T) {
	clientset := kubernetesfake.NewSimpleClientset()
	_, err := k8s.GetNamespacedObjects(clientset)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestGetNamespaceObjects(t *testing.T) {
	restConfig := &rest.Config{}
	ns := "default"
	resource := schema.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"}

	// Mock dynamic client
	scheme := runtime.NewScheme()
	dynamicClient := dynamicfake.NewSimpleDynamicClient(scheme)
	_ = dynamicClient // Avoid unused variable error
	_, err := k8s.GetNamespaceObjects(restConfig, ns, resource)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestGetAPIVersionForResource(t *testing.T) {
	clientset := kubernetesfake.NewSimpleClientset()
	resource := schema.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"}

	_, err := k8s.GetAPIVersionForResource(clientset, resource)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestIsObjectDeleted(t *testing.T) {
	restConfig := &rest.Config{}
	ns := "default"
	resource := schema.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"}
	name := "test-pod"

	isDeleted, _, err := k8s.IsObjectDeleted(restConfig, ns, resource, name)
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
	if isDeleted {
		t.Errorf("Expected object to not be deleted, got deleted")
	}
}

func TestShouldIgnoreGroup(t *testing.T) {
	groupVersion := "metrics.k8s.io/v1beta1"
	shouldIgnore := k8s.ShouldIgnoreGroup(groupVersion)
	if !shouldIgnore {
		t.Errorf("Expected group version to be ignored, got not ignored")
	}
}
