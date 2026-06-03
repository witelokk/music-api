package home_feed

import (
	"context"
	"errors"
	"testing"
	"time"
)

type fakeFeedRefreshJobQueue struct {
	jobs       []FeedRefreshJob
	ackedIDs   []string
	ensureCall int
	readCall   int
	ackErr     error
}

func (q *fakeFeedRefreshJobQueue) Enqueue(ctx context.Context, userID, reason string) error {
	return nil
}

func (q *fakeFeedRefreshJobQueue) EnsureConsumerGroup(ctx context.Context) error {
	q.ensureCall++
	return nil
}

func (q *fakeFeedRefreshJobQueue) Read(ctx context.Context, consumerName string, count int, block time.Duration) ([]FeedRefreshJob, error) {
	q.readCall++
	return q.jobs, nil
}

func (q *fakeFeedRefreshJobQueue) Ack(ctx context.Context, jobIDs ...string) error {
	q.ackedIDs = append(q.ackedIDs, jobIDs...)
	return q.ackErr
}

type fakeFeedComposer struct {
	layout *Layout
	err    error
	calls  int
	reqs   []FeedRequest
}

func (c *fakeFeedComposer) Compose(ctx context.Context, req FeedRequest) (*Layout, error) {
	c.calls++
	c.reqs = append(c.reqs, req)
	return c.layout, c.err
}

type fakeFeedSnapshotRepository struct {
	calls int
	err   error
	saves []savedFeedSnapshot
}

type savedFeedSnapshot struct {
	userID      string
	layout      *Layout
	generatedAt time.Time
}

func (r *fakeFeedSnapshotRepository) Save(ctx context.Context, userID string, layout *Layout, generatedAt time.Time) error {
	r.calls++
	r.saves = append(r.saves, savedFeedSnapshot{
		userID:      userID,
		layout:      layout,
		generatedAt: generatedAt,
	})
	return r.err
}

func TestFeedWorker_ProcessBatch_ComposesSavesAndAcksJobs(t *testing.T) {
	queue := &fakeFeedRefreshJobQueue{
		jobs: []FeedRefreshJob{
			{ID: "1-0", UserID: "user-1", Reason: "user_event"},
		},
	}
	layout := &Layout{}
	composer := &fakeFeedComposer{layout: layout}
	snapshots := &fakeFeedSnapshotRepository{}
	worker := NewFeedWorker(queue, composer, snapshots, FeedWorkerOptions{
		ConsumerName: "worker-1",
	})

	processed, err := worker.ProcessBatch(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if processed != 1 {
		t.Fatalf("expected 1 processed job, got %d", processed)
	}
	if composer.calls != 1 {
		t.Fatalf("expected composer to be called once, got %d", composer.calls)
	}
	if composer.reqs[0].UserID != "user-1" {
		t.Fatalf("expected composer user user-1, got %q", composer.reqs[0].UserID)
	}
	if snapshots.calls != 1 {
		t.Fatalf("expected snapshot save once, got %d", snapshots.calls)
	}
	if snapshots.saves[0].layout != layout {
		t.Fatalf("expected saved layout to be composer layout")
	}
	if len(queue.ackedIDs) != 1 || queue.ackedIDs[0] != "1-0" {
		t.Fatalf("expected ack for job 1-0, got %+v", queue.ackedIDs)
	}
}

func TestFeedWorker_ProcessBatch_DoesNotAckFailedCompose(t *testing.T) {
	queue := &fakeFeedRefreshJobQueue{
		jobs: []FeedRefreshJob{
			{ID: "1-0", UserID: "user-1"},
		},
	}
	composer := &fakeFeedComposer{err: errors.New("compose failed")}
	snapshots := &fakeFeedSnapshotRepository{}
	worker := NewFeedWorker(queue, composer, snapshots, FeedWorkerOptions{
		ConsumerName: "worker-1",
	})

	processed, err := worker.ProcessBatch(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if processed != 0 {
		t.Fatalf("expected 0 processed jobs, got %d", processed)
	}
	if snapshots.calls != 0 {
		t.Fatalf("expected no snapshot save, got %d", snapshots.calls)
	}
	if len(queue.ackedIDs) != 0 {
		t.Fatalf("expected no ack, got %+v", queue.ackedIDs)
	}
}
