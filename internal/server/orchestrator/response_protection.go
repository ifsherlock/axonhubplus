package orchestrator

import (
	"context"
	"errors"
	"fmt"

	"github.com/looplj/axonhub/internal/log"
	"github.com/looplj/axonhub/internal/server/biz"
	"github.com/looplj/axonhub/llm"
	"github.com/looplj/axonhub/llm/pipeline"
	"github.com/looplj/axonhub/llm/streams"
	"github.com/looplj/axonhub/llm/transformer"
)

const responseProtectionRejectedMessage = "response blocked by response protection policy"

func responseProtectionRejectionError() error {
	return fmt.Errorf("%w: %s", transformer.ErrInvalidRequest, responseProtectionRejectedMessage)
}

func responseProtectionFailoverError(ctx context.Context, err error) error {
	log.Error(ctx, "response protection rejected upstream response, triggering channel failover",
		log.Cause(err),
	)

	return pipeline.WrapUpstreamError(err)
}

func protectResponses(inbound *PersistentInboundTransformer) pipeline.Middleware {
	return pipeline.OnLlmResponse("protect-responses", func(ctx context.Context, llmResponse *llm.Response) (*llm.Response, error) {
		if inbound.state.ResponseProtecter == nil {
			return llmResponse, nil
		}

		protected, err := inbound.state.ResponseProtecter.Protect(ctx, llmResponse)
		if err != nil {
			if errors.Is(err, biz.ErrResponseProtectionFailover) {
				return nil, responseProtectionFailoverError(ctx, err)
			}

			if errors.Is(err, biz.ErrResponseProtectionRejected) {
				return nil, responseProtectionRejectionError()
			}

			log.Warn(ctx, "failed to protect response", log.Cause(err))

			return llmResponse, nil
		}

		if protected == nil {
			return llmResponse, nil
		}

		return protected, nil
	})
}

func protectResponseStream(inbound *PersistentInboundTransformer) pipeline.Middleware {
	return pipeline.OnLlmStream("protect-response-stream", func(ctx context.Context, stream streams.Stream[*llm.Response]) (streams.Stream[*llm.Response], error) {
		if inbound.state.ResponseProtecter == nil {
			return stream, nil
		}

		const maxPreReadEvents = 5
		buffered := make([]*llm.Response, 0, maxPreReadEvents)

		for range maxPreReadEvents {
			if !stream.Next() {
				break
			}

			event := stream.Current()
			protected, err := inbound.state.ResponseProtecter.Protect(ctx, event)
			if err != nil {
				stream.Close()

				if errors.Is(err, biz.ErrResponseProtectionFailover) {
					return nil, responseProtectionFailoverError(ctx, err)
				}

				if errors.Is(err, biz.ErrResponseProtectionRejected) {
					return nil, responseProtectionRejectionError()
				}

				log.Warn(ctx, "failed to protect response stream", log.Cause(err))

				return streams.PrependStream(stream, buffered...), nil
			}

			if protected == nil {
				protected = event
			}

			buffered = append(buffered, protected)
		}

		if err := stream.Err(); err != nil {
			stream.Close()
			return nil, err
		}

		rest := streams.MapErr(stream, func(event *llm.Response) (*llm.Response, error) {
			protected, err := inbound.state.ResponseProtecter.Protect(ctx, event)
			if err != nil {
				if errors.Is(err, biz.ErrResponseProtectionFailover) {
					return nil, responseProtectionFailoverError(ctx, err)
				}

				if errors.Is(err, biz.ErrResponseProtectionRejected) {
					return nil, responseProtectionRejectionError()
				}

				log.Warn(ctx, "failed to protect response stream event", log.Cause(err))
				return event, nil
			}

			if protected == nil {
				return event, nil
			}

			return protected, nil
		})

		return streams.PrependStream(rest, buffered...), nil
	})
}
