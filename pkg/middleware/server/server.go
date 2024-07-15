package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/guni1192/cnproxy/pkg/middleware/logger"
	"github.com/guni1192/cnproxy/pkg/middleware/opentelemetry"
	"github.com/guni1192/cnproxy/pkg/service"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

type CNProxyServer struct {
	Port    uint
	Address string

	EnableMetrics bool
}

func (s *CNProxyServer) Serve() error {
	ctx := context.Background()

	h := &service.CNProxyHandler{
		Logger: logger.New(),
	}

	h.Logger.Info("server info", "port", s.Port, "address", s.Address, "enable_metrics", s.EnableMetrics)

	if s.EnableMetrics {
		res := resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("cnproxy"),
		)

		shutdownMetricsProvider, err := opentelemetry.SetupMetricsProvider(ctx, res)
		if err != nil {
			return fmt.Errorf("failed to setup metrics provider: %v", err)
		}
		defer shutdownMetricsProvider(ctx)

		meter := otel.Meter("cnproxy")

		requestCount, err := meter.Int64Counter("request_count")
		if err != nil {
			return fmt.Errorf("failed to create counter: %v", err)
		}
		h.ProxyMetrics = &opentelemetry.ProxyMetrics{
			TotalRequests: requestCount,
		}
		h.ProxyMetrics.TotalRequests.Add(ctx, 0)

		host := fmt.Sprintf("%s:%d", s.Address, s.Port)
		h.Logger.Info("listening", "address", s.Address, "port", s.Port)
		return http.ListenAndServe(host, h)

	} else {
		host := fmt.Sprintf("%s:%d", s.Address, s.Port)
		h.Logger.Info("listening", "address", s.Address, "port", s.Port)
		return http.ListenAndServe(host, h)
	}
}
