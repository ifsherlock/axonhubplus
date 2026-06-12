package biz

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/looplj/axonhub/internal/ent"
	"github.com/looplj/axonhub/internal/objects"
	"github.com/looplj/axonhub/llm"
)

func TestProtectResponseFailoverOnSpam(t *testing.T) {
	rules := []*ent.ResponseProtectionRule{
		{
			Name:    "block spam domain",
			Pattern: `(?i)(dc\.hhhl\.cc|t\.me/UniverseFederation)`,
			Settings: &objects.ResponseProtectionSettings{
				Action: objects.ResponseProtectionActionFailover,
			},
		},
	}

	text := "欢迎加入 https://dc.hhhl.cc/chat/room/amlc1bekzi"
	resp := &llm.Response{Choices: []llm.Choice{{Message: &llm.Message{Content: llm.MessageContent{Content: &text}}}}}

	_, err := ProtectResponse(context.Background(), resp, rules)
	require.ErrorIs(t, err, ErrResponseProtectionFailover)
}

func TestProtectResponseFailoverOnRequestBlocked(t *testing.T) {
	rules := []*ent.ResponseProtectionRule{
		{
			Name:    "block upstream refusal",
			Pattern: `(?i)\bREQUEST_BLOCKED\b`,
			Settings: &objects.ResponseProtectionSettings{
				Action: objects.ResponseProtectionActionFailover,
			},
		},
	}

	text := "REQUEST_BLOCKED\nCategory: REVERSE_ENGINEERING"
	resp := &llm.Response{Choices: []llm.Choice{{Message: &llm.Message{Content: llm.MessageContent{Content: &text}}}}}

	_, err := ProtectResponse(context.Background(), resp, rules)
	require.True(t, errors.Is(err, ErrResponseProtectionFailover))
}

func TestProtectResponseMasksTextParts(t *testing.T) {
	rules := []*ent.ResponseProtectionRule{
		{
			Name:    "mask domain",
			Pattern: `(?i)dc\.hhhl\.cc`,
			Settings: &objects.ResponseProtectionSettings{
				Action:      objects.ResponseProtectionActionMask,
				Replacement: "[filtered]",
			},
		},
	}

	partText := "visit dc.hhhl.cc now"
	resp := &llm.Response{Choices: []llm.Choice{{Delta: &llm.Message{Content: llm.MessageContent{MultipleContent: []llm.MessageContentPart{{Type: "text", Text: &partText}}}}}}}

	protected, err := ProtectResponse(context.Background(), resp, rules)
	require.NoError(t, err)
	require.Equal(t, "visit [filtered] now", *protected.Choices[0].Delta.Content.MultipleContent[0].Text)
}
