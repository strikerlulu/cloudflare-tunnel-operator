# Cloudflare Tunnel Operator

An operator to manage Cloudflare Tunnel resources.

## Architecture

- **CRD (`Tunnel`)**: Defines the desired state (domain, service, port).
- **Controller**: Watches for `Tunnel` resources, reconciles Cloudflare state, and manages lifecycle using finalizers.
- **Global Configuration**:
  - `CLOUDFLARE_ACCOUNT_ID`: Set in your `values.yaml` (Helm) or Deployment environment variables.
  - `CLOUDFLARE_API_TOKEN`: Managed via Kubernetes Secret.

## How it works

1. User creates a `Tunnel` CR.
2. Operator adds a finalizer to the `Tunnel` CR.
3. Operator provisions the Tunnel and DNS record.
4. On `Tunnel` deletion, the operator removes the Cloudflare resources and the finalizer.

## Prerequisites

- Kubernetes cluster
- `cloudflared` deployed in the cluster
- Cloudflare API Token (with Tunnel/DNS permissions)

## Setup and Installation

1. **Helm Chart**: The Helm chart (operator, CRDs, RBAC, and an optional `cloudflared`
   connector) lives at [`./dist/chart`](./dist/chart) and is published to GHCR as an OCI chart.

   Install the published chart:
   ```bash
   helm install cloudflare-tunnel-operator \
     oci://ghcr.io/strikerlulu/charts/cloudflare-tunnel-operator \
     --version <version> \
     --namespace cloudflare-tunnel-operator-system --create-namespace
   ```
   Released versions are listed under [Packages](https://github.com/strikerlulu/cloudflare-tunnel-operator/pkgs/container/charts%2Fcloudflare-tunnel-operator).

   For configuration details (including the optional `cloudflared` connector) and local
   testing instructions, refer to [`dist/chart/README.md`](./dist/chart/README.md).

2. **Configuration**:
   - `CLOUDFLARE_ACCOUNT_ID`: Set in your `values.yaml` (Helm) or Deployment environment variables.
   - `CLOUDFLARE_API_TOKEN`: Managed via Kubernetes Secret.
   - **Default Tunnel Name**: The operator expects a `sharedTunnelName` in the `Tunnel` CR, which refers to your existing Cloudflare Tunnel instance. You must provide a valid `sharedTunnelName` (e.g., `my-master-tunnel`) in your `Tunnel` resource specification.

3. **Create secret** (in your preferred namespace, e.g., `cloudflare-system`):
   ```bash
   kubectl create namespace cloudflare-system
   kubectl create secret generic cloudflare-token --from-literal=api-token=<YOUR_TOKEN> -n cloudflare-system
   ```

4. **Apply CR**:
   ```yaml
   apiVersion: networking.strikerlulu.me/v1alpha1
   kind: Tunnel
   metadata:
     name: tunnel-sample
     namespace: default
   spec:
     domain: example.com
     serviceName: my-service
     servicePort: 8080
     secretRef: cloudflare-token
     secretNamespace: cloudflare-system
     sharedTunnelName: my-master-tunnel
     accountID: <YOUR_ACCOUNT_ID>
   ```
   ```bash
   kubectl apply -f tunnel.yaml
   ```

## Development
- Build: `make docker-build`
- Push: `make docker-push IMG=ghcr.io/strikerlulu/cloudflare-tunnel-operator:latest`
