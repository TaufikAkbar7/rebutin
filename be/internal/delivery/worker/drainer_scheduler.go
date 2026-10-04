package worker

import (
	"context"
	"sync"

	"github.com/sirupsen/logrus"
)

type DrainerScheduler struct {
	appCtx  context.Context
	drainer *Drainer
	mu      sync.Mutex
	cancels map[string]context.CancelFunc
	log     *logrus.Logger
}

func NewDrainerScheduler(appCtx context.Context, drainer *Drainer, log *logrus.Logger) *DrainerScheduler {
	return &DrainerScheduler{
		appCtx:  appCtx,
		drainer: drainer,
		log:     log,
		cancels: make(map[string]context.CancelFunc),
	}
}

func (s *DrainerScheduler) Start(runID string, max, batch int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.cancels[runID]; ok {
		return
	}
	ctx, cancel := context.WithCancel(s.appCtx)
	s.cancels[runID] = cancel
	go func() {
		defer s.Stop(runID)
		if err := s.drainer.Run(ctx, runID, max, batch); err != nil {
			s.log.Errorf("[Scheduler.Start] drainer run %s stopped: %v", runID, err)
		}
	}()
}

func (s *DrainerScheduler) Stop(runID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if cancel, ok := s.cancels[runID]; ok {
		cancel()
		delete(s.cancels, runID)
	}
}
