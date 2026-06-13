package orchestrator

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/looplj/axonhub/internal/server/biz"
	"github.com/looplj/axonhub/llm"
	"github.com/looplj/axonhub/llm/streams"
)

type countingResponseProtecter struct {
	failAt int
	count  int
}

func (p *countingResponseProtecter) Protect(_ context.Context, resp *llm.Response) (*llm.Response, error) {
	p.count++
	if p.count == p.failAt {
		return nil, biz.ErrResponseProtectionFailover
	}

	return resp, nil
}

func TestProtectResponseStream_FailoverAfterPreReadRejectsStream(t *testing.T) {
	protecter := &countingResponseProtecter{failAt: 6}
	inbound := &PersistentInboundTransformer{
		state: &PersistenceState{ResponseProtecter: protecter},
	}

	events := []*llm.Response{
		{}, {}, {}, {}, {}, {},
	}
	middleware := protectResponseStream(inbound)
	protected, err := middleware.OnOutboundLlmStream(context.Background(), streams.SliceStream(events))
	require.NoError(t, err)

	for range 5 {
		require.True(t, protected.Next())
	}

	require.False(t, protected.Next())
	require.True(t, errors.Is(protected.Err(), biz.ErrResponseProtectionFailover))
}
