package handler

import (
	resource "github.com/mycontroller-org/server/v2/plugin/handler/resource"
)

const (
	PluginResourceHandler = "resource"
)

func init() {
	Register(resource.PluginResourceHandler, resource.NewResourcePlugin)
}
