# Elchi Endpoint Discovery Helm Chart

This Helm chart deploys the Elchi Endpoint Discovery plugin on a Kubernetes cluster.

## Prerequisites

- Kubernetes 1.19+
- Helm 3.2.0+

## Installing the Chart

```bash
# Add the repository (when available)
helm repo add elchi https://charts.cloudnativeworks.io
helm repo update

# Install the chart
helm install endpoint-discovery elchi/elchi-endpoint-discovery

# Or install from local chart
helm install endpoint-discovery ./helm/elchi-endpoint-discovery
```

## Configuration

The following table lists the configurable parameters and their default values.

| Parameter | Description | Default |
|-----------|-------------|---------|
| `replicaCount` | Number of replicas | `1` |
| `image.repository` | Image repository | `cloudnativeworks/elchi-endpoint-discovery` |
| `image.tag` | Image tag | `latest` |
| `image.pullPolicy` | Image pull policy | `IfNotPresent` |
| `config.token` | Elchi authentication token | `""` |
| `config.log.level` | Log level | `info` |
| `config.log.format` | Log format (text/json) | `json` |
| `discoveryInterval` | Discovery interval in seconds | `30` |
| `clusterName` | Cluster name (**REQUIRED**) | `""` |
| `resources.requests.cpu` | CPU request | `100m` |
| `resources.requests.memory` | Memory request | `64Mi` |
| `resources.limits.cpu` | CPU limit | `200m` |
| `resources.limits.memory` | Memory limit | `128Mi` |

## Examples

### Basic Installation with Token and Cluster Name

```bash
helm install endpoint-discovery ./helm/elchi-endpoint-discovery \
  --set config.token="your-elchi-token-here" \
  --set clusterName="production-cluster"
```

### Custom Discovery Interval and Resources

```bash
helm install endpoint-discovery ./helm/elchi-endpoint-discovery \
  --set config.token="your-token" \
  --set clusterName="my-cluster" \
  --set discoveryInterval=60 \
  --set resources.requests.memory="128Mi" \
  --set resources.limits.memory="256Mi"
```


### Enable Debug Logging

```bash
helm install endpoint-discovery ./helm/elchi-endpoint-discovery \
  --set config.token="your-token" \
  --set clusterName="my-cluster" \
  --set config.log.level="debug" \
  --set config.log.format="text"
```

## Uninstalling the Chart

```bash
helm uninstall endpoint-discovery
```

## Security

This chart follows security best practices:

- Runs as non-root user (65534)
- Uses read-only root filesystem
- Drops all capabilities
- Minimal RBAC permissions (only nodes, endpoints, services, pods read access)

## Monitoring

The plugin outputs structured logs that can be collected by log aggregation systems like:

- Fluentd/Fluent Bit
- Logstash
- Promtail (Loki)

For JSON logs, set:
```yaml
config:
  log:
    format: "json"
```

## Troubleshooting

### Check Deployment Status
```bash
kubectl get deployments -n <namespace>
kubectl describe deployment <release-name>-elchi-endpoint-discovery
```

### View Pod Status
```bash
kubectl get pods -l app.kubernetes.io/instance=<release-name>
```

### Check Logs
```bash
kubectl logs -l app.kubernetes.io/instance=<release-name> --tail=100
```

### Restart Deployment
```bash
kubectl rollout restart deployment/<release-name>-elchi-endpoint-discovery
```

## Contributing

Please see the main repository for contribution guidelines: https://github.com/CloudNativeWorks/elchi-plugins