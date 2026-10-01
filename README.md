# Dokimos

Dokimos project is a Kubernetes validating admission webhook written in Go. It's designed to enforce custom policies on resources created within a Kubernetes cluster. Specifically, this example implementation prevents pods from using images from Docker Hub unless they are in a designated, whitelisted namespace.

## How It Works

The webhook intercepts `CREATE` requests for Pods sent to the Kubernetes API server. Before any pod is persisted to `etcd`, the API server sends an `AdmissionReview` request to our webhook. The webhook inspects the pod's specification—specifically its namespace—and checks it against a predefined list of allowed namespaces.

- If the pod is in an allowed namespace, the webhook responds with `{"allowed": true}`, and the pod is created.
- If the pod is in a restricted namespace, the webhook responds with `{"allowed": false}` and a message explaining why, and the API server rejects the pod creation.

This provides a powerful mechanism for enforcing cluster-wide policies and security constraints.

## Getting Started

These instructions will get you a copy of the project up and running on your local machine for development and testing purposes.

### Prerequisites

- Go (version 1.24 or higher recommended)
- A running Kubernetes cluster (e.g., Minikube, Kind, Docker Desktop)
- `kubectl` configured to communicate with your cluster
- `ngrok` to expose your local webhook service to your cluster

### Installation & Setup

1.  **Clone the repository:**

    ```sh
    git clone https://github.com/lavishpal/Dokimos.git
    cd Dokimos
    ```

2.  **Run the webhook service:**
    You can run the service directly using go. The server will start on port `8080`.

    ```sh
    go run cmd/api/main.go
    ```

3.  **Expose your local service with ngrok:**
    In a new terminal, expose your local port `8080` to the internet.

    ```sh
    ngrok http 8080
    ```

    Copy the HTTPS forwarding URL provided by ngrok (e.g., `https://<unique-id>.ngrok-free.app`).

4.  **Configure the Webhook in Kubernetes:**
    Open the `kubernetes/validatingwebhookconfiguration.yaml` file and replace the placeholder URL with your ngrok HTTPS URL. Make sure to append the `/validate` path.

    ```yaml
    # kubernetes/validatingwebhookconfiguration.yaml
    // ...
      clientConfig:
        url: "https://<your-ngrok-https-url>/validate"
    // ...
    ```

5.  **Apply the configuration to your cluster:**
    ```sh
    kubectl apply -f kubernetes/validatingwebhookconfiguration.yaml
    ```
    Your webhook is now active and will intercept all pod creation events.

### Testing the Policy

To see the webhook in action, try creating pods in different namespaces.

1.  **Test an allowed namespace (e.g., `default`):**
    Create a file `test-pod-allowed.yaml`:

    ```yaml
    apiVersion: v1
    kind: Pod
    metadata:
      name: test-pod-allowed
    spec:
      containers:
        - name: nginx
          image: nginx:latest # Image from Docker Hub
    ```

    Apply it:

    ```sh
    kubectl apply -f test-pod-allowed.yaml
    # Expected: pod/test-pod-allowed created
    ```

2.  **Test a restricted namespace:**
    First, create a new namespace:
    ```sh
    kubectl create namespace production
    ```
    Create a file `test-pod-denied.yaml`:
    ```yaml
    apiVersion: v1
    kind: Pod
    metadata:
      name: test-pod-denied
      namespace: production
    spec:
      containers:
        - name: nginx
          image: nginx:latest # Image from Docker Hub
    ```
    Apply it:
    ```sh
    kubectl apply -f test-pod-denied.yaml
    # Expected: Error from server (Forbidden):
    # admission webhook "policy-engine-validator.trivago.com" denied the request:
    # Can not use images from Docker Hub in Namespace - production
    ```

## Configuration

The list of namespaces allowed to use Docker Hub images is currently hardcoded in `internal/routes/validate.go`. To customize the policy, modify the `dockerHubAllowedNamespaces` slice.

```go
// internal/routes/validate.go
var (
	dockerHubAllowedNamespaces = []string{"default", "kube-system"} // Modify this list
// ...
)
```