package signal

import (
	"os"
	"os/signal"
)

// Register registers signal handlers and returns
// a channel for controlling signal actions.
func Register(sig ...os.Signal) chan os.Signal {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, sig...)
	return sigChan
}
