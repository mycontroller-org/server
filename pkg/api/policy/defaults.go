package policy

import (
	policyTY "github.com/mycontroller-org/server/v2/pkg/types/policy"
	settingsTY "github.com/mycontroller-org/server/v2/pkg/types/settings"
)

// BuiltInPolicies returns system policies that should always exist.
func BuiltInPolicies() []policyTY.Policy {
	allActions := []string{policyTY.ActionAll}
	allResources := []string{policyTY.ResourceAll}

	rwActions := []string{
		policyTY.ActionGet, policyTY.ActionList, policyTY.ActionCreate, policyTY.ActionUpdate,
		policyTY.ActionDelete, policyTY.ActionEnable, policyTY.ActionDisable, policyTY.ActionReload,
		policyTY.ActionAction,
	}
	// operational resources (no user/policy/settings/backup admin)
	rwResources := []string{
		policyTY.ResourceGateway, policyTY.ResourceNode, policyTY.ResourceSource, policyTY.ResourceField,
		policyTY.ResourceTask, policyTY.ResourceSchedule, policyTY.ResourceHandler, policyTY.ResourceDashboard,
		policyTY.ResourceFirmware, policyTY.ResourceForwardPayload, policyTY.ResourceDataRepository,
		policyTY.ResourceVirtualDevice, policyTY.ResourceVirtualAssistant, policyTY.ResourceServiceToken,
		policyTY.ResourceMetric, policyTY.ResourceAction, policyTY.ResourceStatus, policyTY.ResourceQuickID,
	}

	roActions := []string{policyTY.ActionGet, policyTY.ActionList}
	roResources := append([]string{}, rwResources...)

	// The console reads the system settings document (units, page size, ...) on every
	// page. Grant only that one settings key: backup locations may hold credentials
	// and the dynamic secrets key can reset the jwt secret.
	readSettings := policyTY.Statement{
		Effect:    policyTY.EffectAllow,
		Actions:   []string{policyTY.ActionGet},
		Resources: []string{policyTY.FormatSettingsResource(settingsTY.KeySystemSettings)},
	}

	return []policyTY.Policy{
		{
			ID:          policyTY.PolicyAdmin,
			Description: "Full access to all resources and actions",
			System:      true,
			Statements: []policyTY.Statement{
				{Effect: policyTY.EffectAllow, Actions: allActions, Resources: allResources},
			},
		},
		{
			ID:          policyTY.PolicyReadWrite,
			Description: "Read and write operational resources; no user/policy/settings/backup admin",
			System:      true,
			Statements: []policyTY.Statement{
				{Effect: policyTY.EffectAllow, Actions: rwActions, Resources: rwResources},
				readSettings,
			},
		},
		{
			ID:          policyTY.PolicyReadOnly,
			Description: "Read-only access to operational resources",
			System:      true,
			Statements: []policyTY.Statement{
				{Effect: policyTY.EffectAllow, Actions: roActions, Resources: roResources},
				readSettings,
			},
		},
	}
}
