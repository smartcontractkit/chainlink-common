# example.com/consumer
## Config
- Endpoint: Endpoint is the upstream URL to dial.
- Retry: Retry tunes how a failed request is repeated.
# example.com/upstream
## Retry
- Backoff: Backoff multiplies the delay after each failed attempt.
- MaxAttempts: MaxAttempts is the total tries, including the first.
## Settings
- Region: Region is the deployment region.
- Tenant: Tenant scopes every request. Leave empty for the shared tenant.
