# Example Configuration - E-commerce Website Testing

This example demonstrates how to configure Galick for testing an e-commerce website with realistic user flows.

## Configuration

```yaml
default:
  environment: dev
  scenario: browse
  output_dir: output

environments:
  dev:
    base_url: http://localhost:3000
    headers:
      Content-Type: application/json
  staging:
    base_url: https://staging.shop.example.com
    headers:
      Content-Type: application/json

scenarios:
  browse:
    rate: 20/s
    duration: 60s
    targets:
      - GET /
      - GET /products
      - GET /products/1
      - GET /products/2
      - GET /categories
  cart:
    rate: 10/s
    duration: 60s
    targets:
      - GET /products/1
      - POST /cart/add
      - GET /cart
      - PUT /cart/update
  checkout:
    rate: 5/s
    duration: 60s
    targets:
      - GET /cart
      - POST /checkout/start
      - POST /checkout/payment
      - GET /checkout/confirmation

report:
  formats:
    - json
    - markdown
  thresholds:
    p95: 500ms
    success_rate: 98.0

hooks:
  pre: ./scripts/reset-test-db.sh
  post: ./scripts/analyze-results.sh
```

## Running The Tests

```bash
# Run the browse scenario against the dev environment
galick run browse

# Run the cart scenario against staging
galick run cart --env staging

# Run all scenarios in sequence
for scenario in browse cart checkout; do
  galick run $scenario --env staging
done
```

## Test Data Setup

The pre-hook script (`reset-test-db.sh`) populates the test database with sample products and categories to ensure consistent test results.
