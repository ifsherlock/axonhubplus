package biz

import (
	"context"
	"fmt"

	"github.com/looplj/axonhub/internal/objects"
)

type ResponseProtectionPreviewInput struct {
	Pattern  string
	TestText string
	Settings *objects.ResponseProtectionSettings
}

type ResponseProtectionPreviewResult struct {
	Result   string
	HasMatch bool
}

func (svc *ResponseProtectionRuleService) Preview(ctx context.Context, input ResponseProtectionPreviewInput) (*ResponseProtectionPreviewResult, error) {
	if err := svc.ValidateSettings(input.Pattern, input.Settings); err != nil {
		return nil, err
	}

	re, err := getOrCompilePromptProtectionPattern(input.Pattern)
	if err != nil {
		return nil, fmt.Errorf("invalid regex pattern: %w", err)
	}

	hasMatch, err := re.MatchString(input.TestText)
	if err != nil {
		return nil, err
	}

	result := input.TestText
	if hasMatch {
		switch input.Settings.Action {
		case objects.ResponseProtectionActionMask:
			result, err = re.Replace(input.TestText, input.Settings.Replacement, -1, -1)
			if err != nil {
				return nil, err
			}
		case objects.ResponseProtectionActionReject:
			result = string(objects.ResponseProtectionActionReject)
		case objects.ResponseProtectionActionFailover:
			result = string(objects.ResponseProtectionActionFailover)
		}
	}

	return &ResponseProtectionPreviewResult{
		Result:   result,
		HasMatch: hasMatch,
	}, nil
}
