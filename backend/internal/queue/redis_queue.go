package queue

import (
	"context"

	"github.com/redis/go-redis/v9"
)

// queueKey is the name of the Redis list we use as our job queue.
const queueKey = "jobs:queue"

// RedisQueue wraps a Redis client to push and pop job ID's.
type RedisQueue struct {
	client *redis.Client
}

// NewRedisQueue creates a queue connected to the given Redis address, e.g. "localhost:6379"
func NewRedisQueue(addr string) *RedisQueue {
	client := redis.NewClient(&redis.Options{
		Addr: addr,
	})
	return &RedisQueue{client: client}
}

// Enqueue pushes a job ID onto the queue for a worker to pick up
func (q *RedisQueue) Enqueue(ctx context.Context, JobID string) error {
	return q.client.LPush(ctx, queueKey, JobID).Err()
}

// Dequeue pushes a job ID onto the queue for a worker to pick up.
func (q *RedisQueue) Dequeue(ctx context.Context) (string, error) {
	// BRPop returns a []string: [Key,value]. We only want the value.
	result, err := q.client.BRPop(ctx, 0, queueKey).Result()
	if err != nil {
		return " ", err
	}
	return result[1], nil
}
