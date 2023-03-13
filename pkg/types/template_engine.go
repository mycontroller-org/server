package types

type TemplateEngine interface {
	Execute(templateText string, data interface{}) (string, error)
}
