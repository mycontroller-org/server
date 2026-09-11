package apply

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	dataRepoTY "github.com/mycontroller-org/server/v2/pkg/types/data_repository"
	fieldTY "github.com/mycontroller-org/server/v2/pkg/types/field"
	firmwareTY "github.com/mycontroller-org/server/v2/pkg/types/firmware"
	nodeTY "github.com/mycontroller-org/server/v2/pkg/types/node"
	sourceTY "github.com/mycontroller-org/server/v2/pkg/types/source"
	gwTY "github.com/mycontroller-org/server/v2/plugin/gateway/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeClient struct {
	existing     map[string]string
	existingJSON map[string][]byte
	saved        []Resource
	deleted      []string
	saveErr      error
	findErr      error
}

func (f *fakeClient) find(kind, id, gatewayID, nodeID, sourceID, fieldID string) (string, error) {
	if f.findErr != nil {
		return "", f.findErr
	}
	if f.existing == nil {
		return "", nil
	}
	if id != "" {
		if found, ok := f.existing[kind+"/id/"+id]; ok {
			return found, nil
		}
	}
	if gatewayID != "" {
		if found, ok := f.existing[strings.Join([]string{kind, gatewayID, nodeID, sourceID, fieldID}, "/")]; ok {
			return found, nil
		}
	}
	return "", nil
}

func (f *fakeClient) FindGateway(id string) (string, error) {
	return f.find("gateway", id, "", "", "", "")
}
func (f *fakeClient) FindNode(id, gatewayID, nodeID string) (string, error) {
	return f.find(KindNode, id, gatewayID, nodeID, "", "")
}
func (f *fakeClient) FindSource(id, gatewayID, nodeID, sourceID string) (string, error) {
	return f.find(KindSource, id, gatewayID, nodeID, sourceID, "")
}
func (f *fakeClient) FindField(id, gatewayID, nodeID, sourceID, fieldID string) (string, error) {
	return f.find(KindField, id, gatewayID, nodeID, sourceID, fieldID)
}
func (f *fakeClient) FindFirmware(id string) (string, error) {
	return f.find(KindFirmware, id, "", "", "", "")
}
func (f *fakeClient) FindDataRepository(id string) (string, error) {
	return f.find(KindDataRepository, id, "", "", "", "")
}
func (f *fakeClient) SaveGateway(resource Resource) error {
	if f.saveErr != nil {
		return f.saveErr
	}
	f.saved = append(f.saved, resource)
	return nil
}
func (f *fakeClient) SaveNode(resource Resource) error {
	if f.saveErr != nil {
		return f.saveErr
	}
	f.saved = append(f.saved, resource)
	return nil
}
func (f *fakeClient) SaveSource(resource Resource) error {
	if f.saveErr != nil {
		return f.saveErr
	}
	f.saved = append(f.saved, resource)
	return nil
}
func (f *fakeClient) SaveField(resource Resource) error {
	if f.saveErr != nil {
		return f.saveErr
	}
	f.saved = append(f.saved, resource)
	return nil
}
func (f *fakeClient) SaveFirmware(resource Resource) error {
	if f.saveErr != nil {
		return f.saveErr
	}
	f.saved = append(f.saved, resource)
	return nil
}
func (f *fakeClient) SaveDataRepository(resource Resource) error {
	if f.saveErr != nil {
		return f.saveErr
	}
	f.saved = append(f.saved, resource)
	return nil
}
func (f *fakeClient) DeleteGateway(ids ...string) error {
	f.deleted = append(f.deleted, ids...)
	return nil
}
func (f *fakeClient) DeleteNode(ids ...string) error {
	f.deleted = append(f.deleted, ids...)
	return nil
}
func (f *fakeClient) DeleteSource(ids ...string) error {
	f.deleted = append(f.deleted, ids...)
	return nil
}
func (f *fakeClient) DeleteField(ids ...string) error {
	f.deleted = append(f.deleted, ids...)
	return nil
}
func (f *fakeClient) DeleteFirmware(ids ...string) error {
	f.deleted = append(f.deleted, ids...)
	return nil
}
func (f *fakeClient) DeleteDataRepository(ids ...string) error {
	f.deleted = append(f.deleted, ids...)
	return nil
}
func (f *fakeClient) GetExisting(resource Resource) ([]byte, error) {
	id := resource.ID()
	if f.existingJSON != nil {
		if data, ok := f.existingJSON[id]; ok {
			return data, nil
		}
	}
	if id == "" {
		return []byte("{}"), nil
	}
	return []byte(`{"id":"` + id + `"}`), nil
}

func assertApplyRow(t *testing.T, out, resource, action, status string) {
	t.Helper()
	assert.Contains(t, out, "RESOURCE")
	assert.Contains(t, out, "ACTION")
	assert.Contains(t, out, "STATUS")
	for _, line := range strings.Split(out, "\n") {
		if strings.Contains(line, resource) && strings.Contains(line, action) && strings.Contains(line, status) {
			return
		}
	}
	t.Fatalf("missing row resource=%q action=%q status=%q\noutput:\n%s", resource, action, status, out)
}

func withGateway(existing map[string]string, gatewayIDs ...string) map[string]string {
	if existing == nil {
		existing = map[string]string{}
	}
	for _, gatewayID := range gatewayIDs {
		existing["gateway/id/"+gatewayID] = gatewayID
	}
	return existing
}

func testFirmware(operation, id string) Resource {
	return Resource{
		Kind:      KindFirmware,
		Operation: operation,
		Firmware: &firmwareTY.Firmware{
			ID:          id,
			Description: "fw",
		},
		Payload: map[string]interface{}{"description": "fw"},
	}
}

func testDataRepository(operation, id string) Resource {
	return Resource{
		Kind:      KindDataRepository,
		Operation: operation,
		DataRepository: &dataRepoTY.Config{
			ID:          id,
			Description: "repo",
		},
		Payload: map[string]interface{}{"description": "repo"},
	}
}

func testGateway(operation, id string) Resource {
	return Resource{
		Kind:      KindGateway,
		Operation: operation,
		Gateway: &gwTY.Config{
			ID:          id,
			Description: "gw",
			Enabled:     true,
		},
		Payload: map[string]interface{}{"description": "gw", "enabled": true},
	}
}

func testNode(operation, gatewayID, nodeID, id string) Resource {
	return Resource{
		Kind:      KindNode,
		Operation: operation,
		Node: &nodeTY.Node{
			ID:        id,
			GatewayID: gatewayID,
			NodeID:    nodeID,
			Name:      "n",
		},
		Payload: map[string]interface{}{
			"gatewayId": gatewayID,
			"nodeId":    nodeID,
			"name":      "n",
		},
	}
}

func testSource(operation, gatewayID, nodeID, sourceID, id string) Resource {
	return Resource{
		Kind:      KindSource,
		Operation: operation,
		Src: &sourceTY.Source{
			ID:        id,
			GatewayID: gatewayID,
			NodeID:    nodeID,
			SourceID:  sourceID,
			Name:      "s",
		},
		Payload: map[string]interface{}{
			"gatewayId": gatewayID,
			"nodeId":    nodeID,
			"sourceId":  sourceID,
			"name":      "s",
		},
	}
}

func testField(operation, gatewayID, nodeID, sourceID, fieldID, id string) Resource {
	return Resource{
		Kind:      KindField,
		Operation: operation,
		Field: &fieldTY.Field{
			ID:        id,
			GatewayID: gatewayID,
			NodeID:    nodeID,
			SourceID:  sourceID,
			FieldID:   fieldID,
			Name:      "f",
		},
		Payload: map[string]interface{}{
			"gatewayId": gatewayID,
			"nodeId":    nodeID,
			"sourceId":  sourceID,
			"fieldId":   fieldID,
			"name":      "f",
		},
	}
}

func TestApplyAddFieldLeavesIDEmpty(t *testing.T) {
	client := &fakeClient{existing: map[string]string{
		"source/gw1/n1/s1/": "source-id",
	}}
	resource := Resource{
		Kind:      KindField,
		Operation: OperationAdd,
		Field: &fieldTY.Field{
			GatewayID: "gw1",
			NodeID:    "n1",
			SourceID:  "s1",
			FieldID:   "temp",
			Name:      "Temperature",
		},
	}
	out, errOut := &bytes.Buffer{}, &bytes.Buffer{}
	err := Apply(client, []Resource{resource}, false, false, out, errOut)
	require.NoError(t, err)
	require.Len(t, client.saved, 1)
	assert.Empty(t, client.saved[0].Field.ID)
	assertApplyRow(t, out.String(), "field: gw1.n1.s1.temp", "add", "ok")
}

func TestApplyAddNewResource(t *testing.T) {
	client := &fakeClient{existing: withGateway(nil, "gw1")}
	out, errOut := &bytes.Buffer{}, &bytes.Buffer{}
	err := Apply(client, []Resource{testNode(OperationAdd, "gw1", "n1", "")}, false, false, out, errOut)
	require.NoError(t, err)
	require.Len(t, client.saved, 1)
	assert.NotEmpty(t, client.saved[0].Node.ID)
	assertApplyRow(t, out.String(), "node: gw1.n1", "add", "ok")
	assert.Empty(t, client.deleted)
}

func TestApplyAddExistingFails(t *testing.T) {
	client := &fakeClient{existing: map[string]string{
		"node/gw1/n1//": "existing-id",
	}}
	out, errOut := &bytes.Buffer{}, &bytes.Buffer{}
	err := Apply(client, []Resource{testNode(OperationAdd, "gw1", "n1", "")}, false, false, out, errOut)
	require.Error(t, err)
	assert.Empty(t, client.saved)
	assertApplyRow(t, out.String(), "node: gw1.n1", "add", "failed: already exists")
}

func TestApplyReplaceFromFileFlag(t *testing.T) {
	client := &fakeClient{existing: withGateway(map[string]string{
		"node/gw1/n1//": "existing-id",
	}, "gw1")}
	resource := testNode(OperationAdd, "gw1", "n1", "")
	resource.Replace = true
	out, errOut := &bytes.Buffer{}, &bytes.Buffer{}
	err := Apply(client, []Resource{resource}, false, false, out, errOut)
	require.NoError(t, err)
	assert.Equal(t, []string{"existing-id"}, client.deleted)
	require.Len(t, client.saved, 1)
	assert.Equal(t, "existing-id", client.saved[0].Node.ID)
	assertApplyRow(t, out.String(), "node: gw1.n1", "add", "replaced")
}

func TestApplyAddExistingReplace(t *testing.T) {
	client := &fakeClient{existing: withGateway(map[string]string{
		"node/gw1/n1//": "existing-id",
	}, "gw1")}
	out, errOut := &bytes.Buffer{}, &bytes.Buffer{}
	err := Apply(client, []Resource{testNode(OperationAdd, "gw1", "n1", "")}, true, false, out, errOut)
	require.NoError(t, err)
	assert.Equal(t, []string{"existing-id"}, client.deleted)
	require.Len(t, client.saved, 1)
	assert.Equal(t, "existing-id", client.saved[0].Node.ID)
	assertApplyRow(t, out.String(), "node: gw1.n1", "add", "replaced")
}

func TestApplyReplaceFieldKeepsExistingID(t *testing.T) {
	client := &fakeClient{existing: map[string]string{
		"source/gw1/n1/s1/":    "source-id",
		"field/gw1/n1/s1/temp": "field-id",
	}}
	resource := Resource{
		Kind:      KindField,
		Operation: OperationAdd,
		Field: &fieldTY.Field{
			GatewayID: "gw1",
			NodeID:    "n1",
			SourceID:  "s1",
			FieldID:   "temp",
			Name:      "Temperature",
		},
	}
	out, errOut := &bytes.Buffer{}, &bytes.Buffer{}
	err := Apply(client, []Resource{resource}, true, false, out, errOut)
	require.NoError(t, err)
	assert.Equal(t, []string{"field-id"}, client.deleted)
	require.Len(t, client.saved, 1)
	assert.Equal(t, "field-id", client.saved[0].Field.ID)
	assertApplyRow(t, out.String(), "field: gw1.n1.s1.temp", "add", "replaced")
}

func TestApplyReplaceKeepsDeletedIDWhenFileHasDifferentID(t *testing.T) {
	client := &fakeClient{existing: withGateway(map[string]string{
		"node/gw1/n1//": "existing-id",
	}, "gw1")}
	out, errOut := &bytes.Buffer{}, &bytes.Buffer{}
	err := Apply(client, []Resource{testNode(OperationAdd, "gw1", "n1", "file-id")}, true, false, out, errOut)
	require.NoError(t, err)
	assert.Equal(t, []string{"existing-id"}, client.deleted)
	require.Len(t, client.saved, 1)
	assert.Equal(t, "existing-id", client.saved[0].Node.ID)
}

func TestApplyUpdateMissingFails(t *testing.T) {
	client := &fakeClient{}
	out, errOut := &bytes.Buffer{}, &bytes.Buffer{}
	err := Apply(client, []Resource{testNode(OperationMerge, "gw1", "n1", "")}, false, false, out, errOut)
	require.Error(t, err)
	assert.Empty(t, client.saved)
	assertApplyRow(t, out.String(), "node: gw1.n1", "merge", "failed: not found")
}

func TestApplyUpdateExisting(t *testing.T) {
	client := &fakeClient{existing: withGateway(map[string]string{
		"node/gw1/n1//": "existing-id",
	}, "gw1")}
	out, errOut := &bytes.Buffer{}, &bytes.Buffer{}
	err := Apply(client, []Resource{testNode(OperationMerge, "gw1", "n1", "")}, false, false, out, errOut)
	require.NoError(t, err)
	require.Len(t, client.saved, 1)
	assert.Equal(t, "existing-id", client.saved[0].Node.ID)
	assert.Empty(t, client.deleted)
	assertApplyRow(t, out.String(), "node: gw1.n1", "merge", "ok")
}

func TestApplyUpdateMergesExistingFields(t *testing.T) {
	client := &fakeClient{
		existing: withGateway(map[string]string{
			"node/gw1/n1//": "existing-id",
		}, "gw1"),
		existingJSON: map[string][]byte{
			"existing-id": []byte(`{"id":"existing-id","gatewayId":"gw1","nodeId":"n1","name":"old","labels":{"keep":"yes","room":"kitchen"}}`),
		},
	}
	resource := testNode(OperationMerge, "gw1", "n1", "")
	resource.Payload = map[string]interface{}{
		"gatewayId": "gw1",
		"nodeId":    "n1",
		"name":      "new",
		"labels": map[string]interface{}{
			"room": "living",
		},
	}
	out, errOut := &bytes.Buffer{}, &bytes.Buffer{}
	err := Apply(client, []Resource{resource}, false, false, out, errOut)
	require.NoError(t, err)
	require.Len(t, client.saved, 1)
	assert.Equal(t, "existing-id", client.saved[0].Node.ID)
	assert.Equal(t, "new", client.saved[0].Node.Name)
	assert.Equal(t, "yes", client.saved[0].Node.Labels["keep"])
	assert.Equal(t, "living", client.saved[0].Node.Labels["room"])
}

func TestApplyUpdateKeepsExistingIDWhenFileHasDifferentID(t *testing.T) {
	client := &fakeClient{existing: withGateway(map[string]string{
		"node/gw1/n1//": "existing-id",
	}, "gw1")}
	out, errOut := &bytes.Buffer{}, &bytes.Buffer{}
	err := Apply(client, []Resource{testNode(OperationMerge, "gw1", "n1", "file-id")}, false, false, out, errOut)
	require.NoError(t, err)
	require.Len(t, client.saved, 1)
	assert.Equal(t, "existing-id", client.saved[0].Node.ID)
	assert.Empty(t, client.deleted)
}

func TestApplyParentSaveFailureSkipsChild(t *testing.T) {
	client := &fakeClient{saveErr: fmt.Errorf("boom")}
	out, errOut := &bytes.Buffer{}, &bytes.Buffer{}
	err := Apply(client, []Resource{
		testGateway(OperationAdd, "gw1"),
		testNode(OperationAdd, "gw1", "n1", ""),
	}, false, false, out, errOut)
	require.Error(t, err)
	assert.Empty(t, client.saved)
	assertApplyRow(t, out.String(), "gateway: gw1", "add", "failed: boom")
	assertApplyRow(t, out.String(), "node: gw1.n1", "add", "failed: parent gateway gw1 is not present")
}

func TestApplyReplaceSaveFailureMentionsDeleted(t *testing.T) {
	client := &fakeClient{
		existing: withGateway(map[string]string{
			"node/gw1/n1//": "existing-id",
		}, "gw1"),
		saveErr: fmt.Errorf("disk full"),
	}
	out, errOut := &bytes.Buffer{}, &bytes.Buffer{}
	err := Apply(client, []Resource{testNode(OperationAdd, "gw1", "n1", "")}, true, false, out, errOut)
	require.Error(t, err)
	assert.Equal(t, []string{"existing-id"}, client.deleted)
	assert.Empty(t, client.saved)
	assertApplyRow(t, out.String(), "node: gw1.n1", "add", "deleted existing resource, but recreate failed")
}

func TestApplyDeleteMissingContinues(t *testing.T) {
	client := &fakeClient{existing: withGateway(nil, "gw1")}
	out, errOut := &bytes.Buffer{}, &bytes.Buffer{}
	err := Apply(client, []Resource{
		testNode(OperationDelete, "gw1", "n1", ""),
		testNode(OperationAdd, "gw1", "n2", ""),
	}, false, false, out, errOut)
	require.NoError(t, err)
	assert.Empty(t, client.deleted)
	assert.Empty(t, errOut.String())
	assertApplyRow(t, out.String(), "node: gw1.n1", "delete", "not available")
	assertApplyRow(t, out.String(), "node: gw1.n2", "add", "ok")
	require.Len(t, client.saved, 1)
}

func TestApplyDeleteMissingDryRunContinues(t *testing.T) {
	client := &fakeClient{}
	out, errOut := &bytes.Buffer{}, &bytes.Buffer{}
	err := Apply(client, []Resource{testNode(OperationDelete, "gw1", "n1", "")}, false, true, out, errOut)
	require.NoError(t, err)
	assert.Empty(t, client.deleted)
	assert.Empty(t, errOut.String())
	assertApplyRow(t, out.String(), "node: gw1.n1", "delete", "not available")
}

func TestApplyDeleteExisting(t *testing.T) {
	client := &fakeClient{existing: map[string]string{
		"node/gw1/n1//": "existing-id",
	}}
	out, errOut := &bytes.Buffer{}, &bytes.Buffer{}
	err := Apply(client, []Resource{testNode(OperationDelete, "gw1", "n1", "")}, false, false, out, errOut)
	require.NoError(t, err)
	assert.Equal(t, []string{"existing-id"}, client.deleted)
	assert.Empty(t, client.saved)
	assertApplyRow(t, out.String(), "node: gw1.n1", "delete", "ok")
}

func TestApplyDryRunDoesNotMutate(t *testing.T) {
	client := &fakeClient{existing: withGateway(nil, "gw1")}
	out, errOut := &bytes.Buffer{}, &bytes.Buffer{}
	err := Apply(client, []Resource{testNode(OperationAdd, "gw1", "n1", "")}, false, true, out, errOut)
	require.NoError(t, err)
	assert.Empty(t, client.saved)
	assert.Empty(t, client.deleted)
	assertApplyRow(t, out.String(), "node: gw1.n1", "add", "dry-run")
}

func TestApplyDryRunDetectsExistingAdd(t *testing.T) {
	client := &fakeClient{existing: map[string]string{
		"node/gw1/n1//": "existing-id",
	}}
	out, errOut := &bytes.Buffer{}, &bytes.Buffer{}
	err := Apply(client, []Resource{testNode(OperationAdd, "gw1", "n1", "")}, false, true, out, errOut)
	require.Error(t, err)
	assert.Empty(t, client.saved)
	assert.Empty(t, client.deleted)
	assertApplyRow(t, out.String(), "node: gw1.n1", "add", "failed: already exists")
}

func TestApplyLookupError(t *testing.T) {
	client := &fakeClient{findErr: fmt.Errorf("server down")}
	out, errOut := &bytes.Buffer{}, &bytes.Buffer{}
	err := Apply(client, []Resource{testNode(OperationAdd, "gw1", "n1", "")}, false, true, out, errOut)
	require.Error(t, err)
	assertApplyRow(t, out.String(), "node: gw1.n1", "add", "failed: lookup failed")
}

func TestRunApplyFromStdin(t *testing.T) {
	client := &fakeClient{existing: withGateway(nil, "gw1")}
	in := strings.NewReader(`kind: node
operation: add
gatewayId: gw1
nodeId: n1
name: From stdin
`)
	out, errOut := &bytes.Buffer{}, &bytes.Buffer{}
	err := runApply(client, []string{"-"}, false, true, in, out, errOut)
	require.NoError(t, err)
	assertApplyRow(t, out.String(), "node: gw1.n1", "add", "dry-run")
}

func TestApplyAddNodeFailsWhenGatewayMissing(t *testing.T) {
	client := &fakeClient{}
	out, errOut := &bytes.Buffer{}, &bytes.Buffer{}
	err := Apply(client, []Resource{testNode(OperationAdd, "gw1", "n1", "")}, false, false, out, errOut)
	require.Error(t, err)
	assert.Empty(t, client.saved)
	assertApplyRow(t, out.String(), "node: gw1.n1", "add", "failed: parent gateway gw1 is not present")
}

func TestApplyAddSourceFailsWhenNodeMissing(t *testing.T) {
	client := &fakeClient{existing: withGateway(nil, "gw1")}
	out, errOut := &bytes.Buffer{}, &bytes.Buffer{}
	err := Apply(client, []Resource{testSource(OperationAdd, "gw1", "n1", "s1", "")}, false, false, out, errOut)
	require.Error(t, err)
	assert.Empty(t, client.saved)
	assertApplyRow(t, out.String(), "source: gw1.n1.s1", "add", "failed: parent node gw1.n1 is not present")
}

func TestApplyAddFieldFailsWhenSourceMissing(t *testing.T) {
	client := &fakeClient{existing: withGateway(map[string]string{
		"node/gw1/n1//": "node-id",
	}, "gw1")}
	out, errOut := &bytes.Buffer{}, &bytes.Buffer{}
	err := Apply(client, []Resource{testField(OperationAdd, "gw1", "n1", "s1", "temp", "")}, false, false, out, errOut)
	require.Error(t, err)
	assert.Empty(t, client.saved)
	assertApplyRow(t, out.String(), "field: gw1.n1.s1.temp", "add", "failed: parent source gw1.n1.s1 is not present")
}

func TestApplyUsesParentAddedEarlierInFile(t *testing.T) {
	client := &fakeClient{existing: withGateway(nil, "gw1")}
	out, errOut := &bytes.Buffer{}, &bytes.Buffer{}
	err := Apply(client, []Resource{
		testNode(OperationAdd, "gw1", "n1", ""),
		testSource(OperationAdd, "gw1", "n1", "s1", ""),
		testField(OperationAdd, "gw1", "n1", "s1", "temp", ""),
	}, false, false, out, errOut)
	require.NoError(t, err)
	require.Len(t, client.saved, 3)
	assertApplyRow(t, out.String(), "node: gw1.n1", "add", "ok")
	assertApplyRow(t, out.String(), "source: gw1.n1.s1", "add", "ok")
	assertApplyRow(t, out.String(), "field: gw1.n1.s1.temp", "add", "ok")
}

func TestApplyFailsWhenParentDeletedEarlierInFile(t *testing.T) {
	client := &fakeClient{existing: withGateway(map[string]string{
		"node/gw1/n1//": "node-id",
	}, "gw1")}
	out, errOut := &bytes.Buffer{}, &bytes.Buffer{}
	err := Apply(client, []Resource{
		testNode(OperationDelete, "gw1", "n1", ""),
		testSource(OperationAdd, "gw1", "n1", "s1", ""),
	}, false, false, out, errOut)
	require.Error(t, err)
	assertApplyRow(t, out.String(), "node: gw1.n1", "delete", "ok")
	assertApplyRow(t, out.String(), "source: gw1.n1.s1", "add", "failed: parent node gw1.n1 is not present")
	assert.Equal(t, []string{"node-id"}, client.deleted)
	assert.Empty(t, client.saved)
}

func TestApplyAddGateway(t *testing.T) {
	client := &fakeClient{}
	out, errOut := &bytes.Buffer{}, &bytes.Buffer{}
	err := Apply(client, []Resource{testGateway(OperationAdd, "gw1")}, false, false, out, errOut)
	require.NoError(t, err)
	require.Len(t, client.saved, 1)
	assert.Equal(t, "gw1", client.saved[0].Gateway.ID)
	assertApplyRow(t, out.String(), "gateway: gw1", "add", "ok")
}

func TestApplyReplaceGatewayKeepsID(t *testing.T) {
	client := &fakeClient{existing: withGateway(nil, "gw1")}
	out, errOut := &bytes.Buffer{}, &bytes.Buffer{}
	err := Apply(client, []Resource{testGateway(OperationAdd, "gw1")}, true, false, out, errOut)
	require.NoError(t, err)
	assert.Equal(t, []string{"gw1"}, client.deleted)
	require.Len(t, client.saved, 1)
	assert.Equal(t, "gw1", client.saved[0].Gateway.ID)
	assertApplyRow(t, out.String(), "gateway: gw1", "add", "replaced")
}

func TestApplyDeleteGateway(t *testing.T) {
	client := &fakeClient{existing: withGateway(nil, "gw1")}
	out, errOut := &bytes.Buffer{}, &bytes.Buffer{}
	err := Apply(client, []Resource{testGateway(OperationDelete, "gw1")}, false, false, out, errOut)
	require.NoError(t, err)
	assert.Equal(t, []string{"gw1"}, client.deleted)
	assertApplyRow(t, out.String(), "gateway: gw1", "delete", "ok")
}

func TestApplyNodeUsesGatewayAddedEarlierInFile(t *testing.T) {
	client := &fakeClient{}
	out, errOut := &bytes.Buffer{}, &bytes.Buffer{}
	err := Apply(client, []Resource{
		testGateway(OperationAdd, "gw1"),
		testNode(OperationAdd, "gw1", "n1", ""),
	}, false, false, out, errOut)
	require.NoError(t, err)
	require.Len(t, client.saved, 2)
	assertApplyRow(t, out.String(), "gateway: gw1", "add", "ok")
	assertApplyRow(t, out.String(), "node: gw1.n1", "add", "ok")
}

func TestApplyAddFirmwareAndDataRepository(t *testing.T) {
	client := &fakeClient{}
	out, errOut := &bytes.Buffer{}, &bytes.Buffer{}
	err := Apply(client, []Resource{
		testFirmware(OperationAdd, "stm32-app"),
		testDataRepository(OperationAdd, "ota_stm32_ab"),
	}, false, false, out, errOut)
	require.NoError(t, err)
	require.Len(t, client.saved, 2)
	assertApplyRow(t, out.String(), "firmware: stm32-app", "add", "ok")
	assertApplyRow(t, out.String(), "data-repository: ota_stm32_ab", "add", "ok")
}

func TestApplyReplaceFirmwareKeepsID(t *testing.T) {
	client := &fakeClient{existing: map[string]string{
		"firmware/id/stm32-app": "stm32-app",
	}}
	out, errOut := &bytes.Buffer{}, &bytes.Buffer{}
	err := Apply(client, []Resource{testFirmware(OperationAdd, "stm32-app")}, true, false, out, errOut)
	require.NoError(t, err)
	assert.Equal(t, []string{"stm32-app"}, client.deleted)
	require.Len(t, client.saved, 1)
	assert.Equal(t, "stm32-app", client.saved[0].Firmware.ID)
	assertApplyRow(t, out.String(), "firmware: stm32-app", "add", "replaced")
}

func TestApplyUpdateFailsWhenParentMissing(t *testing.T) {
	client := &fakeClient{existing: map[string]string{
		"node/gw1/n1//": "node-id",
	}}
	out, errOut := &bytes.Buffer{}, &bytes.Buffer{}
	err := Apply(client, []Resource{testNode(OperationMerge, "gw1", "n1", "")}, false, false, out, errOut)
	require.Error(t, err)
	assert.Empty(t, client.saved)
	assertApplyRow(t, out.String(), "node: gw1.n1", "merge", "failed: parent gateway gw1 is not present")
}
