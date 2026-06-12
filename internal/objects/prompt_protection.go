package objects

type PromptProtectionAction string

const (
	PromptProtectionActionMask   PromptProtectionAction = "mask"
	PromptProtectionActionReject PromptProtectionAction = "reject"
)

type PromptProtectionScope string

const (
	PromptProtectionScopeSystem    PromptProtectionScope = "system"
	PromptProtectionScopeDeveloper PromptProtectionScope = "developer"
	PromptProtectionScopeUser      PromptProtectionScope = "user"
	PromptProtectionScopeAssistant PromptProtectionScope = "assistant"
	PromptProtectionScopeTool      PromptProtectionScope = "tool"
)

type PromptProtectionSettings struct {
	Action      PromptProtectionAction  `json:"action"`
	Replacement string                  `json:"replacement,omitempty"`
	Scopes      []PromptProtectionScope `json:"scopes,omitempty"`
}

type ResponseProtectionAction string

const (
	ResponseProtectionActionMask     ResponseProtectionAction = "mask"
	ResponseProtectionActionReject   ResponseProtectionAction = "reject"
	ResponseProtectionActionFailover ResponseProtectionAction = "failover"
)

type ResponseProtectionScope string

const (
	ResponseProtectionScopeText ResponseProtectionScope = "text"
)

type ResponseProtectionSettings struct {
	Action      ResponseProtectionAction  `json:"action"`
	Replacement string                    `json:"replacement,omitempty"`
	Scopes      []ResponseProtectionScope `json:"scopes,omitempty"`
}
