package main

import (
	"context"
	"encoding/json"
	"log"
	"math/rand"
	"time"

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

var locations = []string{"Mumbai", "Bangalore", "Delhi", "Jaipur", "Chennai"}
var currencies = []string{"INR", "USD", "EUR"}

func generateTransaction() Transaction {
	return Transaction{
		ID:        randomID("tx"),
		UserID:    randomID("user"),
		Amount:    rand.Float64()*20000 + 100,
		Currency:  currencies[rand.Intn(len(currencies))],
		Timestamp: time.Now().UTC(),
		Location:  locations[rand.Intn(len(locations))],
	}
}

func randomID(prefix string) string {
	return prefix + "_" + randomString(6)
}

func randomString(n int) string {
	letters := []rune("abcdefghijklmnopqrstuvwxyz0123456789")
	s := make([]rune, n)
	for i := range s {
		s[i] = letters[rand.Intn(len(letters))]
	}
	return string(s)
}

func main() {
	rand.Seed(time.Now().UnixNano())

	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers:  []string{"localhost:9092"},
		Topic:    "transactions",
		Balancer: &kafka.LeastBytes{},
	})
	defer writer.Close()

	for {
		txn := generateTransaction()
		msgBytes, err := json.Marshal(txn)
		if err != nil {
			log.Println("Failed to serialize transaction:", err)
			continue
		}

		err = writer.WriteMessages(context.Background(),
			kafka.Message{
				Key:   []byte(txn.UserID),
				Value: msgBytes,
			})
		if err != nil {
			log.Println("Kafka write error:", err)
		} else {
			log.Printf("Produced: %+v\n", txn)
		}

		time.Sleep(2 * time.Second)
	}
}
