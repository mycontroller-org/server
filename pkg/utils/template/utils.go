package templateutils

import (
	"bytes"
	"html/template"
	"time"

	"github.com/mycontroller-org/backend/v2/pkg/version"
	"go.uber.org/zap"
)

var funcMap = template.FuncMap{
	"now":     time.Now,
	"version": version.Get,
}

// Execute a template
func Execute(templateText string, data interface{}) (string, error) {
	zap.L().Debug("template", zap.String("tpl", templateText))
	tpl, err := template.New("base").Funcs(funcMap).Parse(templateText)
	if err != nil {
		return "", err
	}
	var tplBuffer bytes.Buffer
	err = tpl.Execute(&tplBuffer, data)
	if err != nil {
		return "", err
	}
	return tplBuffer.String(), nil
}
