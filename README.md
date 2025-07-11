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


```

## 📈 Metrics Exposed

The following metrics are exposed via Prometheus:

- `transactions_processed_total` — Count of total processed transactions
- `transactions_anomalies_total` — Count of flagged anomalies
- `transactions_errors_total` — Count of DB or processing errors

> Use these in Grafana with PromQL to create dashboards, alerts, and time-series graphs.

---

## 🧠 Anomalies Detected

| Type              | Detection Logic                          |
|-------------------|-------------------------------------------|
| **High Frequency** | ≥ 3 transactions within 60 seconds        |
| **Location Change**| IP / region changes across countries      |
| **Odd Hour + High ₹** | ₹50,000+ transaction between 12AM–4AM   |

---

## 🛠 Tech Stack

- **Language**: Go (Golang)
- **Queue**: Kafka (Redpanda-compatible)
- **Databases**: PostgreSQL + Redis
- **Monitoring**: Prometheus + Grafana
- **Containerized with**: Docker + Docker Compose

---

## 🧪 To Do (Open for PRs)

- [ ] 🔄 **REST API** for querying recent anomalies
- [ ] 🔔 **Alerting System** (Slack, Email integration)
- [ ] 🧠 **ML Scoring Service** via gRPC/REST (future)
- [ ] ⏪ **Replay Engine** to test historical transactions

Feel free to open issues or PRs for any of the above! 🙌

---

## 🤝 Contributing

We love contributors! Here’s how to get started:

1. 🍴 **Fork** the repository
2. 📥 **Clone** your fork locally
3. ✍️ Make your changes with meaningful commits
4. 📤 Submit a **pull request** with clear description
5. 🧪 Add your name to `CONTRIBUTORS.md` (coming soon)

Thanks for helping improve `fraud-detector-go` 💛
