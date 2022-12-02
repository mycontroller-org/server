package dashboard

import (
	"time"

	"github.com/mycontroller-org/server/v2/pkg/types/cmap"
)

// dashboard types
const (
	TypeDesktop = "desktop"
)

// Config for dashboard
type Config struct {
	ID          string               `json:"id" yaml:"id"`
	Type        string               `json:"type" yaml:"type"`
	Title       string               `json:"title" yaml:"title"`
	Description string               `json:"description" yaml:"description"`
	Favorite    bool                 `json:"favorite" yaml:"favorite"`
	Disabled    bool                 `json:"disabled" yaml:"disabled"`
	Labels      cmap.CustomStringMap `json:"labels" yaml:"labels"`
	Widgets     []Widget             `json:"widgets" yaml:"widgets"`
	ModifiedOn  time.Time            `json:"modifiedOn" yaml:"modifiedOn"`
}

// Widget details
type Widget struct {
	ID                string         `json:"id" yaml:"id"`
	Title             string         `json:"title" yaml:"title"`
	ShowTitle         bool           `json:"showTitle" yaml:"showTitle"`
	ScrollbarDisabled bool           `json:"scrollbarDisabled" yaml:"scrollbarDisabled"`
	Static            bool           `json:"static" yaml:"static"`
	Type              string         `json:"type" yaml:"type"`
	Layout            Layout         `json:"layout" yaml:"layout"`
	Config            cmap.CustomMap `json:"config" yaml:"config"`
}

// Layout details
type Layout struct {
	Width  int `json:"w" yaml:"w"`
	Height int `json:"h" yaml:"h"`
	X      int `json:"x" yaml:"x"`
	Y      int `json:"y" yaml:"y"`
}
