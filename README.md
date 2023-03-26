# GoGraz demo: OpenTelemetry

## Goal

This demo should give you a brief introduction into OpenTelemetry and how to use it to send tracing data from various services via Grafana Agent to a Grafana Tempo instance.

## Architecture

```mermaid
graph LR
    subraph TD
        Caddy -> Website
        Website -> Backend
        Caddy -> GrafanaAgent
    Website -> GrafanaAgent
    Backend -> GrafanaAgent
    GrafanaAgent -> GrafanaTempo
```

## Setup 

At this point, the setup relies on a Grafana Cloud account.
Once you have that, set the following two environment variables and start the Docker-Compose environment:

```
export GRAFANA_USERNAME=...
export GRAFANA_API_KEY=...

go run mage.go build
docker-compose up
```

Then do a simple HTTP request to http://localhost:8080 and you should see traces ending up in Grafana:

```
curl http://localhost:8080
```