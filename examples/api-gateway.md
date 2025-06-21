# Example Configuration - API Gateway Performance Testing

This example demonstrates how to configure Galick for testing an API gateway with multiple endpoints.

## Configuration

```yaml
default:
  environment: dev
  scenario: normal
  output_dir: output

environments:
  dev:
    base_url: http://localhost:8080
    headers:
      Content-Type: application/json
  staging:
    base_url: https://api-staging.example.com
    headers:
      Content-Type: application/json
      Authorization: Bearer ${API_TOKEN}
  production:
    base_url: https://api.example.com
    headers:
      Content-Type: application/json
      Authorization: Bearer ${API_TOKEN}

scenarios:
  normal:
    rate: 50/s
    duration: 60s
    targets:
      - GET /api/products
      - GET /api/categories
      - GET /api/users/me
      - GET /api/settings
  heavy:
    rate: 200/s
    duration: 120s
    targets:
      - GET /api/products
      - GET /api/categories
      - GET /api/users/me
      - POST /api/orders
  spike:
    rate: 500/s
    duration: 30s
    targets:
      - GET /api/products
      - POST /api/orders

report:
  formats:
    - json
    - html
    - markdown
  thresholds:
    p95: 150ms
    p99: 300ms
    success_rate: 99.5

hooks:
  pre: ./scripts/setup-test-data.sh
  post: ./scripts/cleanup-test-data.sh
```

## Running The Tests

```bash
# Run the default scenario against the dev environment
galick run

# Run the heavy scenario against staging
galick run heavy --env staging

# Run the spike test against production (be careful!)
galick run spike --env production

# Generate HTML reports for a previous test
galick report spike --env staging --format html
```

## CI/CD Integration

This example includes a GitHub Actions workflow that runs the tests against the staging environment on every pull request.

See the `.github/workflows/load-test.yml` file for details.
