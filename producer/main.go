package main

import (
	"context"
	cryptoRand "crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	"homework-1/internal"
	mathRand "math/rand"
	"net/http"
	"os"
	"time"
)

const (
	metricsPort     = ":9100"
	nameByteSize    = 8
	payloadByteSize = 16
)

var (
	producedJobs = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "producer_jobs_total",
			Help: "Total number of jobs produced",
		},
		[]string{"producer_id"},
	)
)

func init() {
	prometheus.MustRegister(producedJobs)

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(metricsPort, nil)
	}()
}

type Producer struct {
	ID        string
	QueueName string
	RDB       *redis.Client
	Ctx       context.Context
}

func NewProducer(id string, queueName string, redisAddr string) *Producer {
	rdb := redis.NewClient(&redis.Options{
		Addr: redisAddr,
		DB:   0,
	})

	ctx := context.Background()
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		panic(fmt.Sprintf("could not connect to Redis: %v", err))
	}

	return &Producer{
		ID:        id,
		QueueName: queueName,
		RDB:       rdb,
		Ctx:       ctx,
	}
}

func (p *Producer) Run() {
	fmt.Println("Producer started")

	for {
		job, err := p.createJob()
		if err != nil {
			fmt.Printf("failed to create job: %v\n", err)
			continue
		}

		if err := p.pushJob(job); err != nil {
			fmt.Printf("failed to push job: %v\n", err)
			continue
		}

		producedJobs.WithLabelValues(p.ID).Inc()

		queueLen, err := p.RDB.LLen(p.Ctx, p.QueueName).Result()
		if err != nil {
			fmt.Printf("failed to get queue length: %v\n", err)
			continue
		}

		fmt.Printf("Pushed job: %+v | Queue length: %d\n", job, queueLen)

		delay := time.Duration(mathRand.Intn(2900)+100) * time.Millisecond
		time.Sleep(delay)
	}
}

func (p *Producer) createJob() (internal.Job, error) {
	name, err := randomHex(nameByteSize)
	if err != nil {
		return internal.Job{}, err
	}

	payload, err := randomHex(payloadByteSize)
	if err != nil {
		return internal.Job{}, err
	}

	return internal.NewJob(fmt.Sprintf("Job-%s", name), payload)
}

func (p *Producer) pushJob(job internal.Job) error {
	jobBytes, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("marshal job: %w", err)
	}

	return p.RDB.LPush(p.Ctx, p.QueueName, jobBytes).Err()
}

func randomHex(n int) (string, error) {
	b := make([]byte, n)
	_, err := cryptoRand.Read(b)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func main() {
	queueName := os.Getenv("QUEUE_NAME")
	if queueName == "" {
		queueName = "jobs"
	}

	producerID := os.Getenv("PRODUCER_ID")
	if producerID == "producer-1" || producerID == "" {
		randomBytes := make([]byte, 2) // 2 байта = 4 hex-символа
		_, err := cryptoRand.Read(randomBytes)
		if err != nil {
			panic(fmt.Sprintf("failed to generate random producer ID: %v", err))
		}
		producerID = fmt.Sprintf("producer-%s", hex.EncodeToString(randomBytes))
	}

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	producer := NewProducer(producerID, queueName, redisAddr)
	producer.Run()
}
