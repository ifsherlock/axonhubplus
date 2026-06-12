package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/looplj/axonhub/internal/ent/schema/schematype"
	"github.com/looplj/axonhub/internal/objects"
	"github.com/looplj/axonhub/internal/scopes"
)

type ResponseProtectionRule struct {
	ent.Schema
}

func (ResponseProtectionRule) Mixin() []ent.Mixin {
	return []ent.Mixin{
		TimeMixin{},
		schematype.SoftDeleteMixin{},
	}
}

func (ResponseProtectionRule) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("name", "deleted_at").
			StorageKey("response_protection_rules_by_name").
			Unique(),
	}
}

func (ResponseProtectionRule) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			Comment("Rule name").
			Annotations(entgql.OrderField("NAME")),
		field.String("description").
			Default("").
			Comment("Rule description"),
		field.String("pattern").
			Comment("Regex pattern to match response content"),
		field.Enum("status").
			Values("enabled", "disabled", "archived").
			Default("disabled").
			Annotations(entgql.Skip(entgql.SkipMutationCreateInput)),
		field.JSON("settings", &objects.ResponseProtectionSettings{}).
			Comment("Response protection rule settings"),
	}
}

func (ResponseProtectionRule) Edges() []ent.Edge {
	return nil
}

func (ResponseProtectionRule) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entgql.QueryField(),
		entgql.RelayConnection(),
		entgql.Mutations(entgql.MutationCreate(), entgql.MutationUpdate()),
	}
}

func (ResponseProtectionRule) Policy() ent.Policy {
	return scopes.Policy{
		Query: scopes.QueryPolicy{
			scopes.APIKeyScopeQueryRule(scopes.ScopeReadChannels),
			scopes.OwnerRule(),
			scopes.UserReadScopeRule(scopes.ScopeReadChannels),
		},
		Mutation: scopes.MutationPolicy{
			scopes.OwnerRule(),
			scopes.UserWriteScopeRule(scopes.ScopeWriteChannels),
		},
	}
}
