package dashboard

import "github.com/mycontroller-org/backend/v2/pkg/model/cmap"

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
}

// Widget details
type Widget struct {
	ID        string         `json:"id"`
	Title     string         `json:"title"`
	ShowTitle bool           `json:"showTitle"`
	Type      string         `json:"type"`
	Static    bool           `json:"static"`
	Layout    Layout         `json:"layout"`
	Config    cmap.CustomMap `json:"config"`
}

// Layout details
type Layout struct {
	Width  int `json:"w"`
	Height int `json:"h"`
	X      int `json:"x"`
	Y      int `json:"y"`
}
