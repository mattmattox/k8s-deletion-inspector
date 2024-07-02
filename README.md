# k8s-deletion-inspector

[![Pipeline](https://github.com/mattmattox/k8s-deletion-inspector/actions/workflows/pipeline.yml/badge.svg)](https://github.com/mattmattox/k8s-deletion-inspector/actions/workflows/pipeline.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/mattmattox/k8s-deletion-inspector)](https://goreportcard.com/report/github.com/mattmattox/k8s-deletion-inspector)
[![License](https://img.shields.io/github/license/mattmattox/k8s-deletion-inspector)](https://github.com/mattmattox/k8s-deletion-inspector/blob/main/LICENSE)
[![Release](https://img.shields.io/github/v/release/mattmattox/k8s-deletion-inspector)](https://github.com/mattmattox/k8s-deletion-inspector/releases)
[![Docker Pulls](https://img.shields.io/docker/pulls/cube8021/k8s-deletion-inspector)](https://hub.docker.com/r/cube8021/k8s-deletion-inspector)
[![Docker Stars](https://img.shields.io/docker/stars/cube8021/k8s-deletion-inspector)](https://hub.docker.com/r/cube8021/k8s-deletion-inspector)
[![Maintainability](https://api.codeclimate.com/v1/badges/c40c58eb73fc8e456686/maintainability)](https://codeclimate.com/github/mattmattox/k8s-deletion-inspector/maintainability)
[![Test Coverage](https://api.codeclimate.com/v1/badges/c40c58eb73fc8e456686/test_coverage)](https://codeclimate.com/github/mattmattox/k8s-deletion-inspector/test_coverage)

This project is a Kubernetes deletion inspector that scans the cluster for resources stuck in a deletion state. It uses Prometheus metrics for monitoring and exposes a metrics endpoint for monitoring tool integration.

## Components

- **pkg/config**: Contains configuration loading functionality.
- **pkg/health**: Handles health and readiness checks for the application.
- **pkg/k8s**: Interacts with the Kubernetes cluster to fetch resources and perform actions.
- **pkg/logging**: Provides logging setup for the application using Logrus.
- **pkg/metrics**: Handles Prometheus metrics setup and exposure.
- **pkg/scan**: Initiates the scan of the Kubernetes cluster to find stuck resources.
- **pkg/version**: Contains version information of the application.
- **main.go**: Main entry point of the application, sets up necessary components and starts the scan loop.

## Setup

1. **Configuration**: Set the required environment variables (e.g., `DEBUG`, `METRICS_PORT`, `KUBECONFIG`) to configure the application.
2. **Run**: Start the application to begin scanning the Kubernetes cluster for stuck resources.

## Usage

- The `StartScan` function is responsible for initiating the scan of the cluster to find stuck resources.
- The `GetStuckObjectsHandler` handles requests for stuck objects in the cluster.
- The `ForceDeleteOldResource` forcibly deletes resources stuck in a deletion state for a specified duration.

## How to Run

1. Ensure you have a Kubernetes cluster configured and accessible.
2. Set the required environment variables.
3. Run the application using `go run main.go`.

## Helm Installation

To deploy `k8s-deletion-inspector` using Helm, follow these steps:

1. **Add Helm Repository**:

    ```bash
    helm repo add supporttools https://charts.support.tools
    ```

2. **Update Helm Repositories**:

    ```bash
    helm repo update
    ```

3. **Install the Chart**:

    ```bash
    helm install k8s-deletion-inspector supporttools/k8s-deletion-inspector -f values.yaml
    ```

### Example `values.yaml`

```yaml
# Default values for k8s-deletion-inspector.
# This is a YAML-formatted file.
# Declare variables to be passed into your templates.

settings:
  debug: false
  metrics:
    port: 9000
  deleteAfter: 72 ## Number of hours to wait before force deleting the resource
  scanInterval: 24 ## Number of hours to wait before scanning for resources to delete

replicaCount: 1

image:
  repository: cube8021/k8s-deletion-inspector
  pullPolicy: IfNotPresent
  # Overrides the image tag whose default is the chart appVersion.
  tag: ${IMAGE_TAG}

imagePullSecrets: []
nameOverride: ""
fullnameOverride: k8s-deletion-inspector

serviceAccount:
  # Specifies whether a service account should be created
  create: true
  # Automatically mount a ServiceAccount's API credentials?
  automount: true
  # Annotations to add to the service account
  annotations: {}
  # The name of the service account to use.
  # If not set and create is true, a name is generated using the fullname template
  name: "k8s-deletion-inspector"

podAnnotations: {}
podLabels: {}

podSecurityContext:
  {}
  # fsGroup: 2000

securityContext:
  {}
  # capabilities:
  #   drop:
  #   - ALL
  # readOnlyRootFilesystem: true
  # runAsNonRoot: true
  # runAsUser: 1000

service:
  type: ClusterIP
  port: 9000

resources:
  {}
  # We usually recommend not to specify default resources and to leave this as a conscious
  # choice for the user. This also increases chances charts run on environments with little
  # resources, such as Minikube. If you do want to specify resources, uncomment the following
  # lines, adjust them as necessary, and remove the curly braces after 'resources:'.
  # limits:
  #   cpu: 100m
  #   memory: 128Mi
  # requests:
  #   cpu: 100m
  #   memory: 128Mi

# Additional volumes on the output Deployment definition.
volumes: []
# - name: foo
#   secret:
#     secretName: mysecret
#     optional: false

# Additional volumeMounts on the output Deployment definition.
volumeMounts: []
# - name: foo
#   mountPath: "/etc/foo"
#   readOnly: true

nodeSelector: {}

tolerations: []

affinity: {}
```

## Author

This project was developed by [Matt Mattox](https://github.com/mattmattox).
