package main

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	"homework-1/internal"
	"math/rand"
	"net/http"
	"os"
	"time"
)

const (
	metricsPort = ":9100"
	errorChance = 5 // 1 of 5 means 20% chance of error
)

var (
	processedJobs = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "consumer_jobs_total",
			Help: "Total number of jobs processed by consumer",
		},
		[]string{"consumer_id"},
	)
	failedJobs = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "consumer_jobs_failed_total",
			Help: "Total number of jobs failed and returned to queue",
		},
		[]string{"consumer_id"},
	)
)

func init() {
	prometheus.MustRegister(processedJobs)
	prometheus.MustRegister(failedJobs)

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(metricsPort, nil)
	}()
}

type Consumer struct {
	ID        string
	QueueName string
	RDB       *redis.Client
	Ctx       context.Context
}

func NewConsumer(id, queueName, redisAddr string) *Consumer {
	rdb := redis.NewClient(&redis.Options{
		Addr: redisAddr,
		DB:   0,
	})

	ctx := context.Background()
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		panic(fmt.Sprintf("could not connect to Redis: %v", err))
	}

	return &Consumer{
		ID:        id,
		QueueName: queueName,
		RDB:       rdb,
		Ctx:       ctx,
	}
}

func (c *Consumer) Run() {
	fmt.Println("Consumer started")

	for {
		result, err := c.RDB.RPop(c.Ctx, c.QueueName).Result()
		if err == redis.Nil {
			time.Sleep(500 * time.Millisecond)
			continue
		}
		if err != nil {
			fmt.Printf("error popping job from queue: %v\n", err)
			time.Sleep(500 * time.Millisecond)
			continue
		}

		var job internal.Job
		err = json.Unmarshal([]byte(result), &job)
		if err != nil {
			fmt.Printf("error unmarshaling job: %v\n", err)
			continue
		}

		fmt.Printf("Consumed job: %+v\n", job)

		workDuration := time.Duration(rand.Intn(1000)) * time.Millisecond
		time.Sleep(workDuration)

		if rand.Intn(errorChance) == 0 {
			fmt.Printf("error processing job %s, returning to queue\n", job.Name)
			job.Failed = true
			err := c.returnJobToQueue(job)
			if err != nil {
				fmt.Printf("error returning job to queue: %v\n", err)
			}

			failedJobs.WithLabelValues(c.ID).Inc()
			processedJobs.WithLabelValues(c.ID).Inc()
			continue
		}

		fmt.Printf("Job %s processed successfully\n", job.Name)
		processedJobs.WithLabelValues(c.ID).Inc()
	}
}

func (c *Consumer) returnJobToQueue(job internal.Job) error {
	jobBytes, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("error marshaling job: %w", err)
	}

	return c.RDB.LPush(c.Ctx, c.QueueName, jobBytes).Err()
}

func main() {
	queueName := os.Getenv("QUEUE_NAME")
	if queueName == "" {
		queueName = "jobs"
	}

	consumerID := os.Getenv("CONSUMER_ID")
	if consumerID == "consumer-1" || consumerID == "" {
		randomBytes := make([]byte, 2) // 2 байта = 4 hex-символа
		_, err := rand.Read(randomBytes)
		if err != nil {
			panic(fmt.Sprintf("failed to generate random consumer ID: %v", err))
		}
		consumerID = fmt.Sprintf("consumer-%s", hex.EncodeToString(randomBytes))
	}

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	consumer := NewConsumer(consumerID, queueName, redisAddr)
	consumer.Run()
}
