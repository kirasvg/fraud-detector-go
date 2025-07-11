# 💸 fraud-detector-go

Real-time transaction anomaly detection in Go using Kafka, Redis, and PostgreSQL.

![grafana](docs/demo.gif)

## 🔧 Features

- Kafka consumer for financial transactions
- Detects anomalies (high-frequency, location change, odd-hour high-value)
- Real-time metrics via Prometheus
- Beautiful dashboards via Grafana
- Docker-compose setup for instant boot

## 🚀 Quick Start

```bash
git clone https://github.com/kartikparsoya/fraud-detector-go
cd fraud-detector-go
docker-compose up -d
go run cmd/consumer/main.go
go run cmd/producer/main.go
