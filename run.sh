export OTEL_EXPORTER_OTLP_METRICS_PROTOCOL=grpc
export OTEL_EXPORTER_OTLP_ENDPOINT=http://127.0.0.1:4317
export OTEL_LOG_LEVEL=debug
export OTEL_INSECURE=true

./bin/cnproxy --enable-metrics --port 1192
