package gateway

import (
	"time"

	"github.com/mycontroller-org/backend/v2/pkg/model"
	"github.com/mycontroller-org/backend/v2/pkg/model/cmap"
)

// Config struct
type Config struct {
	ID             string               `json:"id"`
	Name           string               `json:"name"`
	Enabled        bool                 `json:"enabled"`
	Provider       cmap.CustomMap       `json:"provider"`
	MessageLogger  cmap.CustomMap       `json:"messageLogger"`
	Labels         cmap.CustomStringMap `json:"labels"`
	Others         cmap.CustomMap       `json:"others"`
	State          *model.State         `json:"state"`
	LastModifiedOn time.Time            `json:"lastModifiedOn"`
}
