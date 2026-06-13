package biz

import (
	"context"
	"errors"
	"strings"

	"github.com/looplj/axonhub/internal/ent"
	"github.com/looplj/axonhub/internal/log"
	"github.com/looplj/axonhub/internal/objects"
	"github.com/looplj/axonhub/llm"
)

var (
	ErrResponseProtectionRejected = errors.New("response blocked by response protection policy")
	ErrResponseProtectionFailover = errors.New("response rejected, trigger channel failover")
)

func IsResponseProtectionFailover(err error) bool {
	return errors.Is(err, ErrResponseProtectionFailover)
}

type ResponseProtectionResult struct {
	Response     *llm.Response
	MatchedRules []*ent.ResponseProtectionRule
	Rejected     bool
	Failover     bool
}

// ApplyResponseProtectionRules applies response protection rules to a unified LLM response.
func ApplyResponseProtectionRules(resp *llm.Response, rules []*ent.ResponseProtectionRule) ResponseProtectionResult {
	if resp == nil || len(rules) == 0 {
		return ResponseProtectionResult{Response: resp}
	}

	var matchedRules []*ent.ResponseProtectionRule

	for _, rule := range rules {
		if rule.Settings == nil || !responseProtectionRuleAppliesToText(rule.Settings.Scopes) {
			continue
		}

		updatedResp, matched := applyResponseProtectionRuleToResponse(resp, rule)
		if !matched {
			continue
		}

		if rule.Settings.Action == objects.ResponseProtectionActionReject {
			return ResponseProtectionResult{
				Response:     updatedResp,
				MatchedRules: []*ent.ResponseProtectionRule{rule},
				Rejected:     true,
			}
		}

		if rule.Settings.Action == objects.ResponseProtectionActionFailover {
			return ResponseProtectionResult{
				Response:     updatedResp,
				MatchedRules: []*ent.ResponseProtectionRule{rule},
				Rejected:     true,
				Failover:     true,
			}
		}

		resp = updatedResp
		matchedRules = append(matchedRules, rule)
	}

	return ResponseProtectionResult{
		Response:     resp,
		MatchedRules: matchedRules,
	}
}

func ProtectResponse(ctx context.Context, resp *llm.Response, rules []*ent.ResponseProtectionRule) (*llm.Response, error) {
	if len(rules) == 0 {
		if log.DebugEnabled(ctx) {
			log.Debug(ctx, "no enabled response protection rules")
		}

		return resp, nil
	}

	result := ApplyResponseProtectionRules(resp, rules)
	if len(result.MatchedRules) == 0 {
		if log.DebugEnabled(ctx) {
			log.Debug(ctx, "response protection passed without rule match", log.Int("rule_count", len(rules)))
		}

		return resp, nil
	}

	if result.Rejected {
		log.Warn(ctx, "response protection rejected response",
			log.String("rule_name", result.MatchedRules[0].Name),
			log.Bool("failover", result.Failover),
		)

		if result.Failover {
			return result.Response, ErrResponseProtectionFailover
		}

		return result.Response, ErrResponseProtectionRejected
	}

	if log.DebugEnabled(ctx) {
		log.Debug(ctx, "response protection masked response", log.Any("rules", result.MatchedRules))
	}

	return result.Response, nil
}

func applyResponseProtectionRuleToResponse(resp *llm.Response, rule *ent.ResponseProtectionRule) (*llm.Response, bool) {
	matched := false

	for i, choice := range resp.Choices {
		if choice.Message != nil {
			updated, msgMatched := applyResponseProtectionRuleToMessage(*choice.Message, rule)
			if msgMatched {
				resp.Choices[i].Message = &updated
				matched = true
			}
		}

		if choice.Delta != nil {
			updated, msgMatched := applyResponseProtectionRuleToMessage(*choice.Delta, rule)
			if msgMatched {
				resp.Choices[i].Delta = &updated
				matched = true
			}
		}
	}

	if resp.Completion != nil {
		for i, choice := range resp.Completion.Choices {
			if choice.Text == "" || !MatchPromptProtectionRule(rule.Pattern, choice.Text) {
				continue
			}

			if rule.Settings.Action == objects.ResponseProtectionActionMask {
				resp.Completion.Choices[i].Text = ReplacePromptProtectionRule(rule.Pattern, choice.Text, rule.Settings.Replacement)
			}

			matched = true
		}
	}

	if resp.Error != nil && resp.Error.Detail.Message != "" && MatchPromptProtectionRule(rule.Pattern, resp.Error.Detail.Message) {
		if rule.Settings.Action == objects.ResponseProtectionActionMask {
			resp.Error.Detail.Message = ReplacePromptProtectionRule(rule.Pattern, resp.Error.Detail.Message, rule.Settings.Replacement)
		}

		matched = true
	}

	return resp, matched
}

func applyResponseProtectionRuleToMessage(msg llm.Message, rule *ent.ResponseProtectionRule) (llm.Message, bool) {
	matched := false

	if msg.Content.Content != nil && *msg.Content.Content != "" && MatchPromptProtectionRule(rule.Pattern, *msg.Content.Content) {
		if rule.Settings.Action == objects.ResponseProtectionActionMask {
			masked := ReplacePromptProtectionRule(rule.Pattern, *msg.Content.Content, rule.Settings.Replacement)
			msg.Content = llm.MessageContent{Content: &masked}
		}

		matched = true
	}

	for i, part := range msg.Content.MultipleContent {
		if !strings.EqualFold(part.Type, "text") || part.Text == nil || *part.Text == "" {
			continue
		}

		if !MatchPromptProtectionRule(rule.Pattern, *part.Text) {
			continue
		}

		if rule.Settings.Action == objects.ResponseProtectionActionMask {
			masked := ReplacePromptProtectionRule(rule.Pattern, *part.Text, rule.Settings.Replacement)
			msg.Content.MultipleContent[i].Text = &masked
		}

		matched = true
	}

	return msg, matched
}

func responseProtectionRuleAppliesToText(scopes []objects.ResponseProtectionScope) bool {
	if len(scopes) == 0 {
		return true
	}

	for _, scope := range scopes {
		if scope == objects.ResponseProtectionScopeText {
			return true
		}
	}

	return false
}
