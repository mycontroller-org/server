package policy

import (
	types "github.com/mycontroller-org/server/v2/pkg/types"
	dashboardTY "github.com/mycontroller-org/server/v2/pkg/types/dashboard"
	dataRepoTY "github.com/mycontroller-org/server/v2/pkg/types/data_repository"
	fieldTY "github.com/mycontroller-org/server/v2/pkg/types/field"
	firmwareTY "github.com/mycontroller-org/server/v2/pkg/types/firmware"
	fwdPayloadTY "github.com/mycontroller-org/server/v2/pkg/types/forward_payload"
	nodeTY "github.com/mycontroller-org/server/v2/pkg/types/node"
	policyTY "github.com/mycontroller-org/server/v2/pkg/types/policy"
	schedulerTY "github.com/mycontroller-org/server/v2/pkg/types/scheduler"
	sourceTY "github.com/mycontroller-org/server/v2/pkg/types/source"
	taskTY "github.com/mycontroller-org/server/v2/pkg/types/task"
	vdTY "github.com/mycontroller-org/server/v2/pkg/types/virtual_device"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"
	gatewayTY "github.com/mycontroller-org/server/v2/plugin/gateway/types"
	handlerTY "github.com/mycontroller-org/server/v2/plugin/handler/types"
	vaTY "github.com/mycontroller-org/server/v2/plugin/virtual_assistant/types"
)

// ResolveResource rewrites access.Resource when path id is a storage UUID (or id)
// so policy match uses business names (e.g. gatewayId.nodeId) instead of UUID.
func (a *API) ResolveResource(access *RequestAccess) {
	if access == nil || access.Name == "" || access.Kind == "" {
		return
	}
	biz, err := a.ResolveBusinessName(access.Kind, access.Name)
	if err != nil || biz == "" {
		return
	}
	access.Name = biz
	access.Resource = FormatResource(access.Kind, biz)
}

// ResolveBusinessName loads the entity by storage id and returns the policy resource name.
func (a *API) ResolveBusinessName(kind, storageID string) (string, error) {
	if storageID == "" || a.storage == nil {
		return "", nil
	}
	filters := []storageTY.Filter{{Key: types.KeyID, Value: storageID}}

	switch kind {
	case policyTY.ResourceGateway:
		var e gatewayTY.Config
		if err := a.storage.FindOne(types.EntityGateway, &e, filters); err != nil {
			return "", err
		}
		return BusinessName(kind, e), nil
	case policyTY.ResourceNode:
		var e nodeTY.Node
		if err := a.storage.FindOne(types.EntityNode, &e, filters); err != nil {
			return "", err
		}
		return BusinessName(kind, e), nil
	case policyTY.ResourceSource:
		var e sourceTY.Source
		if err := a.storage.FindOne(types.EntitySource, &e, filters); err != nil {
			return "", err
		}
		return BusinessName(kind, e), nil
	case policyTY.ResourceField:
		var e fieldTY.Field
		if err := a.storage.FindOne(types.EntityField, &e, filters); err != nil {
			return "", err
		}
		return BusinessName(kind, e), nil
	case policyTY.ResourceTask:
		var e taskTY.Config
		if err := a.storage.FindOne(types.EntityTask, &e, filters); err != nil {
			return "", err
		}
		return BusinessName(kind, e), nil
	case policyTY.ResourceSchedule:
		var e schedulerTY.Config
		if err := a.storage.FindOne(types.EntitySchedule, &e, filters); err != nil {
			return "", err
		}
		return BusinessName(kind, e), nil
	case policyTY.ResourceHandler:
		var e handlerTY.Config
		if err := a.storage.FindOne(types.EntityHandler, &e, filters); err != nil {
			return "", err
		}
		return BusinessName(kind, e), nil
	case policyTY.ResourceDashboard:
		var e dashboardTY.Config
		if err := a.storage.FindOne(types.EntityDashboard, &e, filters); err != nil {
			return "", err
		}
		return BusinessName(kind, e), nil
	case policyTY.ResourceFirmware:
		var e firmwareTY.Firmware
		if err := a.storage.FindOne(types.EntityFirmware, &e, filters); err != nil {
			return "", err
		}
		return BusinessName(kind, e), nil
	case policyTY.ResourceForwardPayload:
		var e fwdPayloadTY.Config
		if err := a.storage.FindOne(types.EntityForwardPayload, &e, filters); err != nil {
			return "", err
		}
		return BusinessName(kind, e), nil
	case policyTY.ResourceDataRepository:
		var e dataRepoTY.Config
		if err := a.storage.FindOne(types.EntityDataRepository, &e, filters); err != nil {
			return "", err
		}
		return BusinessName(kind, e), nil
	case policyTY.ResourceVirtualDevice:
		var e vdTY.VirtualDevice
		if err := a.storage.FindOne(types.EntityVirtualDevice, &e, filters); err != nil {
			return "", err
		}
		return BusinessName(kind, e), nil
	case policyTY.ResourceVirtualAssistant:
		var e vaTY.Config
		if err := a.storage.FindOne(types.EntityVirtualAssistant, &e, filters); err != nil {
			return "", err
		}
		return BusinessName(kind, e), nil
	default:
		// unknown kind: keep path id as resource name
		return storageID, nil
	}
}
