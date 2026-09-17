package apply

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseResourcesGateway(t *testing.T) {
	data := []byte(`
kind: gateway
operation: add
id: mysensor
description: MySensors USB
enabled: true
provider:
  type: mysensors_v2
`)
	resources, err := ParseResources(data, "gateway.yaml")
	require.NoError(t, err)
	require.Len(t, resources, 1)
	assert.Equal(t, KindGateway, resources[0].Kind)
	assert.Equal(t, OperationAdd, resources[0].Operation)
	require.NotNil(t, resources[0].Gateway)
	assert.Equal(t, "mysensor", resources[0].Gateway.ID)
	assert.Equal(t, "MySensors USB", resources[0].Gateway.Description)
	assert.True(t, resources[0].Gateway.Enabled)
	assert.Equal(t, "mysensors_v2", resources[0].Gateway.Provider["type"])
}

func TestParseResourcesFirmwareAndDataRepository(t *testing.T) {
	data := []byte(`
kind: firmware
operation: add
id: stm32-app
description: Slot A image
labels:
  ms_flash_slot: A
---
kind: data-repository
operation: add
id: ota_stm32_ab
description: STM32 A/B OTA policy
readOnly: true
data:
  disabled: false
`)
	resources, err := ParseResources(data, "fw.yaml")
	require.NoError(t, err)
	require.Len(t, resources, 2)
	assert.Equal(t, KindFirmware, resources[0].Kind)
	assert.Equal(t, "stm32-app", resources[0].Firmware.ID)
	assert.Equal(t, "Slot A image", resources[0].Firmware.Description)
	assert.Equal(t, "A", resources[0].Firmware.Labels["ms_flash_slot"])
	assert.Equal(t, KindDataRepository, resources[1].Kind)
	assert.Equal(t, "ota_stm32_ab", resources[1].DataRepository.ID)
	assert.True(t, resources[1].DataRepository.ReadOnly)
	assert.Equal(t, false, resources[1].DataRepository.Data["disabled"])
}

func TestParseResourcesUserAndPolicy(t *testing.T) {
	data := []byte(`
kind: user
operation: add
username: alice
password: secret
email: alice@example.com
policies:
  - admin
---
kind: policy
operation: add
id: sensors-read
description: read sensors
statements:
  - effect: Allow
    actions: ["get", "list"]
    resources: ["node:*", "field:*"]
`)
	resources, err := ParseResources(data, "acl.yaml")
	require.NoError(t, err)
	require.Len(t, resources, 2)
	assert.Equal(t, KindUser, resources[0].Kind)
	assert.Equal(t, "alice", resources[0].User.Username)
	assert.Equal(t, "secret", resources[0].User.Password)
	assert.Equal(t, []string{"admin"}, resources[0].User.Policies)
	assert.True(t, resources[0].User.Enabled)
	assert.Equal(t, KindPolicy, resources[1].Kind)
	assert.Equal(t, "sensors-read", resources[1].Policy.ID)
	require.Len(t, resources[1].Policy.Statements, 1)
	assert.Equal(t, "Allow", resources[1].Policy.Statements[0].Effect)
}

func TestParseResourcesServiceAccount(t *testing.T) {
	data := []byte(`
kind: sa
operation: add
name: ci-bot
description: CI automation
neverExpire: true
statements:
  - effect: Allow
    actions: ["get", "list"]
    resources: ["node:*"]
`)
	resources, err := ParseResources(data, "sa.yaml")
	require.NoError(t, err)
	require.Len(t, resources, 1)
	assert.Equal(t, KindServiceAccount, resources[0].Kind)
	require.NotNil(t, resources[0].ServiceAccount)
	assert.Equal(t, "ci-bot", resources[0].ServiceAccount.Name)
	assert.Equal(t, "CI automation", resources[0].ServiceAccount.Description)
	assert.True(t, resources[0].ServiceAccount.NeverExpire)
	assert.True(t, resources[0].ServiceAccount.Enabled)
	require.Len(t, resources[0].ServiceAccount.Statements, 1)
	assert.Equal(t, "Allow", resources[0].ServiceAccount.Statements[0].Effect)
	assert.Equal(t, []string{"get", "list"}, resources[0].ServiceAccount.Statements[0].Actions)
	assert.Equal(t, []string{"node:*"}, resources[0].ServiceAccount.Statements[0].Resources)
	assert.Equal(t, "service-account: ci-bot", resources[0].TableResource())
}

func TestParseResourcesServiceAccountForUser(t *testing.T) {
	data := []byte(`
kind: service-account
operation: add
name: mobile
username: alice
neverExpire: true
`)
	resources, err := ParseResources(data, "sa.yaml")
	require.NoError(t, err)
	require.Len(t, resources, 1)
	assert.Equal(t, "alice", resources[0].ServiceAccount.Username)
	assert.Equal(t, "mobile", resources[0].ServiceAccount.Name)
	assert.True(t, resources[0].ServiceAccount.Enabled)
	assert.Equal(t, "service-account: alice.mobile", resources[0].TableResource())
}

func TestParseResourcesUserEnabledFalse(t *testing.T) {
	data := []byte(`
kind: user
operation: add
username: bob
password: secret
enabled: false
`)
	resources, err := ParseResources(data, "user.yaml")
	require.NoError(t, err)
	require.Len(t, resources, 1)
	assert.False(t, resources[0].User.Enabled)
}

func TestParseResourcesYAMLSingle(t *testing.T) {
	data := []byte(`
kind: node
operation: add
gatewayId: gw1
nodeId: n1
name: Living Room
labels:
  room: living
`)
	resources, err := ParseResources(data, "node.yaml")
	require.NoError(t, err)
	require.Len(t, resources, 1)
	assert.Equal(t, KindNode, resources[0].Kind)
	assert.Equal(t, OperationAdd, resources[0].Operation)
	require.NotNil(t, resources[0].Node)
	assert.Equal(t, "gw1", resources[0].Node.GatewayID)
	assert.Equal(t, "n1", resources[0].Node.NodeID)
	assert.Equal(t, "Living Room", resources[0].Node.Name)
	assert.Equal(t, "living", resources[0].Node.Labels["room"])
}

func TestParseResourcesYAMLMultiDoc(t *testing.T) {
	data := []byte(`
kind: node
operation: add
gatewayId: gw1
nodeId: n1
name: Node 1
---
kind: source
operation: update
gatewayId: gw1
nodeId: n1
sourceId: s1
name: Temperature
---
kind: field
operation: delete
gatewayId: gw1
nodeId: n1
sourceId: s1
fieldId: temp
`)
	resources, err := ParseResources(data, "mixed.yaml")
	require.NoError(t, err)
	require.Len(t, resources, 3)
	assert.Equal(t, KindNode, resources[0].Kind)
	assert.Equal(t, OperationAdd, resources[0].Operation)
	assert.Equal(t, KindSource, resources[1].Kind)
	assert.Equal(t, OperationMerge, resources[1].Operation)
	assert.Equal(t, "s1", resources[1].Src.SourceID)
	assert.Equal(t, KindField, resources[2].Kind)
	assert.Equal(t, OperationDelete, resources[2].Operation)
	assert.Equal(t, "temp", resources[2].Field.FieldID)
}

func TestParseResourcesYAMLList(t *testing.T) {
	data := []byte(`
- kind: node
  operation: create
  gatewayId: gw1
  nodeId: "1"
  name: Node 1
- kind: sources
  operation: add
  gatewayId: gw1
  nodeId: "1"
  sourceId: s1
  name: Source 1
`)
	resources, err := ParseResources(data, "list.yaml")
	require.NoError(t, err)
	require.Len(t, resources, 2)
	assert.Equal(t, KindNode, resources[0].Kind)
	assert.Equal(t, OperationAdd, resources[0].Operation)
	assert.Equal(t, "1", resources[0].Node.NodeID)
	assert.Equal(t, KindSource, resources[1].Kind)
	assert.Equal(t, "s1", resources[1].Src.SourceID)
}

func TestParseResourcesJSONArray(t *testing.T) {
	data := []byte(`[
  {"kind":"node","operation":"add","gatewayId":"gw1","nodeId":"n1","name":"N1"},
  {"kind":"field","operation":"remove","id":"field-1"}
]`)
	resources, err := ParseResources(data, "resources.json")
	require.NoError(t, err)
	require.Len(t, resources, 2)
	assert.Equal(t, KindNode, resources[0].Kind)
	assert.Equal(t, KindField, resources[1].Kind)
	assert.Equal(t, OperationDelete, resources[1].Operation)
	assert.Equal(t, "field-1", resources[1].Field.ID)
}

func TestParseResourcesJSONObject(t *testing.T) {
	data := []byte(`{"kind":"source","operation":"update","gatewayId":"gw1","nodeId":"n1","sourceId":"s1","name":"S1"}`)
	resources, err := ParseResources(data, "source.json")
	require.NoError(t, err)
	require.Len(t, resources, 1)
	assert.Equal(t, KindSource, resources[0].Kind)
	assert.Equal(t, OperationMerge, resources[0].Operation)
	assert.Equal(t, "S1", resources[0].Src.Name)
}

func TestParseResourcesValidation(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr string
	}{
		{
			name:    "missing kind",
			input:   "operation: add\ngatewayId: gw1\nnodeId: n1\n",
			wantErr: "kind is required",
		},
		{
			name:    "missing operation",
			input:   "kind: node\ngatewayId: gw1\nnodeId: n1\n",
			wantErr: "operation is required",
		},
		{
			name:    "unsupported kind",
			input:   "kind: task\noperation: add\nid: t1\n",
			wantErr: "unsupported kind",
		},
		{
			name:    "gateway missing id",
			input:   "kind: gateway\noperation: add\ndescription: usb\n",
			wantErr: "gateway requires id",
		},
		{
			name:    "firmware missing id",
			input:   "kind: firmware\noperation: add\ndescription: img\n",
			wantErr: "firmware requires id",
		},
		{
			name:    "data-repository missing id",
			input:   "kind: data-repo\noperation: add\ndescription: repo\n",
			wantErr: "data-repository requires id",
		},
		{
			name:    "service-account missing name",
			input:   "kind: service-account\noperation: add\ndescription: ci\n",
			wantErr: "service-account requires name",
		},
		{
			name:    "unsupported operation",
			input:   "kind: node\noperation: patch\ngatewayId: gw1\nnodeId: n1\n",
			wantErr: "unsupported operation",
		},
		{
			name:    "node missing keys",
			input:   "kind: node\noperation: add\nname: only-name\n",
			wantErr: "node requires gatewayId and nodeId",
		},
		{
			name:    "empty file",
			input:   "   \n",
			wantErr: "no resources found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseResources([]byte(tt.input), "test.yaml")
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestParseResourcesNumericIDs(t *testing.T) {
	data := []byte(`
kind: field
operation: add
gatewayId: gw1
nodeId: 1
sourceId: 2
fieldId: 3
name: Temperature
`)
	resources, err := ParseResources(data, "numeric.yaml")
	require.NoError(t, err)
	require.Len(t, resources, 1)
	assert.Equal(t, "1", resources[0].Field.NodeID)
	assert.Equal(t, "2", resources[0].Field.SourceID)
	assert.Equal(t, "3", resources[0].Field.FieldID)
}

func TestParseResourcesItemsList(t *testing.T) {
	data := []byte(`
kind: source
operation: add
replace: true
items:
  - gatewayId: mysensor
    nodeId: "1"
    sourceId: dht
    fieldId: temperature
    name: Temperature
    metricType: gauge
    unit: °C
    labels:
      location: living-room
  - gatewayId: mysensor
    nodeId: "1"
    sourceId: dht
    fieldId: humidity
    name: Humidity
    unit: "%"
`)
	resources, err := ParseResources(data, "items.yaml")
	require.NoError(t, err)
	require.Len(t, resources, 2)

	assert.Equal(t, KindField, resources[0].Kind)
	assert.Equal(t, OperationAdd, resources[0].Operation)
	assert.True(t, resources[0].Replace)
	require.NotNil(t, resources[0].Field)
	assert.Equal(t, "mysensor", resources[0].Field.GatewayID)
	assert.Equal(t, "1", resources[0].Field.NodeID)
	assert.Equal(t, "dht", resources[0].Field.SourceID)
	assert.Equal(t, "temperature", resources[0].Field.FieldID)
	assert.Equal(t, "Temperature", resources[0].Field.Name)
	assert.Equal(t, "gauge", resources[0].Field.MetricType)
	assert.Equal(t, "°C", resources[0].Field.Unit)
	assert.Equal(t, "living-room", resources[0].Field.Labels["location"])

	assert.Equal(t, KindField, resources[1].Kind)
	assert.Equal(t, "humidity", resources[1].Field.FieldID)
	assert.True(t, resources[1].Replace)
}

func TestParseResourcesItemsInheritDefaults(t *testing.T) {
	data := []byte(`
kind: field
operation: add
gatewayId: mysensor
nodeId: "1"
sourceId: dht
items:
  - fieldId: temperature
    name: Temperature
  - fieldId: humidity
    name: Humidity
`)
	resources, err := ParseResources(data, "defaults.yaml")
	require.NoError(t, err)
	require.Len(t, resources, 2)
	assert.Equal(t, "mysensor", resources[0].Field.GatewayID)
	assert.Equal(t, "dht", resources[0].Field.SourceID)
	assert.Equal(t, "temperature", resources[0].Field.FieldID)
	assert.Equal(t, "humidity", resources[1].Field.FieldID)
	assert.Equal(t, "Humidity", resources[1].Field.Name)
}

func TestParseResourcesItemsSourceWithoutFieldID(t *testing.T) {
	data := []byte(`
kind: source
operation: add
items:
  - gatewayId: gw1
    nodeId: n1
    sourceId: s1
    name: DHT
`)
	resources, err := ParseResources(data, "sources.yaml")
	require.NoError(t, err)
	require.Len(t, resources, 1)
	assert.Equal(t, KindSource, resources[0].Kind)
	assert.False(t, resources[0].Replace)
	assert.Equal(t, "s1", resources[0].Src.SourceID)
	assert.Equal(t, "DHT", resources[0].Src.Name)
}

func TestParseResourcesItemsEmpty(t *testing.T) {
	_, err := ParseResources([]byte("kind: source\noperation: add\nitems: []\n"), "empty.yaml")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "items must not be empty")
}

func TestResourceIdentity(t *testing.T) {
	resources, err := ParseResources([]byte(`
kind: field
operation: add
id: abc
gatewayId: gw1
nodeId: n1
sourceId: s1
fieldId: temp
`), "id.yaml")
	require.NoError(t, err)
	assert.True(t, strings.Contains(resources[0].Identity(), "gw1.n1.s1.temp"))
	assert.True(t, strings.Contains(resources[0].Identity(), "id=abc"))
}
