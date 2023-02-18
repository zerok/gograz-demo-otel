# GoGraz demo: OpenTelemetry

## Goal

This demo should give you a brief introduction into OpenTelemetry and how to
use it to send tracing data from various services via Grafana Agent to a
Grafana Tempo instance.

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
