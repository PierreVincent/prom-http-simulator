# Prometheus HTTP Metrics Simulator

This service simulates Prometheus metrics for what would be a simple HTTP API microservice.

The simplest way to use this simulator is through the published docker image:

```
docker run -p 8080:8080 pierrevincent/prom-http-simulator:0.1
```

## Exposed metrics

It exposes the following Prometheus metrics under the `/metrics` endpoint:
- Standard Golang metrics
- `http_requests_total`: Request counter, label by `endpoint` and `status`
- `http_request_duration_milliseconds`: Request latency histogram

## Runtime options

Endpoints, request rate and latency profiles are hardcoded, but come with some uncertainty so that metrics look somewhat realistic.

It is also possible to simulate variation of metrics while the service is running, 

### Spike Mode

Under spike mode, the number of requests is multiplied by a factor between 5 and 15, and latency is doubled.

Spike mode can be on, off or random. Changing spike mode can be done with:
```
# ON
curl -X POST http://SERVICE_URL:8080/spike/on

# OFF
curl -X POST http://SERVICE_URL:8080/spike/off

# RANDOM
curl -X POST http://SERVICE_URL:8080/spike/random
```

### Error rate

Error rate by default is 1%. It can be changed to a number between 0 and 100 with:
```
# Setting error rate to 50%
curl -H 'Content-Type: application/json' -X PUT -d '{"error_rate": 50}' http://SERVICE_URL:8080/error_rate
```
