package api

import (
	"fmt"
	"strings"

	"github.com/mycontroller-org/server/v2/pkg/json"
	dataRepoTY "github.com/mycontroller-org/server/v2/pkg/types/data_repository"
	fieldTY "github.com/mycontroller-org/server/v2/pkg/types/field"
	firmwareTY "github.com/mycontroller-org/server/v2/pkg/types/firmware"
	nodeTY "github.com/mycontroller-org/server/v2/pkg/types/node"
	sourceTY "github.com/mycontroller-org/server/v2/pkg/types/source"
	gwTY "github.com/mycontroller-org/server/v2/plugin/gateway/types"
	"github.com/tidwall/sjson"
)

// SetResourcePath loads a resource, sets a dotted JSON path, and saves it.
func (c *Client) SetResourcePath(kind, selector, keyPath, value string, rawText bool) error {
	if strings.TrimSpace(keyPath) == "" {
		return fmt.Errorf("key path is required")
	}
	if strings.TrimSpace(selector) == "" {
		return fmt.Errorf("resource id is required")
	}

	switch kind {
	case "gateway":
		item, err := c.FindGateway(selector)
		if err != nil {
			return err
		}
		if item == nil {
			return fmt.Errorf("gateway %s is not present", selector)
		}
		updated := &gwTY.Config{}
		if err := applyJSONPath(item, keyPath, value, rawText, updated); err != nil {
			return err
		}
		return c.SaveGateway(updated)
	case "node":
		item, err := c.findNodeSelector(selector)
		if err != nil {
			return err
		}
		if item == nil {
			return fmt.Errorf("node %s is not present", selector)
		}
		updated := &nodeTY.Node{}
		if err := applyJSONPath(item, keyPath, value, rawText, updated); err != nil {
			return err
		}
		return c.SaveNode(updated)
	case "source":
		item, err := c.findSourceSelector(selector)
		if err != nil {
			return err
		}
		if item == nil {
			return fmt.Errorf("source %s is not present", selector)
		}
		updated := &sourceTY.Source{}
		if err := applyJSONPath(item, keyPath, value, rawText, updated); err != nil {
			return err
		}
		return c.SaveSource(updated)
	case "field":
		item, err := c.findFieldSelector(selector)
		if err != nil {
			return err
		}
		if item == nil {
			return fmt.Errorf("field %s is not present", selector)
		}
		updated := &fieldTY.Field{}
		if err := applyJSONPath(item, keyPath, value, rawText, updated); err != nil {
			return err
		}
		return c.SaveField(updated)
	case "firmware":
		item, err := c.FindFirmware(selector)
		if err != nil {
			return err
		}
		if item == nil {
			return fmt.Errorf("firmware %s is not present", selector)
		}
		updated := &firmwareTY.Firmware{}
		if err := applyJSONPath(item, keyPath, value, rawText, updated); err != nil {
			return err
		}
		return c.SaveFirmware(updated)
	case "data-repository":
		item, err := c.FindDataRepository(selector)
		if err != nil {
			return err
		}
		if item == nil {
			return fmt.Errorf("data-repository %s is not present", selector)
		}
		updated := &dataRepoTY.Config{}
		if err := applyJSONPath(item, keyPath, value, rawText, updated); err != nil {
			return err
		}
		return c.SaveDataRepository(updated)
	default:
		return fmt.Errorf("unsupported kind %q", kind)
	}
}

func (c *Client) ResolveGatewayIDs(selectors []string, requireEnabled bool) ([]string, error) {
	ids := make([]string, 0, len(selectors))
	missing := make([]string, 0)
	disabled := make([]string, 0)
	for _, selector := range selectors {
		selector = strings.TrimPrefix(selector, "gateway:")
		item, err := c.FindGateway(selector)
		if err != nil {
			return ids, err
		}
		if item == nil {
			missing = append(missing, selector)
			continue
		}
		if requireEnabled && !item.Enabled {
			disabled = append(disabled, selector)
			continue
		}
		ids = append(ids, item.ID)
	}
	parts := make([]string, 0, 2)
	if len(missing) > 0 {
		parts = append(parts, "gateway(s) not present: "+strings.Join(missing, ", "))
	}
	if len(disabled) > 0 {
		parts = append(parts, "gateway(s) disabled: "+strings.Join(disabled, ", "))
	}
	if len(parts) > 0 {
		return ids, fmt.Errorf("%s", strings.Join(parts, "; "))
	}
	return ids, nil
}

func (c *Client) ResolveNodeIDs(selectors []string) ([]string, error) {
	ids := make([]string, 0, len(selectors))
	missing := make([]string, 0)
	for _, selector := range selectors {
		item, err := c.findNodeSelector(selector)
		if err != nil {
			return ids, err
		}
		if item == nil {
			missing = append(missing, selector)
			continue
		}
		ids = append(ids, item.ID)
	}
	if len(missing) > 0 {
		return ids, fmt.Errorf("node(s) not present: %s", strings.Join(missing, ", "))
	}
	return ids, nil
}

func (c *Client) findNodeSelector(selector string) (*nodeTY.Node, error) {
	selector = strings.TrimPrefix(selector, "node:")
	item, err := c.FindNode(selector, "", "")
	if err != nil || item != nil {
		return item, err
	}
	parts := strings.SplitN(selector, ".", 2)
	if len(parts) == 2 {
		return c.FindNode("", parts[0], parts[1])
	}
	return nil, nil
}

func (c *Client) findSourceSelector(selector string) (*sourceTY.Source, error) {
	item, err := c.FindSource(selector, "", "", "")
	if err != nil || item != nil {
		return item, err
	}
	parts := strings.Split(selector, ".")
	if len(parts) == 3 {
		return c.FindSource("", parts[0], parts[1], parts[2])
	}
	return nil, nil
}

func (c *Client) findFieldSelector(selector string) (*fieldTY.Field, error) {
	item, err := c.FindField(selector, "", "", "", "")
	if err != nil || item != nil {
		return item, err
	}
	parts := strings.Split(selector, ".")
	if len(parts) == 4 {
		return c.FindField("", parts[0], parts[1], parts[2], parts[3])
	}
	return nil, nil
}

func applyJSONPath(in interface{}, keyPath, value string, rawText bool, out interface{}) error {
	raw, err := json.Marshal(in)
	if err != nil {
		return err
	}
	updated, err := sjson.Set(string(raw), keyPath, decodeSetValue(value, rawText))
	if err != nil {
		return fmt.Errorf("invalid key path %q: %w", keyPath, err)
	}
	if err := json.Unmarshal([]byte(updated), out); err != nil {
		return fmt.Errorf("failed to apply key path %q: %w", keyPath, err)
	}
	return nil
}

func decodeSetValue(value string, rawText bool) interface{} {
	if rawText {
		return value
	}
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return value
	}
	var decoded interface{}
	if err := json.Unmarshal([]byte(trimmed), &decoded); err != nil {
		return value
	}
	return decoded
}
