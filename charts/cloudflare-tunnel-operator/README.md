# Cloudflare Tunnel Operator

A Kubernetes operator to manage Cloudflare Tunnels.

## Installation

### Using Helm

To install the operator with the optional `cloudflared` connector:

```bash
helm install my-operator ./charts/cloudflare-tunnel-operator \
  --set cloudflared.enabled=true \
  --set cloudflared.tunnelToken="YOUR_TUNNEL_TOKEN_HERE"
```

If you already have a secret with the tunnel token:

```bash
helm install my-operator ./charts/cloudflare-tunnel-operator \
  --set cloudflared.enabled=true \
  --set cloudflared.existingSecretName="my-existing-secret" \
  --set cloudflared.existingSecretNamespace="my-namespace"
```

## Configuration

| Parameter | Description | Default |
| :--- | :--- | :--- |
| `cloudflared.enabled` | Enable the cloudflared connector deployment | `false` |
| `cloudflared.tunnelToken` | The Cloudflare tunnel token | `""` |
| `cloudflared.existingSecretName` | Use an existing secret for the token | `""` |
| `cloudflared.existingSecretNamespace` | Namespace of the existing secret | `""` |
