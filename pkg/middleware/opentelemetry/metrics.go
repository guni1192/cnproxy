package opentelemetry

import (
	"go.opentelemetry.io/otel/metric"
)

type ProxyMetrics struct {
	TotalRequests metric.Int64Counter
}
