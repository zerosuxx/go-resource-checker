# go-resource-checker

## Available environments
- `SLACK_WEBHOOK_URL`
- `FORCE_SUCCESS_RESPONSE=[0|1]`
- `RESOURCE_URLS`

## Usage
```
RESOURCE_URLS='["http://localhost:1234"]' go run checker.go server -addr=localhost:8000
go run checker.go check -url=[tcp|udp|http|https]://localhost:1234 -timeout=10
```