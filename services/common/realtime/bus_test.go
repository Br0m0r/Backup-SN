package realtime

import (
	"context"
	"testing"
	"time"
)

type transportContract struct{}

func (transportContract) Publish(context.Context, []byte) error         { return nil }
func (transportContract) Subscribe(context.Context, func([]byte)) error { return nil }
func (transportContract) MarkOnline(context.Context, int) error         { return nil }
func (transportContract) MarkOffline(context.Context, int) error        { return nil }
func (transportContract) IsOnline(context.Context, int) (bool, error)   { return false, nil }
func (transportContract) PresenceRefreshInterval() time.Duration        { return time.Second }

func TestTransportContractCanBeFaked(t *testing.T) {
	var transport Transport = transportContract{}
	if transport.PresenceRefreshInterval() != time.Second {
		t.Fatal("unexpected refresh interval")
	}
}
