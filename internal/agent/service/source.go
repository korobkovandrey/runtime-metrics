package service

import (
	"context"
	"sync"

	"github.com/korobkovandrey/runtime-metrics/internal/model"
	"golang.org/x/sync/errgroup"
)

// Source is a structure that provides a source of metrics
type Source struct {
	pollCount *model.Metric
	data      []*model.Metric
	mu        sync.RWMutex
}

// NewSource returns a new instance of Source
func NewSource() *Source {
	return &Source{
		pollCount: model.NewMetricCounter("PollCount", 0),
	}
}

// Collect collects metrics
func (s *Source) Collect(ctx context.Context) error {
	finalCh := make(chan *model.Metric)
	g := new(errgroup.Group)
	g.Go(func() error {
		for m := range genPullMetrics() {
			finalCh <- m
		}
		return nil
	})
	g.Go(func() error {
		for m := range genGopsutilMetrics(ctx) {
			if m.err != nil {
				return m.err
			}
			finalCh <- m.m
		}
		return nil
	})
	doneCh := make(chan struct{})
	go func() {
		defer close(doneCh)
		s.mu.RLock()
		data := make([]*model.Metric, 0, len(s.data))
		s.mu.RUnlock()
		for m := range finalCh {
			data = append(data, m)
		}
		s.mu.Lock()
		defer s.mu.Unlock()
		s.data = data
		*s.pollCount.Delta++
	}()
	err := g.Wait()
	close(finalCh)
	<-doneCh
	return err
}

// Get returns metrics
func (s *Source) Get() (data []*model.Metric, delta int64) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	data = make([]*model.Metric, len(s.data)+1)
	for i, m := range s.data {
		data[i] = m.Clone()
	}
	data[len(data)-1] = s.pollCount.Clone()
	delta = *s.pollCount.Delta
	return data, delta
}

// Commit commits metrics
func (s *Source) Commit(delta int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	*s.pollCount.Delta -= delta
}
