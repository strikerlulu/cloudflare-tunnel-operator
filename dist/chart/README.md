# cloudflare-tunnel-operator

A Helm chart that installs the cloudflare-tunnel-operator controller, its CRDs and RBAC,
and (optionally) a `cloudflared` connector Deployment.

## Installing

From the published OCI chart:

```bash
helm install cloudflare-tunnel-operator \
  oci://ghcr.io/strikerlulu/charts/cloudflare-tunnel-operator \
  --version <version> \
  --namespace cloudflare-tunnel-operator-system --create-namespace
```

From a local checkout:

```bash
helm install cloudflare-tunnel-operator ./dist/chart \
  --namespace cloudflare-tunnel-operator-system --create-namespace
```

## Optional: cloudflared connector

Enable a `cloudflared` Deployment that runs a shared/master tunnel alongside the operator:

```bash
helm install cloudflare-tunnel-operator oci://ghcr.io/strikerlulu/charts/cloudflare-tunnel-operator \
  --set cloudflared.enabled=true \
  --set cloudflared.tunnelToken="YOUR_TUNNEL_TOKEN_HERE"
```

Or, to reuse an existing Secret instead of supplying a token directly:

```bash
helm install cloudflare-tunnel-operator oci://ghcr.io/strikerlulu/charts/cloudflare-tunnel-operator \
  --set cloudflared.enabled=true \
  --set cloudflared.existingSecretName="my-existing-secret" \
  --set cloudflared.existingSecretNamespace="my-namespace"
```

## Testing locally

From the repository root, with Go and Docker installed:

```bash
# 1. Install kind and helm (one-time)
make install-helm
curl -Lo ./kind https://kind.sigs.k8s.io/dl/latest/kind-linux-$(go env GOARCH)
chmod +x ./kind && sudo mv ./kind /usr/local/bin/kind

# 2. Create a local cluster and load the operator image into it
kind create cluster
IMG=controller:latest make docker-build
kind load docker-image controller:latest

# 3. Lint the chart
helm lint ./dist/chart

# 4. Install the chart against the kind cluster
IMG=controller:latest make helm-deploy

# 5. Check the release
make helm-status
kubectl get pods -n cloudflare-tunnel-operator-system
kubectl get crd | grep tunnels.networking.strikerlulu.me
```

To test the optional `cloudflared` connector as well:

```bash
helm upgrade --install cloudflare-tunnel-operator ./dist/chart \
  --namespace cloudflare-tunnel-operator-system --create-namespace \
  --set manager.image.repository=controller \
  --set manager.image.tag=latest \
  --set cloudflared.enabled=true \
  --set cloudflared.tunnelToken="YOUR_TUNNEL_TOKEN_HERE"
```

Tear down with `make helm-uninstall` and `kind delete cluster`.

## Values

See [`values.yaml`](./values.yaml) for the full list of configurable values, including
`manager.*`, `rbac.*`, `crd.*`, `metrics.*`, `certManager.*`, `prometheus.*`,
`networkPolicy.*`, and `cloudflared.*`.
