package apply

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/mycontroller-org/server/v2/pkg/json"
	dataRepoTY "github.com/mycontroller-org/server/v2/pkg/types/data_repository"
	fieldTY "github.com/mycontroller-org/server/v2/pkg/types/field"
	firmwareTY "github.com/mycontroller-org/server/v2/pkg/types/firmware"
	nodeTY "github.com/mycontroller-org/server/v2/pkg/types/node"
	policyTY "github.com/mycontroller-org/server/v2/pkg/types/policy"
	svcAccountTY "github.com/mycontroller-org/server/v2/pkg/types/service_account"
	sourceTY "github.com/mycontroller-org/server/v2/pkg/types/source"
	userTY "github.com/mycontroller-org/server/v2/pkg/types/user"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	gwTY "github.com/mycontroller-org/server/v2/plugin/gateway/types"
	"gopkg.in/yaml.v3"
)

const (
	KindGateway        = "gateway"
	KindNode           = "node"
	KindSource         = "source"
	KindField          = "field"
	KindFirmware       = "firmware"
	KindDataRepository = "data-repository"
	KindUser           = "user"
	KindPolicy         = "policy"
	KindServiceAccount = "service-account"

	OperationAdd    = "add"
	OperationMerge  = "merge"
	OperationDelete = "delete"
)

// Resource is one node, source, or field from a YAML/JSON file.
type Resource struct {
	Kind           string
	Operation      string
	Replace        bool
	Index          int
	Source         string
	Gateway        *gwTY.Config
	Node           *nodeTY.Node
	Src            *sourceTY.Source
	Field          *fieldTY.Field
	Firmware       *firmwareTY.Firmware
	DataRepository *dataRepoTY.Config
	User           *userTY.User
	Policy         *policyTY.Policy
	ServiceAccount *svcAccountTY.ServiceAccount
	Payload        map[string]interface{}
}

func (r Resource) ID() string {
	switch r.Kind {
	case KindGateway:
		if r.Gateway != nil {
			return r.Gateway.ID
		}
	case KindNode:
		if r.Node != nil {
			return r.Node.ID
		}
	case KindSource:
		if r.Src != nil {
			return r.Src.ID
		}
	case KindField:
		if r.Field != nil {
			return r.Field.ID
		}
	case KindFirmware:
		if r.Firmware != nil {
			return r.Firmware.ID
		}
	case KindDataRepository:
		if r.DataRepository != nil {
			return r.DataRepository.ID
		}
	case KindUser:
		if r.User != nil {
			return r.User.ID
		}
	case KindPolicy:
		if r.Policy != nil {
			return r.Policy.ID
		}
	case KindServiceAccount:
		if r.ServiceAccount != nil {
			return r.ServiceAccount.ID
		}
	}
	return ""
}

func (r Resource) SetID(id string) {
	switch r.Kind {
	case KindGateway:
		if r.Gateway != nil {
			r.Gateway.ID = id
		}
	case KindNode:
		if r.Node != nil {
			r.Node.ID = id
		}
	case KindSource:
		if r.Src != nil {
			r.Src.ID = id
		}
	case KindField:
		if r.Field != nil {
			r.Field.ID = id
		}
	case KindFirmware:
		if r.Firmware != nil {
			r.Firmware.ID = id
		}
	case KindDataRepository:
		if r.DataRepository != nil {
			r.DataRepository.ID = id
		}
	case KindUser:
		if r.User != nil {
			r.User.ID = id
		}
	case KindPolicy:
		if r.Policy != nil {
			r.Policy.ID = id
		}
	case KindServiceAccount:
		if r.ServiceAccount != nil {
			r.ServiceAccount.ID = id
		}
	}
}

func (r Resource) NaturalKeys() (gatewayID, nodeID, sourceID, fieldID string) {
	switch r.Kind {
	case KindGateway:
		if r.Gateway != nil {
			return r.Gateway.ID, "", "", ""
		}
	case KindNode:
		if r.Node != nil {
			return r.Node.GatewayID, r.Node.NodeID, "", ""
		}
	case KindSource:
		if r.Src != nil {
			return r.Src.GatewayID, r.Src.NodeID, r.Src.SourceID, ""
		}
	case KindField:
		if r.Field != nil {
			return r.Field.GatewayID, r.Field.NodeID, r.Field.SourceID, r.Field.FieldID
		}
	case KindFirmware:
		if r.Firmware != nil {
			return r.Firmware.ID, "", "", ""
		}
	case KindDataRepository:
		if r.DataRepository != nil {
			return r.DataRepository.ID, "", "", ""
		}
	case KindUser:
		if r.User != nil {
			return r.User.Username, "", "", ""
		}
	case KindPolicy:
		if r.Policy != nil {
			return r.Policy.ID, "", "", ""
		}
	case KindServiceAccount:
		if r.ServiceAccount != nil {
			userRef := r.ServiceAccount.Username
			if userRef == "" {
				userRef = r.ServiceAccount.UserID
			}
			return userRef, r.ServiceAccount.Name, "", ""
		}
	}
	return "", "", "", ""
}

func (r Resource) TableResource() string {
	gatewayID, nodeID, sourceID, fieldID := r.NaturalKeys()
	parts := make([]string, 0, 4)
	if gatewayID != "" {
		parts = append(parts, gatewayID)
	}
	if nodeID != "" {
		parts = append(parts, nodeID)
	}
	if sourceID != "" {
		parts = append(parts, sourceID)
	}
	if fieldID != "" {
		parts = append(parts, fieldID)
	}
	if len(parts) > 0 {
		return fmt.Sprintf("%s: %s", r.Kind, strings.Join(parts, "."))
	}
	if id := r.ID(); id != "" {
		return fmt.Sprintf("%s: %s", r.Kind, id)
	}
	return r.Kind
}

func (r Resource) Identity() string {
	id := r.ID()
	gatewayID, nodeID, sourceID, fieldID := r.NaturalKeys()
	parts := make([]string, 0, 4)
	if gatewayID != "" {
		parts = append(parts, gatewayID)
	}
	if nodeID != "" {
		parts = append(parts, nodeID)
	}
	if sourceID != "" {
		parts = append(parts, sourceID)
	}
	if fieldID != "" {
		parts = append(parts, fieldID)
	}
	natural := strings.Join(parts, ".")
	switch {
	case id != "" && natural != "":
		return fmt.Sprintf("%s %s (id=%s)", r.Kind, natural, id)
	case id != "":
		return fmt.Sprintf("%s id=%s", r.Kind, id)
	case natural != "":
		return fmt.Sprintf("%s %s", r.Kind, natural)
	default:
		return fmt.Sprintf("%s #%d", r.Kind, r.Index+1)
	}
}

func (r Resource) Validate() error {
	if r.Kind == "" {
		return fmt.Errorf("kind is required")
	}
	if r.Operation == "" {
		return fmt.Errorf("operation is required")
	}

	gatewayID, nodeID, sourceID, fieldID := r.NaturalKeys()
	hasID := r.ID() != ""

	switch r.Kind {
	case KindGateway:
		if !hasID {
			return fmt.Errorf("gateway requires id")
		}
	case KindFirmware:
		if !hasID {
			return fmt.Errorf("firmware requires id")
		}
	case KindDataRepository:
		if !hasID {
			return fmt.Errorf("data-repository requires id")
		}
	case KindUser:
		if r.Operation == OperationAdd && (gatewayID == "" || (r.User != nil && r.User.Password == "")) {
			return fmt.Errorf("add user requires username and password")
		}
		if r.Operation != OperationAdd && !hasID && gatewayID == "" {
			return fmt.Errorf("user requires id or username")
		}
	case KindPolicy:
		if r.Operation != OperationDelete && !hasID && r.Operation != OperationAdd {
			return fmt.Errorf("policy requires id")
		}
	case KindServiceAccount:
		if r.Operation != OperationDelete && nodeID == "" {
			return fmt.Errorf("service-account requires name")
		}
		if r.Operation == OperationDelete && !hasID && nodeID == "" {
			return fmt.Errorf("delete service-account requires id or name")
		}
	case KindNode:
		if r.Operation != OperationDelete && (gatewayID == "" || nodeID == "") {
			return fmt.Errorf("node requires gatewayId and nodeId")
		}
		if r.Operation == OperationDelete && !hasID && (gatewayID == "" || nodeID == "") {
			return fmt.Errorf("delete node requires id or gatewayId+nodeId")
		}
	case KindSource:
		if r.Operation != OperationDelete && (gatewayID == "" || nodeID == "" || sourceID == "") {
			return fmt.Errorf("source requires gatewayId, nodeId and sourceId")
		}
		if r.Operation == OperationDelete && !hasID && (gatewayID == "" || nodeID == "" || sourceID == "") {
			return fmt.Errorf("delete source requires id or gatewayId+nodeId+sourceId")
		}
	case KindField:
		if r.Operation != OperationDelete && (gatewayID == "" || nodeID == "" || sourceID == "" || fieldID == "") {
			return fmt.Errorf("field requires gatewayId, nodeId, sourceId and fieldId")
		}
		if r.Operation == OperationDelete && !hasID && (gatewayID == "" || nodeID == "" || sourceID == "" || fieldID == "") {
			return fmt.Errorf("delete field requires id or gatewayId+nodeId+sourceId+fieldId")
		}
	default:
		return fmt.Errorf("unsupported kind %q", r.Kind)
	}

	switch r.Operation {
	case OperationAdd, OperationMerge, OperationDelete:
	default:
		return fmt.Errorf("unsupported operation %q", r.Operation)
	}
	return nil
}

// ParseResources reads one or more resources from YAML or JSON.
// Supported shapes: a single object, a list, or YAML documents separated by ---.
func ParseResources(data []byte, source string) ([]Resource, error) {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 {
		return nil, fmt.Errorf("no resources found in %s", sourceName(source))
	}

	var docs []map[string]interface{}
	var err error
	if looksLikeJSON(trimmed) {
		docs, err = parseJSONDocuments(trimmed)
	} else {
		docs, err = parseYAMLDocuments(trimmed)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", sourceName(source), err)
	}
	if len(docs) == 0 {
		return nil, fmt.Errorf("no resources found in %s", sourceName(source))
	}

	resources := make([]Resource, 0, len(docs))
	for i, doc := range docs {
		parsed, err := resourcesFromMap(doc, len(resources), source)
		if err != nil {
			return nil, fmt.Errorf("%s: resource %d: %w", sourceName(source), i+1, err)
		}
		resources = append(resources, parsed...)
	}
	return resources, nil
}

func looksLikeJSON(data []byte) bool {
	if len(data) == 0 {
		return false
	}
	switch data[0] {
	case '{', '[':
		return true
	default:
		return false
	}
}

func parseJSONDocuments(data []byte) ([]map[string]interface{}, error) {
	var raw interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}
	return documentsFromValue(raw)
}

func parseYAMLDocuments(data []byte) ([]map[string]interface{}, error) {
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	var docs []map[string]interface{}
	for {
		var raw interface{}
		err := decoder.Decode(&raw)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if raw == nil {
			continue
		}
		parsed, err := documentsFromValue(raw)
		if err != nil {
			return nil, err
		}
		docs = append(docs, parsed...)
	}
	return docs, nil
}

func documentsFromValue(raw interface{}) ([]map[string]interface{}, error) {
	switch value := raw.(type) {
	case map[string]interface{}:
		return []map[string]interface{}{value}, nil
	case []interface{}:
		docs := make([]map[string]interface{}, 0, len(value))
		for i, item := range value {
			doc, ok := item.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("item %d is not an object", i+1)
			}
			docs = append(docs, doc)
		}
		return docs, nil
	default:
		return nil, fmt.Errorf("expected an object or a list of objects, got %T", raw)
	}
}

func resourcesFromMap(doc map[string]interface{}, index int, source string) ([]Resource, error) {
	rawItems, hasItems := doc["items"]
	if !hasItems || rawItems == nil {
		resource, err := resourceFromMap(doc, index, source)
		if err != nil {
			return nil, err
		}
		return []Resource{resource}, nil
	}

	items, err := asObjectList(rawItems)
	if err != nil {
		return nil, fmt.Errorf("items: %w", err)
	}
	if len(items) == 0 {
		return nil, fmt.Errorf("items must not be empty")
	}

	defaults := copyMap(doc)
	delete(defaults, "items")

	resources := make([]Resource, 0, len(items))
	for i, item := range items {
		merged := copyMap(defaults)
		for key, value := range item {
			merged[key] = value
		}
		resource, err := resourceFromMap(merged, index+i, source)
		if err != nil {
			return nil, fmt.Errorf("items[%d]: %w", i, err)
		}
		resources = append(resources, resource)
	}
	return resources, nil
}

func asObjectList(raw interface{}) ([]map[string]interface{}, error) {
	list, ok := raw.([]interface{})
	if !ok {
		return nil, fmt.Errorf("must be a list of objects")
	}
	items := make([]map[string]interface{}, 0, len(list))
	for i, item := range list {
		doc, ok := item.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("item %d is not an object", i+1)
		}
		items = append(items, doc)
	}
	return items, nil
}

func resourceFromMap(doc map[string]interface{}, index int, source string) (Resource, error) {
	kind, err := effectiveKind(doc)
	if err != nil {
		return Resource{}, err
	}
	operation, err := normalizeOperation(stringValue(doc, "operation"))
	if err != nil {
		return Resource{}, err
	}

	payload := copyMap(doc)
	delete(payload, "kind")
	delete(payload, "operation")
	delete(payload, "replace")
	delete(payload, "items")

	resource := Resource{
		Kind:      kind,
		Operation: operation,
		Replace:   boolValue(doc, "replace"),
		Index:     index,
		Source:    source,
		Payload:   payload,
	}

	switch kind {
	case KindGateway:
		gateway := &gwTY.Config{}
		if err := utils.MapToStruct(utils.TagNameJSON, payload, gateway); err != nil {
			return Resource{}, fmt.Errorf("invalid gateway: %w", err)
		}
		resource.Gateway = gateway
	case KindNode:
		node := &nodeTY.Node{}
		if err := utils.MapToStruct(utils.TagNameJSON, payload, node); err != nil {
			return Resource{}, fmt.Errorf("invalid node: %w", err)
		}
		resource.Node = node
	case KindSource:
		src := &sourceTY.Source{}
		if err := utils.MapToStruct(utils.TagNameJSON, payload, src); err != nil {
			return Resource{}, fmt.Errorf("invalid source: %w", err)
		}
		resource.Src = src
	case KindField:
		field := &fieldTY.Field{}
		if err := utils.MapToStruct(utils.TagNameJSON, payload, field); err != nil {
			return Resource{}, fmt.Errorf("invalid field: %w", err)
		}
		resource.Field = field
	case KindFirmware:
		firmware := &firmwareTY.Firmware{}
		if err := utils.MapToStruct(utils.TagNameJSON, payload, firmware); err != nil {
			return Resource{}, fmt.Errorf("invalid firmware: %w", err)
		}
		resource.Firmware = firmware
	case KindDataRepository:
		item := &dataRepoTY.Config{}
		if err := utils.MapToStruct(utils.TagNameJSON, payload, item); err != nil {
			return Resource{}, fmt.Errorf("invalid data-repository: %w", err)
		}
		resource.DataRepository = item
	case KindUser:
		user := &userTY.User{}
		if err := utils.MapToStruct(utils.TagNameJSON, payload, user); err != nil {
			return Resource{}, fmt.Errorf("invalid user: %w", err)
		}
		if _, ok := payload["enabled"]; !ok {
			user.Enabled = true
		}
		resource.User = user
	case KindPolicy:
		policy := &policyTY.Policy{}
		if err := utils.MapToStruct(utils.TagNameJSON, payload, policy); err != nil {
			return Resource{}, fmt.Errorf("invalid policy: %w", err)
		}
		resource.Policy = policy
	case KindServiceAccount:
		account := &svcAccountTY.ServiceAccount{}
		if err := utils.MapToStruct(utils.TagNameJSON, payload, account); err != nil {
			return Resource{}, fmt.Errorf("invalid service-account: %w", err)
		}
		if _, ok := payload["enabled"]; !ok {
			account.Enabled = true
		}
		resource.ServiceAccount = account
	}

	if err := resource.Validate(); err != nil {
		return Resource{}, err
	}
	return resource, nil
}

func effectiveKind(doc map[string]interface{}) (string, error) {
	if hasNonEmpty(doc, "fieldId") {
		return KindField, nil
	}
	return normalizeKind(stringValue(doc, "kind"))
}

func normalizeKind(kind string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(kind)) {
	case KindGateway, "gw", "gateways":
		return KindGateway, nil
	case KindNode, "nodes":
		return KindNode, nil
	case KindSource, "sources":
		return KindSource, nil
	case KindField, "fields":
		return KindField, nil
	case KindFirmware, "firmwares", "fw":
		return KindFirmware, nil
	case KindDataRepository, "datarepository", "data-repo", "data-repositories", "datarepo":
		return KindDataRepository, nil
	case KindUser, "users":
		return KindUser, nil
	case KindPolicy, "policies":
		return KindPolicy, nil
	case KindServiceAccount, "service-accounts", "sa":
		return KindServiceAccount, nil
	case "":
		return "", fmt.Errorf("kind is required")
	default:
		return "", fmt.Errorf("unsupported kind %q (supported: gateway, node, source, field, firmware, data-repository, user, policy, service-account)", kind)
	}
}

func normalizeOperation(operation string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(operation)) {
	case OperationAdd, "create":
		return OperationAdd, nil
	case OperationMerge, "update":
		return OperationMerge, nil
	case OperationDelete, "remove":
		return OperationDelete, nil
	case "":
		return "", fmt.Errorf("operation is required")
	default:
		return "", fmt.Errorf("unsupported operation %q (supported: add, merge, delete)", operation)
	}
}

func stringValue(doc map[string]interface{}, key string) string {
	value, ok := doc[key]
	if !ok || value == nil {
		return ""
	}
	switch typed := value.(type) {
	case string:
		return typed
	default:
		return fmt.Sprintf("%v", typed)
	}
}

func hasNonEmpty(doc map[string]interface{}, key string) bool {
	return strings.TrimSpace(stringValue(doc, key)) != ""
}

func boolValue(doc map[string]interface{}, key string) bool {
	value, ok := doc[key]
	if !ok || value == nil {
		return false
	}
	switch typed := value.(type) {
	case bool:
		return typed
	case string:
		switch strings.ToLower(strings.TrimSpace(typed)) {
		case "true", "yes", "1":
			return true
		default:
			return false
		}
	default:
		return false
	}
}

func copyMap(in map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{}, len(in))
	for key, value := range in {
		out[key] = value
	}
	return out
}

func sourceName(source string) string {
	if source == "" {
		return "input"
	}
	return source
}
