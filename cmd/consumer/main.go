package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
)

type Transaction struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Amount    float64   `json:"amount"`
	Currency  string    `json:"currency"`
	Timestamp time.Time `json:"timestamp"`
	Location  string    `json:"location"`
}

var (
	ctx         = context.Background()
	txProcessed = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "transactions_processed_total",
		Help: "Total number of transactions processed",
	})
	txErrors = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "transactions_errors_total",
		Help: "Total number of DB insert errors",
	})
	txAnomalies = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "transactions_anomalies_total",
		Help: "Total number of anomalies detected",
	})
)

func initMetrics() {
	prometheus.MustRegister(txProcessed, txErrors, txAnomalies)

	go func() {
		log.Println("🔭 Prometheus metrics exposed at :2112/metrics")
		http.Handle("/metrics", promhttp.Handler())
		log.Fatal(http.ListenAndServe(":2112", nil))
	}()
}

func connectRedis() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
}

func getCountry(city string) string {
	city = strings.ToLower(city)
	switch city {
	case "delhi", "mumbai", "bangalore", "pune", "kolkata":
		return "India"
	case "new york", "chicago", "san francisco":
		return "USA"
	case "london", "manchester":
		return "UK"
	case "sydney", "melbourne":
		return "Australia"
	default:
		return "Unknown"
	}
}

func checkAnomalies(rdb *redis.Client, txn Transaction) []string {
	flags := []string{}

	// High frequency: >3 txns in 1 minute
	key := "txn:" + txn.UserID
	rdb.LPush(ctx, key, txn.Timestamp.Format(time.RFC3339))
	rdb.LTrim(ctx, key, 0, 9)

	times, err := rdb.LRange(ctx, key, 0, -1).Result()
	if err == nil && len(times) >= 3 {
		var count int
		now := txn.Timestamp
		for _, t := range times {
			ts, _ := time.Parse(time.RFC3339, t)
			if now.Sub(ts) < time.Minute {
				count++
			}
		}
		if count >= 3 {
			flags = append(flags, "⚠️ High frequency transactions")
		}
	}

	// Location change
	lastLocKey := "last_location:" + txn.UserID
	prevLoc, err := rdb.Get(ctx, lastLocKey).Result()
	if err == redis.Nil {
		_ = rdb.Set(ctx, lastLocKey, txn.Location, 0).Err()
	} else if err == nil {
		if getCountry(prevLoc) != getCountry(txn.Location) {
			flags = append(flags, fmt.Sprintf("⚠️ Location change: %s → %s", prevLoc, txn.Location))
		}
		_ = rdb.Set(ctx, lastLocKey, txn.Location, 0).Err()
	}

	// Odd hour + high value
	if txn.Timestamp.Hour() >= 0 && txn.Timestamp.Hour() < 4 && txn.Amount > 50000 {
		flags = append(flags, "⚠️ High-value transaction at odd hour")
	}

	return flags
}

func main() {
	initMetrics()
	rdb := connectRedis()
	defer rdb.Close()

	// DB connect
	db, err := sql.Open("postgres", "postgres://kartikparsoya:secret@localhost:5433/finance?sslmode=disable")
	if err != nil {
		log.Fatal("DB connect error:", err)
	}
	defer db.Close()

	// Kafka reader
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{"localhost:9092"},
		Topic:   "transactions",
		GroupID: "finance-group",
	})
	defer reader.Close()

	log.Println("🚀 Consumer started...")

	for {
		m, err := reader.ReadMessage(ctx)
		if err != nil {
			log.Println("Kafka read error:", err)
			continue
		}

		var txn Transaction
		if err := json.Unmarshal(m.Value, &txn); err != nil {
			log.Println("JSON decode error:", err)
			continue
		}

		// Store in DB
		_, err = db.Exec(`INSERT INTO transactions (id, user_id, amount, currency, timestamp, location)
		                  VALUES ($1, $2, $3, $4, $5, $6)`,
			txn.ID, txn.UserID, txn.Amount, txn.Currency, txn.Timestamp, txn.Location)

		if err != nil {
			log.Println("❌ DB insert error:", err)
			txErrors.Inc()
			continue
		}
		txProcessed.Inc()
		log.Printf("✅ Stored: %s | ₹%.2f from %s", txn.UserID, txn.Amount, txn.Location)

		// Check anomalies
		flags := checkAnomalies(rdb, txn)
		if len(flags) > 0 {
			txAnomalies.Inc()
			for _, f := range flags {
				log.Printf("🚨 Anomaly for %s: %s", txn.UserID, f)
			}
		}
	}
}
