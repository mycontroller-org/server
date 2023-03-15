package types

type VariablesEngine interface {
	Load(input map[string]interface{}) (map[string]interface{}, error)
	TemplateEngine() TemplateEngine
}

const (
	VariableTypeString            = "string"
	VariableTypeResourceByQuickID = "resource_by_quick_id"
	VariableTypeResourceByLabels  = "resource_by_labels"
	VariableTypeWebhook           = "webhook"
)
