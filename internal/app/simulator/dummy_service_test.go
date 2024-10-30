package simulator

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_DummyService(t *testing.T) {
	service := NewDummyService(Config{
		DummyMinAmountToWait: 5,
		DummyMaxAmountToWait: 10,
	})

	tests := []struct {
		name       string
		amount     int
		assertFunc func(*testing.T, time.Duration, error)
	}{
		{
			name:   "less than min amount",
			amount: 1,
			assertFunc: func(t *testing.T, duration time.Duration, err error) {
				assert.NoError(t, err)
				assert.InDelta(t, time.Millisecond, duration, float64(time.Millisecond))
			},
		},
		{
			name:   "more than min amount",
			amount: 6,
			assertFunc: func(t *testing.T, duration time.Duration, err error) {
				assert.NoError(t, err)
				assert.InDelta(t, 6*time.Millisecond, duration, float64(time.Millisecond))
			},
		},
		{
			name:   "more than max amount",
			amount: 10000,
			assertFunc: func(t *testing.T, duration time.Duration, err error) {
				assert.NoError(t, err)
				assert.InDelta(t, 10*time.Millisecond, duration, float64(2*time.Millisecond))
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			now := time.Now()
			err := service.Process(test.amount)
			duration := time.Since(now)
			test.assertFunc(t, duration, err)
		})
	}
}
