package quota

import (
	"context"
	"github.com/google/uuid"
	"github.com/skybi/data-server/internal/apikey"
	"github.com/skybi/data-server/internal/hashmap"
)

// Tracker keeps track of the used quota of API keys and updates it in batches in order to reduce database traffic
type Tracker struct {
	repo apikey.Repository

	usedQuotas hashmap.Map[uuid.UUID, int64]
}

// NewTracker creates a new API key quota tracker
func NewTracker(repo apikey.Repository) *Tracker {
	return &Tracker{
		repo:       repo,
		usedQuotas: hashmap.NewNormal[uuid.UUID, int64](),
	}
}

// Get returns the current used API quota of a specific API key
func (tracker *Tracker) Get(key *apikey.Key) int64 {
	current, ok := tracker.usedQuotas.Lookup(key.ID)
	if !ok {
		current = key.UsedQuota
	}
	return current
}

// Accumulate accumulates the used API quota of a specific API key by 1
func (tracker *Tracker) Accumulate(key *apikey.Key) {
	tracker.usedQuotas.Set(key.ID, tracker.Get(key)+1)
}

// Flush sends all changed API quota to the database and resets the counters
func (tracker *Tracker) Flush() (int, error) {
	amount := tracker.usedQuotas.Size()
	if amount == 0 {
		return 0, nil
	}

	var err error
	tracker.usedQuotas.BootstrappedManipulation(func(raw map[uuid.UUID]int64) {
		err = tracker.repo.UpdateManyQuotas(context.Background(), raw)
	})
	if err != nil {
		return 0, err
	}
	tracker.usedQuotas.Clear()
	return amount, nil
}
