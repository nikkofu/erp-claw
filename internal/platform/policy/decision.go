package policy

// Decision represents the outcome of policy evaluation for a command.
type Decision string

const (
	DecisionAllow           Decision = "ALLOW"
	DecisionAllowWithGuard  Decision = "ALLOW_WITH_GUARD"
	DecisionRequireApproval Decision = "REQUIRE_APPROVAL"
	DecisionDeny            Decision = "DENY"
)
