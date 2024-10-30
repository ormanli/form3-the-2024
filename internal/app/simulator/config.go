package simulator

import (
	"time"
)

// Config defines configuration of application. Values are parsed from environment variables.
type Config struct {
	ServerPort                    int           `split_words:"true" default:"11111"`
	ServerHost                    string        `split_words:"true" default:"localhost"`
	ServerGracefulShutdownTimeout time.Duration `split_words:"true" default:"3s"`
	InitDebug                     bool          `split_words:"true"`
	DummyMinAmountToWait          int           `split_words:"true" default:"100"`
	DummyMaxAmountToWait          int           `split_words:"true" default:"10000"`
}
