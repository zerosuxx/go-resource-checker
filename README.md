# go-resource-checker

# Usage
```
RESOURCE_URLS='["http://localhost:1234"]' go run checker.go server -addr=localhost:8000
go run checker.go check -url=[tcp|udp|http|https]://localhost:1234 -timeout=10
```