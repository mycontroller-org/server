package api

import (
	"fmt"
	"net/http"

	"github.com/mycontroller-org/server/v2/pkg/json"
	settingsTY "github.com/mycontroller-org/server/v2/pkg/types/settings"
)

func (c *Client) GetSystemSettings() (*settingsTY.Settings, error) {
	res, err := c.executeJson(API_SETTINGS_SYSTEM, http.MethodGet, nil, nil, nil, http.StatusOK)
	if err != nil {
		return nil, err
	}
	item := &settingsTY.Settings{}
	if err := json.Unmarshal(res.Body, item); err != nil {
		return nil, err
	}
	return item, nil
}

func (c *Client) UpdateSettings(settings *settingsTY.Settings) error {
	_, err := c.executeJson(API_SETTINGS, http.MethodPost, nil, nil, settings, http.StatusOK)
	return err
}

func (c *Client) SetSystemSettingsPath(keyPath, value string, rawText bool) error {
	settings, err := c.GetSystemSettings()
	if err != nil {
		return err
	}
	if settings.Spec == nil {
		settings.Spec = map[string]interface{}{}
	}
	updated := map[string]interface{}{}
	if err := applyJSONPath(settings.Spec, keyPath, value, rawText, &updated); err != nil {
		return err
	}
	settings.ID = settingsTY.KeySystemSettings
	settings.Spec = updated
	return c.UpdateSettings(settings)
}

func (c *Client) MergeSystemSettings(overlay map[string]interface{}) error {
	if len(overlay) == 0 {
		return fmt.Errorf("settings overlay is empty")
	}
	settings, err := c.GetSystemSettings()
	if err != nil {
		return err
	}
	if spec, ok := overlay["spec"].(map[string]interface{}); ok && overlay["id"] != nil {
		overlay = spec
	}
	settings.ID = settingsTY.KeySystemSettings
	settings.Spec = mergeSettingMaps(settings.Spec, overlay)
	return c.UpdateSettings(settings)
}

func mergeSettingMaps(base, overlay map[string]interface{}) map[string]interface{} {
	if base == nil {
		base = map[string]interface{}{}
	}
	out := make(map[string]interface{}, len(base)+len(overlay))
	for key, value := range base {
		out[key] = value
	}
	for key, value := range overlay {
		existing, ok := out[key]
		if !ok {
			out[key] = value
			continue
		}
		existingMap, existingIsMap := existing.(map[string]interface{})
		valueMap, valueIsMap := value.(map[string]interface{})
		if existingIsMap && valueIsMap {
			out[key] = mergeSettingMaps(existingMap, valueMap)
			continue
		}
		out[key] = value
	}
	return out
}
