package templateutils

import (
	"bytes"
	"html/template"
	"time"

	"github.com/mycontroller-org/backend/v2/pkg/json"
	"github.com/mycontroller-org/backend/v2/pkg/version"
)

var funcMap = template.FuncMap{
	"now":     time.Now,
	"version": version.Get,
	"marshal": marshal,
	"toJson":  marshal,
}

func marshal(v interface{}) template.JS {
	a, _ := json.Marshal(v)
	return template.JS(a)
}

// Execute a template
func Execute(templateText string, data interface{}) (string, error) {
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
