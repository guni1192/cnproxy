package opentelemetry

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
)

// func Connect(targetURL string) (*grpc.ClientConn, error) {
// 	conn, err := grpc.NewClient(targetURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to connect to otlp server: %w", err)
// 	}
//
// 	return conn, nil
// }

func SetupMetricsProvider(ctx context.Context, res *resource.Resource) (func(context.Context) error, error) {
	metricsExporter, err := otlpmetricgrpc.New(ctx, otlpmetricgrpc.WithInsecure())
	if err != nil {
		return nil, fmt.Errorf("failed to create the collector exporter: %w", err)
	}

	meterProvider := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricsExporter)),
		sdkmetric.WithResource(res),
	)
	otel.SetMeterProvider(meterProvider)

	return meterProvider.Shutdown, nil
}
