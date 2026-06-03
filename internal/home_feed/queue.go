package home_feed

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	DefaultFeedRefreshStream = "home_feed_refresh"
	DefaultFeedRefreshGroup  = "home_feed_workers"
)

var (
	ErrFeedRefreshUserIDRequired       = errors.New("feed refresh user_id is required")
	ErrFeedRefreshConsumerNameRequired = errors.New("feed refresh consumer name is required")
)

type FeedRefreshJob struct {
	ID          string
	UserID      string
	Reason      string
	RequestedAt time.Time
}

type FeedRefreshQueue interface {
	Enqueue(ctx context.Context, userID, reason string) error
}

type FeedRefreshJobQueue interface {
	FeedRefreshQueue
	EnsureConsumerGroup(ctx context.Context) error
	Read(ctx context.Context, consumerName string, count int, block time.Duration) ([]FeedRefreshJob, error)
	Ack(ctx context.Context, jobIDs ...string) error
}

type RedisFeedRefreshQueue struct {
	client       redis.Cmdable
	stream       string
	group        string
	debounceTTL  time.Duration
	debounceKey  string
	maxLenApprox int64
}

type RedisFeedRefreshQueueOptions struct {
	Stream       string
	Group        string
	DebounceTTL  time.Duration
	DebounceKey  string
	MaxLenApprox int64
}

func NewRedisFeedRefreshQueue(client redis.Cmdable, opts RedisFeedRefreshQueueOptions) *RedisFeedRefreshQueue {
	if opts.Stream == "" {
		opts.Stream = DefaultFeedRefreshStream
	}
	if opts.Group == "" {
		opts.Group = DefaultFeedRefreshGroup
	}
	if opts.DebounceKey == "" {
		opts.DebounceKey = "home_feed_refresh_scheduled"
	}
	if opts.DebounceTTL == 0 {
		opts.DebounceTTL = time.Minute
	}

	return &RedisFeedRefreshQueue{
		client:       client,
		stream:       opts.Stream,
		group:        opts.Group,
		debounceTTL:  opts.DebounceTTL,
		debounceKey:  opts.DebounceKey,
		maxLenApprox: opts.MaxLenApprox,
	}
}

func (q *RedisFeedRefreshQueue) Enqueue(ctx context.Context, userID, reason string) error {
	if userID == "" {
		return ErrFeedRefreshUserIDRequired
	}

	if q.debounceTTL > 0 {
		scheduled, err := q.client.SetNX(ctx, q.debounceKey+":"+userID, "1", q.debounceTTL).Result()
		if err != nil {
			return err
		}
		if !scheduled {
			return nil
		}
	}

	now := time.Now().UTC()
	args := &redis.XAddArgs{
		Stream: q.stream,
		Values: map[string]any{
			"user_id":      userID,
			"reason":       reason,
			"requested_at": now.Format(time.RFC3339Nano),
		},
	}
	if q.maxLenApprox > 0 {
		args.MaxLen = q.maxLenApprox
		args.Approx = true
	}

	if err := q.client.XAdd(ctx, args).Err(); err != nil {
		if q.debounceTTL > 0 {
			_ = q.client.Del(ctx, q.debounceKey+":"+userID).Err()
		}
		return err
	}
	return nil
}

func (q *RedisFeedRefreshQueue) EnsureConsumerGroup(ctx context.Context) error {
	err := q.client.XGroupCreateMkStream(ctx, q.stream, q.group, "0").Err()
	if err != nil && !strings.Contains(err.Error(), "BUSYGROUP") {
		return err
	}
	return nil
}

func (q *RedisFeedRefreshQueue) Read(ctx context.Context, consumerName string, count int, block time.Duration) ([]FeedRefreshJob, error) {
	if consumerName == "" {
		return nil, ErrFeedRefreshConsumerNameRequired
	}
	if count <= 0 {
		count = 10
	}

	streams, err := q.client.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group:    q.group,
		Consumer: consumerName,
		Streams:  []string{q.stream, ">"},
		Count:    int64(count),
		Block:    block,
	}).Result()
	if errors.Is(err, redis.Nil) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	jobs := make([]FeedRefreshJob, 0)
	for _, stream := range streams {
		for _, message := range stream.Messages {
			job, err := feedRefreshJobFromMessage(message)
			if err != nil {
				return nil, err
			}
			jobs = append(jobs, job)
		}
	}
	return jobs, nil
}

func (q *RedisFeedRefreshQueue) Ack(ctx context.Context, jobIDs ...string) error {
	if len(jobIDs) == 0 {
		return nil
	}
	return q.client.XAck(ctx, q.stream, q.group, jobIDs...).Err()
}

func feedRefreshJobFromMessage(message redis.XMessage) (FeedRefreshJob, error) {
	userID, ok := message.Values["user_id"].(string)
	if !ok || userID == "" {
		return FeedRefreshJob{}, fmt.Errorf("feed refresh job %s missing user_id", message.ID)
	}

	job := FeedRefreshJob{
		ID:     message.ID,
		UserID: userID,
	}
	if reason, ok := message.Values["reason"].(string); ok {
		job.Reason = reason
	}
	if rawRequestedAt, ok := message.Values["requested_at"].(string); ok && rawRequestedAt != "" {
		requestedAt, err := time.Parse(time.RFC3339Nano, rawRequestedAt)
		if err != nil {
			return FeedRefreshJob{}, fmt.Errorf("parse feed refresh job %s requested_at: %w", message.ID, err)
		}
		job.RequestedAt = requestedAt
	}

	return job, nil
}
