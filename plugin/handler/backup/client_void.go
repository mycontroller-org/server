// +build !core,standalone
// added two tags to ignore vscode editor warning

package exporter

import (
	"errors"

	"github.com/mycontroller-org/backend/v2/pkg/model"
)

// Client for backup service
type Client interface {
	Start() error
	Post(variables map[string]interface{}) error
	Close() error
	State() *model.State
}

// Init backup client
func Init(cfg interface{}) (Client, error) {
	return nil, errors.New("backup handler will be available only with core or all-in-one package")
}
