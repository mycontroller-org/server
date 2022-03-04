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
	ID          string               `json:"id"`
	Type        string               `json:"type"`
	Title       string               `json:"title"`
	Description string               `json:"description"`
	Favorite    bool                 `json:"favorite"`
	Disabled    bool                 `json:"disabled"`
	Labels      cmap.CustomStringMap `json:"labels"`
	Widgets     []Widget             `json:"widgets"`
	ModifiedOn  time.Time            `json:"modifiedOn"`
}

// Widget details
type Widget struct {
	ID                string         `json:"id"`
	Title             string         `json:"title"`
	ShowTitle         bool           `json:"showTitle"`
	ScrollbarDisabled bool           `json:"scrollbarDisabled"`
	Static            bool           `json:"static"`
	Type              string         `json:"type"`
	Layout            Layout         `json:"layout"`
	Config            cmap.CustomMap `json:"config"`
}

// Layout details
type Layout struct {
	Width  int `json:"w"`
	Height int `json:"h"`
	X      int `json:"x"`
	Y      int `json:"y"`
}
