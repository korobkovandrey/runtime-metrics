#!/bin/sh

#for v1.62.* +

mkdir -p golangci-lint
golangci-lint run -c .golangci-1.62.yml | jq > ./golangci-lint/report.json
