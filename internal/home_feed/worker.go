package home_feed

import (
	"context"
	"errors"
	"log/slog"
	"time"
)

type FeedWorker struct {
	queue        FeedRefreshJobQueue
	composer     FeedComposer
	snapshots    FeedSnapshotWriter
	consumerName string
	batchSize    int
	block        time.Duration
	logger       *slog.Logger
}

type FeedWorkerOptions struct {
	ConsumerName string
	BatchSize    int
	Block        time.Duration
	Logger       *slog.Logger
}

func NewFeedWorker(
	queue FeedRefreshJobQueue,
	composer FeedComposer,
	snapshots FeedSnapshotWriter,
	opts FeedWorkerOptions,
) *FeedWorker {
	if opts.BatchSize <= 0 {
		opts.BatchSize = 10
	}
	if opts.Block == 0 {
		opts.Block = 5 * time.Second
	}
	if opts.Logger == nil {
		opts.Logger = slog.Default()
	}

	return &FeedWorker{
		queue:        queue,
		composer:     composer,
		snapshots:    snapshots,
		consumerName: opts.ConsumerName,
		batchSize:    opts.BatchSize,
		block:        opts.Block,
		logger:       opts.Logger,
	}
}

func (w *FeedWorker) Run(ctx context.Context) error {
	if err := w.queue.EnsureConsumerGroup(ctx); err != nil {
		return err
	}

	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		if _, err := w.ProcessBatch(ctx); err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return err
			}
			w.logger.Error("failed to process home feed refresh batch", slog.String("error", err.Error()))
		}
	}
}

func (w *FeedWorker) ProcessBatch(ctx context.Context) (int, error) {
	jobs, err := w.queue.Read(ctx, w.consumerName, w.batchSize, w.block)
	if err != nil {
		return 0, err
	}

	processed := 0
	for _, job := range jobs {
		if err := w.processJob(ctx, job); err != nil {
			w.logger.Error("failed to refresh home feed",
				slog.String("job_id", job.ID),
				slog.String("user_id", job.UserID),
				slog.String("reason", job.Reason),
				slog.String("error", err.Error()),
			)
			continue
		}
		processed++
	}

	return processed, nil
}

func (w *FeedWorker) processJob(ctx context.Context, job FeedRefreshJob) error {
	now := time.Now().UTC()
	layout, err := w.composer.Compose(ctx, FeedRequest{
		UserID: job.UserID,
		Now:    now,
	})
	if err != nil {
		return err
	}

	if err := w.snapshots.Save(ctx, job.UserID, layout, now); err != nil {
		return err
	}

	return w.queue.Ack(ctx, job.ID)
}
