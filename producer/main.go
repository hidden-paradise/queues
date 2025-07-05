package main

import (
	"context"
	cryptoRand "crypto/rand"
	"encoding/hex"
	"fmt"
	"github.com/redis/go-redis/v9"
	mathRand "math/rand"
	"time"
)

func randomHash(n int) (string, error) {
	b := make([]byte, n)
	_, err := cryptoRand.Read(b)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func main() {
	ctx := context.Background()

	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   0,
	})

	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		panic(fmt.Sprintf("could not connect to Redis: %v", err))
	}

	fmt.Println("Producer started")

	for {
		hash, err := randomHash(16)
		if err != nil {
			fmt.Printf("error generating random hash: %v\n", err)
			continue
		}

		err = rdb.LPush(ctx, "myqueue", hash).Err()
		if err != nil {
			fmt.Printf("error writing to redis: %v\n", err)
			continue
		}

		fmt.Printf("Pushed to queue: %s\n", hash)

		delay := time.Duration(mathRand.Intn(1900)+100) * time.Millisecond
		time.Sleep(delay)
	}
}
