# cnproxy

cnproxy is Cloud Native forward proxy.


## Features

* http/https supprot (http CONNECT method)
* http healthcheck
* structured logging
* metrics
* FQDN allow list for access control
* tracing (TODO)


## Usage

### Basic Usage

1. Launch the proxy server.

```shell
./run.sh
```

2. send a request to the proxy server.

```shell
curl -s --proxy http://localhost:1192 https://example.com
```

### FQDN Restriction

You can restrict accessible FQDNs using the `--allowed-fqdn` flag. Only connections to specified FQDNs will be allowed.

```shell
# Allow specific FQDNs
cnproxy --allowed-fqdn example.com --allowed-fqdn api.github.com

# Using environment variable
export CNPROXY_ALLOWED_FQDN="example.com,google.com"
cnproxy
```

When an FQDN is not in the allow list, the proxy will return `403 Forbidden`.

```shell
# Allowed FQDN - succeeds
curl -x http://localhost:8080 http://example.com

# Blocked FQDN - returns 403 Forbidden
curl -x http://localhost:8080 http://blocked-site.com
```

### Command-line Options

```
--port value, -p value         port number (default: 8080) [$CNPROXY_PORT]
--address value, -a value      address (default: "0.0.0.0") [$CNPROXY_ADDRESS]
--enable-metrics               enable metrics (OTLP) (default: false) [$CNPROXY_ENABLE_METRICS]
--allowed-fqdn value           allowed FQDNs for proxy connections (can be specified multiple times) [$CNPROXY_ALLOWED_FQDN]
```
