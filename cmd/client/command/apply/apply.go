package apply

import (
	"errors"
	"fmt"
	"io"

	"github.com/mycontroller-org/server/v2/pkg/json"
	dataRepoTY "github.com/mycontroller-org/server/v2/pkg/types/data_repository"
	fieldTY "github.com/mycontroller-org/server/v2/pkg/types/field"
	firmwareTY "github.com/mycontroller-org/server/v2/pkg/types/firmware"
	nodeTY "github.com/mycontroller-org/server/v2/pkg/types/node"
	sourceTY "github.com/mycontroller-org/server/v2/pkg/types/source"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	gwTY "github.com/mycontroller-org/server/v2/plugin/gateway/types"
	"github.com/olekukonko/tablewriter"
)

// ErrApplyFailed is returned after the result table is printed when any row failed.
var ErrApplyFailed = errors.New("apply failed")

const (
	actionAdd          = "add"
	actionMerge        = "merge"
	actionDelete       = "delete"
	actionReplace      = "replace"
	actionNotAvailable = "not available"
)

// ResourceClient looks up and mutates nodes, sources, and fields.
type ResourceClient interface {
	FindGateway(id string) (idFound string, err error)
	FindNode(id, gatewayID, nodeID string) (idFound string, err error)
	FindSource(id, gatewayID, nodeID, sourceID string) (idFound string, err error)
	FindField(id, gatewayID, nodeID, sourceID, fieldID string) (idFound string, err error)
	FindFirmware(id string) (idFound string, err error)
	FindDataRepository(id string) (idFound string, err error)
	SaveGateway(resource Resource) error
	SaveNode(resource Resource) error
	SaveSource(resource Resource) error
	SaveField(resource Resource) error
	SaveFirmware(resource Resource) error
	SaveDataRepository(resource Resource) error
	DeleteGateway(ids ...string) error
	DeleteNode(ids ...string) error
	DeleteSource(ids ...string) error
	DeleteField(ids ...string) error
	DeleteFirmware(ids ...string) error
	DeleteDataRepository(ids ...string) error
	GetExisting(resource Resource) ([]byte, error)
}

type plannedAction struct {
	Resource   Resource
	Action     string
	ExistingID string
	Err        error
}

type applyRow struct {
	Resource string
	Action   string
	Status   string
}

// Apply plans and optionally executes resource operations.
// When dryRun is true, the server is still queried so add-already-exists can be verified.
func Apply(client ResourceClient, resources []Resource, replace, dryRun bool, out, errOut io.Writer) error {
	_ = errOut
	if len(resources) == 0 {
		return fmt.Errorf("no resources to apply")
	}

	plans := make([]plannedAction, 0, len(resources))
	planFailed := false
	pending := newPendingParents()
	for _, resource := range resources {
		plan := planResource(client, resource, replace, pending)
		plans = append(plans, plan)
		if plan.Err != nil {
			planFailed = true
			continue
		}
		switch plan.Action {
		case actionAdd, actionMerge, actionReplace:
			pending.remember(plan.Resource, true)
		case actionDelete:
			pending.remember(plan.Resource, false)
		}
	}

	rows := make([]applyRow, 0, len(plans))
	var applyErr error
	executed := newPendingParents()
	for _, plan := range plans {
		var execErr error
		if plan.Err == nil && plan.Action != actionNotAvailable && !dryRun {
			if err := checkExecutedParent(plan.Resource, executed); err != nil {
				execErr = err
			} else if err := executePlan(client, plan); err != nil {
				execErr = err
			}
			switch plan.Action {
			case actionAdd, actionMerge, actionReplace:
				executed.remember(plan.Resource, execErr == nil)
			case actionDelete:
				executed.remember(plan.Resource, false)
			}
		}
		if execErr != nil {
			applyErr = execErr
		}
		rows = append(rows, applyRow{
			Resource: plan.Resource.TableResource(),
			Action:   tableAction(plan),
			Status:   tableStatus(plan, dryRun, execErr),
		})
	}

	printApplyTable(out, rows)

	if planFailed || applyErr != nil {
		return ErrApplyFailed
	}
	return nil
}

func tableAction(plan plannedAction) string {
	if plan.Resource.Operation != "" {
		return plan.Resource.Operation
	}
	if plan.Action == actionNotAvailable {
		return actionDelete
	}
	if plan.Action == actionReplace {
		return actionAdd
	}
	if plan.Action != "" {
		return plan.Action
	}
	return "-"
}

func tableStatus(plan plannedAction, dryRun bool, execErr error) string {
	if plan.Err != nil {
		return "failed: " + plan.Err.Error()
	}
	if plan.Action == actionNotAvailable {
		return actionNotAvailable
	}
	if dryRun {
		return "dry-run"
	}
	if execErr != nil {
		return "failed: " + execErr.Error()
	}
	if plan.Action == actionReplace {
		return "replaced"
	}
	return "ok"
}

func printApplyTable(out io.Writer, rows []applyRow) {
	table := tablewriter.NewWriter(out)
	table.SetHeader([]string{"RESOURCE", "ACTION", "STATUS"})
	table.SetAutoFormatHeaders(false)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetAutoWrapText(false)
	table.SetBorder(false)
	table.SetHeaderLine(false)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("  ")
	table.SetRowSeparator("")
	table.SetTablePadding("  ")
	table.SetNoWhiteSpace(true)
	for _, row := range rows {
		table.Append([]string{row.Resource, row.Action, row.Status})
	}
	table.Render()
}

func planResource(client ResourceClient, resource Resource, replace bool, pending *pendingParents) plannedAction {
	existingID, err := findExisting(client, resource)
	if err != nil {
		return plannedAction{Resource: resource, Err: fmt.Errorf("lookup failed: %w", err)}
	}

	plan := plannedAction{Resource: resource, ExistingID: existingID, Action: resource.Operation}
	switch resource.Operation {
	case OperationAdd:
		if existingID != "" {
			if !replace && !resource.Replace {
				plan.Err = fmt.Errorf("already exists")
				return plan
			}
			// delete then add with the same id so existing references stay valid
			resource.SetID(existingID)
			plan.Resource = resource
			plan.Action = actionReplace
		} else {
			assignSaveID(&resource)
			plan.Resource = resource
			plan.Action = actionAdd
		}
	case OperationMerge:
		if existingID == "" {
			plan.Err = fmt.Errorf("not found")
			return plan
		}
		resource.SetID(existingID)
		existingJSON, err := client.GetExisting(resource)
		if err != nil {
			plan.Err = fmt.Errorf("lookup failed: %w", err)
			return plan
		}
		if err := mergeExisting(&resource, existingJSON); err != nil {
			plan.Err = fmt.Errorf("merge failed: %w", err)
			return plan
		}
		plan.Resource = resource
		plan.Action = actionMerge
	case OperationDelete:
		if existingID == "" {
			plan.Action = actionNotAvailable
			return plan
		}
		resource.SetID(existingID)
		plan.Resource = resource
		plan.Action = actionDelete
	default:
		plan.Err = fmt.Errorf("unsupported operation %q", resource.Operation)
		return plan
	}

	if plan.Err == nil && plan.Action != actionDelete && plan.Action != actionNotAvailable {
		if err := checkParent(client, plan.Resource, pending); err != nil {
			plan.Err = err
		}
	}
	return plan
}

type pendingParents struct {
	gateways map[string]bool
	nodes    map[string]bool
	sources  map[string]bool
}

func newPendingParents() *pendingParents {
	return &pendingParents{
		gateways: map[string]bool{},
		nodes:    map[string]bool{},
		sources:  map[string]bool{},
	}
}

func nodeParentKey(gatewayID, nodeID string) string {
	return gatewayID + "." + nodeID
}

func sourceParentKey(gatewayID, nodeID, sourceID string) string {
	return gatewayID + "." + nodeID + "." + sourceID
}

func (p *pendingParents) remember(resource Resource, present bool) {
	if p == nil {
		return
	}
	gatewayID, nodeID, sourceID, _ := resource.NaturalKeys()
	switch resource.Kind {
	case KindGateway:
		if id := resource.ID(); id != "" {
			p.gateways[id] = present
		}
	case KindNode:
		p.nodes[nodeParentKey(gatewayID, nodeID)] = present
	case KindSource:
		p.sources[sourceParentKey(gatewayID, nodeID, sourceID)] = present
	}
}

func (p *pendingParents) knownGateway(gatewayID string) (present bool, known bool) {
	if p == nil {
		return false, false
	}
	present, known = p.gateways[gatewayID]
	return present, known
}

func (p *pendingParents) knownNode(gatewayID, nodeID string) (present bool, known bool) {
	if p == nil {
		return false, false
	}
	present, known = p.nodes[nodeParentKey(gatewayID, nodeID)]
	return present, known
}

func (p *pendingParents) knownSource(gatewayID, nodeID, sourceID string) (present bool, known bool) {
	if p == nil {
		return false, false
	}
	present, known = p.sources[sourceParentKey(gatewayID, nodeID, sourceID)]
	return present, known
}

func checkParent(client ResourceClient, resource Resource, pending *pendingParents) error {
	gatewayID, nodeID, sourceID, _ := resource.NaturalKeys()
	switch resource.Kind {
	case KindNode:
		if present, known := pending.knownGateway(gatewayID); known {
			if !present {
				return fmt.Errorf("parent gateway %s is not present", gatewayID)
			}
			return nil
		}
		id, err := client.FindGateway(gatewayID)
		if err != nil {
			return fmt.Errorf("parent lookup failed: %w", err)
		}
		if id == "" {
			return fmt.Errorf("parent gateway %s is not present", gatewayID)
		}
	case KindSource:
		if present, known := pending.knownNode(gatewayID, nodeID); known {
			if !present {
				return fmt.Errorf("parent node %s.%s is not present", gatewayID, nodeID)
			}
			return nil
		}
		id, err := client.FindNode("", gatewayID, nodeID)
		if err != nil {
			return fmt.Errorf("parent lookup failed: %w", err)
		}
		if id == "" {
			return fmt.Errorf("parent node %s.%s is not present", gatewayID, nodeID)
		}
	case KindField:
		if present, known := pending.knownSource(gatewayID, nodeID, sourceID); known {
			if !present {
				return fmt.Errorf("parent source %s.%s.%s is not present", gatewayID, nodeID, sourceID)
			}
			return nil
		}
		id, err := client.FindSource("", gatewayID, nodeID, sourceID)
		if err != nil {
			return fmt.Errorf("parent lookup failed: %w", err)
		}
		if id == "" {
			return fmt.Errorf("parent source %s.%s.%s is not present", gatewayID, nodeID, sourceID)
		}
	}
	return nil
}

func checkExecutedParent(resource Resource, executed *pendingParents) error {
	gatewayID, nodeID, sourceID, _ := resource.NaturalKeys()
	switch resource.Kind {
	case KindNode:
		if present, known := executed.knownGateway(gatewayID); known && !present {
			return fmt.Errorf("parent gateway %s is not present", gatewayID)
		}
	case KindSource:
		if present, known := executed.knownNode(gatewayID, nodeID); known && !present {
			return fmt.Errorf("parent node %s.%s is not present", gatewayID, nodeID)
		}
	case KindField:
		if present, known := executed.knownSource(gatewayID, nodeID, sourceID); known && !present {
			return fmt.Errorf("parent source %s.%s.%s is not present", gatewayID, nodeID, sourceID)
		}
	}
	return nil
}

func mergeExisting(resource *Resource, existingJSON []byte) error {
	if len(existingJSON) == 0 {
		return fmt.Errorf("existing resource is empty")
	}
	var base map[string]interface{}
	if err := json.Unmarshal(existingJSON, &base); err != nil {
		return err
	}
	if base == nil {
		base = map[string]interface{}{}
	}
	mergedMap := deepMergeMaps(base, resource.Payload)
	mergedMap["id"] = resource.ID()
	merged, err := json.Marshal(mergedMap)
	if err != nil {
		return err
	}
	return decodeMerged(resource, merged)
}

func decodeMerged(resource *Resource, merged []byte) error {
	switch resource.Kind {
	case KindGateway:
		resource.Gateway = &gwTY.Config{}
		return json.Unmarshal(merged, resource.Gateway)
	case KindNode:
		resource.Node = &nodeTY.Node{}
		return json.Unmarshal(merged, resource.Node)
	case KindSource:
		resource.Src = &sourceTY.Source{}
		return json.Unmarshal(merged, resource.Src)
	case KindField:
		resource.Field = &fieldTY.Field{}
		return json.Unmarshal(merged, resource.Field)
	case KindFirmware:
		resource.Firmware = &firmwareTY.Firmware{}
		return json.Unmarshal(merged, resource.Firmware)
	case KindDataRepository:
		resource.DataRepository = &dataRepoTY.Config{}
		return json.Unmarshal(merged, resource.DataRepository)
	default:
		return fmt.Errorf("unsupported kind %q", resource.Kind)
	}
}

// assignSaveID sets an id when the HTTP API requires one.
// Field create leaves id empty so the server treats it as a new resource.
func assignSaveID(resource *Resource) {
	if resource.ID() != "" {
		return
	}
	if resource.Kind == KindField || resource.Kind == KindGateway || resource.Kind == KindFirmware || resource.Kind == KindDataRepository {
		return
	}
	resource.SetID(utils.RandUUID())
}

func executePlan(client ResourceClient, plan plannedAction) error {
	switch plan.Action {
	case actionAdd, actionMerge:
		return saveResource(client, plan.Resource)
	case actionReplace:
		if err := deleteResource(client, plan.Resource, plan.ExistingID); err != nil {
			return fmt.Errorf("replace delete failed: %w", err)
		}
		if err := saveResource(client, plan.Resource); err != nil {
			if plan.Resource.Kind == KindFirmware {
				return fmt.Errorf("deleted existing resource, but recreate failed: %w (upload the firmware binary again)", err)
			}
			return fmt.Errorf("deleted existing resource, but recreate failed: %w", err)
		}
		return nil
	case actionDelete:
		return deleteResource(client, plan.Resource, plan.ExistingID)
	default:
		return fmt.Errorf("unknown action %q", plan.Action)
	}
}

func findExisting(client ResourceClient, resource Resource) (string, error) {
	id := resource.ID()
	gatewayID, nodeID, sourceID, fieldID := resource.NaturalKeys()
	switch resource.Kind {
	case KindGateway:
		return client.FindGateway(id)
	case KindFirmware:
		return client.FindFirmware(id)
	case KindDataRepository:
		return client.FindDataRepository(id)
	case KindNode:
		return client.FindNode(id, gatewayID, nodeID)
	case KindSource:
		return client.FindSource(id, gatewayID, nodeID, sourceID)
	case KindField:
		return client.FindField(id, gatewayID, nodeID, sourceID, fieldID)
	default:
		return "", fmt.Errorf("unsupported kind %q", resource.Kind)
	}
}

func saveResource(client ResourceClient, resource Resource) error {
	switch resource.Kind {
	case KindGateway:
		return client.SaveGateway(resource)
	case KindFirmware:
		return client.SaveFirmware(resource)
	case KindDataRepository:
		return client.SaveDataRepository(resource)
	case KindNode:
		return client.SaveNode(resource)
	case KindSource:
		return client.SaveSource(resource)
	case KindField:
		return client.SaveField(resource)
	default:
		return fmt.Errorf("unsupported kind %q", resource.Kind)
	}
}

func deleteResource(client ResourceClient, resource Resource, existingID string) error {
	id := existingID
	if id == "" {
		id = resource.ID()
	}
	if id == "" {
		return fmt.Errorf("missing id for delete")
	}
	switch resource.Kind {
	case KindGateway:
		return client.DeleteGateway(id)
	case KindFirmware:
		return client.DeleteFirmware(id)
	case KindDataRepository:
		return client.DeleteDataRepository(id)
	case KindNode:
		return client.DeleteNode(id)
	case KindSource:
		return client.DeleteSource(id)
	case KindField:
		return client.DeleteField(id)
	default:
		return fmt.Errorf("unsupported kind %q", resource.Kind)
	}
}
