package strategies

import (
	"context"
	"github.com/goletan/events-service/shared/types"
)

// Strategy defines the interface for sending events.
type Strategy interface {
	// Send sends an event to its destination.
	Send(ctx context.Context, event *types.Event) error
}
