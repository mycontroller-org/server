package types

type VariablesEngine interface {
	Load(input map[string]string) (map[string]interface{}, error)
	TemplateEngine() TemplateEngine
}
