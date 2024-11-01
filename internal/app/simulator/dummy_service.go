package simulator

import (
	"time"
)

// DummyService is a service that processes amounts with configurable delays.
type DummyService struct {
	cfg Config
}

// NewDummyService creates a new instance of DummyService with the given configuration.
func NewDummyService(cfg Config) *DummyService {
	return &DummyService{cfg: cfg}
}

// Process processes the amount with configurable delays based on the service's configuration.
// If the amount is greater than DummyMinAmountToWait, it will sleep for the specified duration.
// If the amount exceeds DummyMaxAmountToWait, it will cap the delay at DummyMaxAmountToWait.
func (d *DummyService) Process(amount int) error {
	if amount > d.cfg.DummyMinAmountToWait {
		if amount > d.cfg.DummyMaxAmountToWait {
			amount = d.cfg.DummyMaxAmountToWait
		}
		time.Sleep(time.Duration(amount) * time.Millisecond)
	}

	return nil
}
