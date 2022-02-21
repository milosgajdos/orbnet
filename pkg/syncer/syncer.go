package syncer

import (
	"context"
)

// Syncer is used for syncing GitHub repos.
type Syncer interface {
	// Sync syncs GitHub repos read from the given channel.
	Sync(context.Context, <-chan interface{})
}
