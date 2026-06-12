package biz

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/samber/lo"
	"go.uber.org/fx"

	"github.com/looplj/axonhub/internal/authz"
	"github.com/looplj/axonhub/internal/ent"
	"github.com/looplj/axonhub/internal/ent/responseprotectionrule"
	"github.com/looplj/axonhub/internal/log"
	"github.com/looplj/axonhub/internal/objects"
	"github.com/looplj/axonhub/internal/pkg/watcher"
	"github.com/looplj/axonhub/internal/pkg/xcache"
	"github.com/looplj/axonhub/internal/pkg/xcache/live"
	"github.com/looplj/axonhub/internal/pkg/xerrors"
	"github.com/looplj/axonhub/llm"
)

type ResponseProtectionRuleServiceParams struct {
	fx.In

	CacheConfig xcache.Config
	Ent         *ent.Client
}

func NewResponseProtectionRuleService(params ResponseProtectionRuleServiceParams) *ResponseProtectionRuleService {
	svc := &ResponseProtectionRuleService{
		AbstractService: &AbstractService{
			db: params.Ent,
		},
	}

	cacheMode := params.CacheConfig.Mode
	if cacheMode == "" {
		cacheMode = xcache.ModeMemory
	}

	watcherMode := cacheMode
	if watcherMode == xcache.ModeTwoLevel {
		watcherMode = watcher.ModeRedis
	}

	notifier, err := watcher.NewWatcherFromConfig[live.CacheEvent[struct{}]](watcher.Config{
		Mode:  watcherMode,
		Redis: params.CacheConfig.Redis,
	}, watcher.WatcherFromConfigOptions{
		RedisChannel: "axonhub:cache:response_protection_rules",
		Buffer:       32,
	})
	if err != nil {
		panic(fmt.Errorf("response protection rule watcher init failed: %w", err))
	}

	svc.responseProtectionRuleNotifier = notifier
	svc.enabledRulesCache = live.NewCache(live.Options[[]*ent.ResponseProtectionRule]{
		Name:            "axonhub:enabled_response_protection_rules",
		InitialValue:    []*ent.ResponseProtectionRule{},
		RefreshInterval: 30 * time.Second,
		DebounceDelay:   500 * time.Millisecond,
		RefreshFunc:     svc.onEnabledRulesRefreshed,
		Watcher:         notifier,
	})

	if err := svc.enabledRulesCache.Load(context.Background(), true); err != nil {
		panic(fmt.Errorf("response protection rule cache initial load failed: %w", err))
	}

	return svc
}

type ResponseProtectionRuleService struct {
	*AbstractService

	enabledRulesCache              *live.Cache[[]*ent.ResponseProtectionRule]
	responseProtectionRuleNotifier watcher.Notifier[live.CacheEvent[struct{}]]
}

func (svc *ResponseProtectionRuleService) ValidateSettings(pattern string, settings *objects.ResponseProtectionSettings) error {
	if _, err := getOrCompilePromptProtectionPattern(pattern); err != nil {
		return fmt.Errorf("invalid regex pattern: %w", err)
	}

	if settings == nil {
		return fmt.Errorf("settings are required")
	}

	validActions := []objects.ResponseProtectionAction{
		objects.ResponseProtectionActionMask,
		objects.ResponseProtectionActionReject,
		objects.ResponseProtectionActionFailover,
	}
	if !slices.Contains(validActions, settings.Action) {
		return fmt.Errorf("invalid action: %s", settings.Action)
	}

	if settings.Action == objects.ResponseProtectionActionMask && settings.Replacement == "" {
		return fmt.Errorf("replacement is required for mask action")
	}

	validScopes := []objects.ResponseProtectionScope{objects.ResponseProtectionScopeText}
	for _, scope := range settings.Scopes {
		if !slices.Contains(validScopes, scope) {
			return fmt.Errorf("invalid scope: %s", scope)
		}
	}

	return nil
}

func (svc *ResponseProtectionRuleService) Protect(ctx context.Context, resp *llm.Response) (*llm.Response, error) {
	rules, err := svc.ListEnabledRules(ctx)
	if err != nil {
		log.Warn(ctx, "failed to load enabled response protection rules", log.Cause(err))
		return nil, err
	}

	return ProtectResponse(ctx, resp, rules)
}

func (svc *ResponseProtectionRuleService) Stop() {
	if svc.enabledRulesCache != nil {
		svc.enabledRulesCache.Stop()
	}
}

func (svc *ResponseProtectionRuleService) onEnabledRulesRefreshed(ctx context.Context, _ []*ent.ResponseProtectionRule, lastUpdate time.Time) ([]*ent.ResponseProtectionRule, time.Time, bool, error) {
	ctx = authz.WithSystemBypass(ctx, "response-protection-rule-cache")
	client := svc.entFromContext(ctx)

	rules, err := client.ResponseProtectionRule.Query().
		Where(responseprotectionrule.StatusEQ(responseprotectionrule.StatusEnabled)).
		Order(ent.Asc(responseprotectionrule.FieldID)).
		All(ctx)
	if err != nil {
		return nil, lastUpdate, false, err
	}

	newUpdateTime := lastUpdate
	if len(rules) > 0 {
		newUpdateTime = lo.MaxBy(rules, func(a, b *ent.ResponseProtectionRule) bool {
			return a.UpdatedAt.After(b.UpdatedAt)
		}).UpdatedAt
	}

	return rules, newUpdateTime, true, nil
}

func (svc *ResponseProtectionRuleService) asyncReloadEnabledRules() {
	if svc.responseProtectionRuleNotifier == nil {
		return
	}

	if err := svc.responseProtectionRuleNotifier.Notify(context.Background(), live.NewForceRefreshEvent[struct{}]()); err != nil {
		log.Warn(context.Background(), "response protection rule cache watcher notify failed", log.Cause(err))
	}
}

func (svc *ResponseProtectionRuleService) CreateRule(ctx context.Context, input ent.CreateResponseProtectionRuleInput) (*ent.ResponseProtectionRule, error) {
	if input.Settings == nil {
		return nil, fmt.Errorf("settings are required")
	}

	if err := svc.ValidateSettings(input.Pattern, input.Settings); err != nil {
		return nil, err
	}

	existing, err := svc.entFromContext(ctx).ResponseProtectionRule.Query().
		Where(responseprotectionrule.Name(input.Name)).
		First(ctx)
	if err != nil && !ent.IsNotFound(err) {
		return nil, fmt.Errorf("failed to check existing response protection rule: %w", err)
	}

	if existing != nil {
		return nil, xerrors.DuplicateNameError("response protection rule", input.Name)
	}

	rule, err := svc.entFromContext(ctx).ResponseProtectionRule.Create().
		SetInput(input).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create response protection rule: %w", err)
	}

	svc.asyncReloadEnabledRules()

	return rule, nil
}

func (svc *ResponseProtectionRuleService) UpdateRule(ctx context.Context, id int, input *ent.UpdateResponseProtectionRuleInput) (*ent.ResponseProtectionRule, error) {
	current, err := svc.entFromContext(ctx).ResponseProtectionRule.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to query response protection rule: %w", err)
	}

	pattern := lo.FromPtrOr(input.Pattern, current.Pattern)

	settings := current.Settings
	if input.Settings != nil {
		settings = input.Settings
	}

	if err := svc.ValidateSettings(pattern, settings); err != nil {
		return nil, err
	}

	rule, err := svc.entFromContext(ctx).ResponseProtectionRule.UpdateOneID(id).
		SetInput(*input).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to update response protection rule: %w", err)
	}

	svc.asyncReloadEnabledRules()

	return rule, nil
}

func (svc *ResponseProtectionRuleService) DeleteRule(ctx context.Context, id int) error {
	if err := svc.entFromContext(ctx).ResponseProtectionRule.DeleteOneID(id).Exec(ctx); err != nil {
		return fmt.Errorf("failed to delete response protection rule: %w", err)
	}

	svc.asyncReloadEnabledRules()

	return nil
}

func (svc *ResponseProtectionRuleService) UpdateRuleStatus(ctx context.Context, id int, status responseprotectionrule.Status) (*ent.ResponseProtectionRule, error) {
	rule, err := svc.entFromContext(ctx).ResponseProtectionRule.UpdateOneID(id).
		SetStatus(status).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to update response protection rule status: %w", err)
	}

	svc.asyncReloadEnabledRules()

	return rule, nil
}

func (svc *ResponseProtectionRuleService) BulkDeleteRules(ctx context.Context, ids []int) error {
	if len(ids) == 0 {
		return nil
	}

	if _, err := svc.entFromContext(ctx).ResponseProtectionRule.Delete().
		Where(responseprotectionrule.IDIn(ids...)).
		Exec(ctx); err != nil {
		return fmt.Errorf("failed to bulk delete response protection rules: %w", err)
	}

	svc.asyncReloadEnabledRules()

	return nil
}

func (svc *ResponseProtectionRuleService) BulkDisableRules(ctx context.Context, ids []int) error {
	if len(ids) == 0 {
		return nil
	}

	if _, err := svc.entFromContext(ctx).ResponseProtectionRule.Update().
		Where(responseprotectionrule.IDIn(ids...)).
		SetStatus(responseprotectionrule.StatusDisabled).
		Save(ctx); err != nil {
		return fmt.Errorf("failed to bulk disable response protection rules: %w", err)
	}

	svc.asyncReloadEnabledRules()

	return nil
}

func (svc *ResponseProtectionRuleService) BulkEnableRules(ctx context.Context, ids []int) error {
	if len(ids) == 0 {
		return nil
	}

	if _, err := svc.entFromContext(ctx).ResponseProtectionRule.Update().
		Where(responseprotectionrule.IDIn(ids...)).
		SetStatus(responseprotectionrule.StatusEnabled).
		Save(ctx); err != nil {
		return fmt.Errorf("failed to bulk enable response protection rules: %w", err)
	}

	svc.asyncReloadEnabledRules()

	return nil
}

func (svc *ResponseProtectionRuleService) ListEnabledRules(ctx context.Context) ([]*ent.ResponseProtectionRule, error) {
	if svc.enabledRulesCache != nil {
		return svc.enabledRulesCache.GetData(), nil
	}

	rules, err := svc.entFromContext(ctx).ResponseProtectionRule.Query().
		Where(responseprotectionrule.StatusEQ(responseprotectionrule.StatusEnabled)).
		Order(ent.Asc(responseprotectionrule.FieldID)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list enabled response protection rules: %w", err)
	}

	return rules, nil
}
