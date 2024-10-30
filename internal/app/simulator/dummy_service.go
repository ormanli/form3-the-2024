package simulator

import (
	"time"
)

type DummyService struct {
	cfg Config
}

func NewDummyService(cfg Config) *DummyService {
	return &DummyService{cfg: cfg}
}

func (d *DummyService) Process(amount int) error {
	if amount > d.cfg.DummyMinAmountToWait {
		if amount > d.cfg.DummyMaxAmountToWait {
			amount = d.cfg.DummyMaxAmountToWait
		}
		time.Sleep(time.Duration(amount) * time.Millisecond)
	}

	return nil
}
