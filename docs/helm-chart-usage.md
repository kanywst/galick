# Galick Helm Chart Usage Guide

Galick Helm Chart is a convenient tool for running HTTP load tests on Kubernetes clusters.

## Overview

This Helm Chart provides the following features:

- **One-time Load Testing**: Execute load tests as Kubernetes Jobs
- **Scheduled Load Testing**: Run periodic load tests using CronJobs
- **Configuration Management**: Manage test configurations using ConfigMaps
- **Result Persistence**: Persist test results using PersistentVolumes
- **Demo Server**: Built-in demo API server for testing

## Prerequisites

- Kubernetes cluster (v1.19+)
- Helm 3.0+
- kubectl command-line tool

## Installation

### 1. Helm Chart Installation

```bash
# Install with default settings
helm install my-galick ./helm-chart/galick

# Install with custom settings
helm install my-galick ./helm-chart/galick -f custom-values.yaml

# Install to specific namespace
helm install my-galick ./helm-chart/galick -n galick-ns --create-namespace
```

### 2. Install with Demo Server

```bash
helm install my-galick ./helm-chart/galick \
  --set galick.demoServer.enabled=true
```

## Configuration Options

### Basic Configuration

| Parameter | Description | Default Value |
|-----------|-------------|---------------|
| `replicaCount` | Number of replicas | `1` |
| `image.repository` | Galick image repository | `galick` |
| `image.tag` | Image tag | `latest` |
| `image.pullPolicy` | Image pull policy | `IfNotPresent` |

### Galick-specific Configuration

| Parameter | Description | Default Value |
|-----------|-------------|---------------|
| `galick.job.enabled` | Run as Job | `true` |
| `galick.cronjob.enabled` | Run as CronJob | `false` |
| `galick.cronjob.schedule` | CronJob schedule | `"0 2 * * *"` |
| `galick.persistence.enabled` | Enable result persistence | `true` |
| `galick.persistence.size` | Storage size | `1Gi` |
| `galick.demoServer.enabled` | Enable demo server | `false` |

## Usage Examples

### 1. Basic Load Test

```bash
# Install with default settings
helm install basic-test ./helm-chart/galick

# Check test execution status
kubectl get jobs
kubectl logs -l app.kubernetes.io/name=galick -f
```

### 2. Load Test with Custom Configuration

Create `custom-values.yaml`:

```yaml
galick:
  config:
    environments:
      production:
        base_url: "https://api.example.com"
        headers:
          Authorization: "Bearer YOUR_TOKEN"
    scenarios:
      stress:
        rate: "100/s"
        duration: "5m"
        targets:
          - "GET /api/v1/products"
          - "GET /api/v1/users"
  
  env:
    - name: GALICK_DEFAULT_ENVIRONMENT
      value: "production"
    - name: GALICK_DEFAULT_SCENARIO
      value: "stress"
```

```bash
helm install stress-test ./helm-chart/galick -f custom-values.yaml
```

### 3. Scheduled Execution

```yaml
galick:
  job:
    enabled: false
  cronjob:
    enabled: true
    schedule: "0 */6 * * *"  # Execute every 6 hours
  config:
    scenarios:
      monitoring:
        rate: "5/s"
        duration: "2m"
        targets:
          - "GET /health"
```

```bash
helm install monitoring-test ./helm-chart/galick -f monitoring-values.yaml
```

### 4. Test with Demo Server

```bash
# Install with demo server
helm install demo-test ./helm-chart/galick \
  --set galick.demoServer.enabled=true \
  --set galick.config.environments.dev.base_url="http://demo-server:8080"

# Access demo server
kubectl port-forward svc/demo-server 8080:8080
curl http://localhost:8080/api/health
```

## Checking Results

### 1. Check Test Execution Status

```bash
# Check Job status
kubectl get jobs -l app.kubernetes.io/name=galick

# Check logs
kubectl logs -l app.kubernetes.io/name=galick -f

# Check CronJob (for scheduled execution)
kubectl get cronjobs -l app.kubernetes.io/name=galick
```

### 2. Retrieve Test Results

```bash
# Check PVC
kubectl get pvc -l app.kubernetes.io/name=galick

# Access result files
kubectl exec -it $(kubectl get pods -l app.kubernetes.io/name=galick -o jsonpath="{.items[0].metadata.name}") \
  -- ls -la /data/output

# Download result files
kubectl cp $(kubectl get pods -l app.kubernetes.io/name=galick -o jsonpath="{.items[0].metadata.name}"):/data/output ./local-results
```

### 3. Generate Reports

After retrieving result files, generate reports locally:

```bash
# Generate HTML report
galick report --results ./local-results/dev/simple/results.json --format html

# Generate Markdown report
galick report --results ./local-results/dev/simple/results.json --format markdown
```

## Advanced Configuration

### 1. Distributed Load Testing

Run multiple Jobs in parallel for high-load testing:

```yaml
galick:
  job:
    parallelism: 5  # Run with 5 Pods in parallel
    completions: 5
  config:
    scenarios:
      distributed:
        rate: "200/s"  # 200/s per Pod, 1000/s total
        duration: "10m"
```

### 2. Environment Variable Configuration

```yaml
galick:
  env:
    - name: GALICK_DEFAULT_ENVIRONMENT
      value: "staging"
    - name: TARGET_HOST
      value: "api.staging.example.com"
  config:
    environments:
      staging:
        base_url: "https://$(TARGET_HOST)"
```

### 3. Custom Volume Addition

```yaml
galick:
  extraVolumes:
    - name: custom-scripts
      spec:
        configMap:
          name: custom-scripts
  extraVolumeMounts:
    - name: custom-scripts
      mountPath: /custom-scripts
      readOnly: true
```

## Troubleshooting

### Common Issues

1. **Job does not complete**
   ```bash
   # Check Job details
   kubectl describe job <job-name>
   
   # Check Pod logs
   kubectl logs -l app.kubernetes.io/name=galick
   ```

2. **PVC cannot be mounted**

   ```bash
   # Check PVC status
   kubectl get pvc
   kubectl describe pvc <pvc-name>

   # Check StorageClass
   kubectl get storageclass
   ```

3. **Configuration not applied**

   ```bash
   # Check ConfigMap contents
   kubectl get configmap <configmap-name> -o yaml

   # Check Helm configuration
   helm get values <release-name>
   ```

### Log Checking Methods

```bash
# Real-time logs
kubectl logs -l app.kubernetes.io/name=galick -f

# Previous logs
kubectl logs -l app.kubernetes.io/name=galick --previous

# Specific Pod logs
kubectl logs <pod-name>
```

## Uninstallation

```bash
# Delete Helm release
helm uninstall my-galick

# Delete PVC (if needed)
kubectl delete pvc -l app.kubernetes.io/name=galick

# Delete namespace (if using dedicated namespace)
kubectl delete namespace galick-ns
```

## Security Considerations

1. **ServiceAccount**: Creates minimal-privilege ServiceAccount by default
2. **Security Context**: Runs containers as non-root user
3. **Network Policy**: Configure network policies as needed
4. **Secrets**: Use Kubernetes Secrets for authentication information

## Performance Optimization

1. **Resource Limits**: Set appropriate CPU/memory limits
2. **Node Selection**: Run on high-performance nodes
3. **Distributed Execution**: Use multiple Pods for large-scale tests
4. **Storage**: Use high-performance storage classes

## Reference Links

- [Galick Official Documentation](https://github.com/kanywst/galick)
- [Kubernetes Jobs](https://kubernetes.io/docs/concepts/workloads/controllers/job/)
- [Kubernetes CronJobs](https://kubernetes.io/docs/concepts/workloads/controllers/cron-jobs/)
- [Helm Official Documentation](https://helm.sh/docs/)
