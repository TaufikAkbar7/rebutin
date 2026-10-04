package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"rebutin/internal/domain"
	"time"

	"github.com/sirupsen/logrus"
)

type Drainer struct {
	repo     domain.SessionCacheRepository
	interval time.Duration
	log      *logrus.Logger
}

func NewDrainer(repo domain.SessionCacheRepository, interval time.Duration, log *logrus.Logger) *Drainer {
	return &Drainer{repo: repo, interval: interval, log: log}
}

func (r *Drainer) Run(ctx context.Context, runID string, max, batch int) error {
	if batch <= 0 {
		return fmt.Errorf("batch must be > 0, got %d", batch)
	}

	if batch > max {
		batch = max
	}

	t := time.NewTicker(r.interval)
	defer t.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-t.C:
			admitted, err := r.repo.Drain(ctx, runID, max, batch)
			if err != nil {
				// graceful shutdown
				if ctx.Err() != nil {
					return nil
				}
				r.log.Errorf("[Drainer.Run] Failed drain run %s: %v", runID, err)
				// next tick
				continue
			}
			if len(admitted) > 0 {
				valuesJSON, _ := json.MarshalIndent(admitted, "", "  ")
				r.log.Infof("[Drainer.Run] run %s | participant left the waiting room %s\n", runID, valuesJSON)
			}
		}
	}
}
