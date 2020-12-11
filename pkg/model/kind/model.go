package kind

import (
	"time"

	"github.com/mycontroller-org/backend/v2/pkg/model/cmap"
)

// kind types
const (
	TypeExporter = "exporter"
)

// Kind model
type Kind struct {
	ID             string               `json:"id"`
	Type           string               `json:"type"`
	Name           string               `json:"name"`
	Labels         cmap.CustomStringMap `json:"labels"`
	Spec           cmap.CustomMap       `json:"spec"`
	LastModifiedOn time.Time            `json:"lastModifiedOn"`
}

// Clone kind resource
func (k *Kind) Clone() Kind {
	newKind := Kind{
		ID:             k.ID,
		Type:           k.Type,
		Name:           k.Name,
		Labels:         k.Labels.Clone(),
		Spec:           k.Spec.Clone(),
		LastModifiedOn: k.LastModifiedOn,
	}
	return newKind
}
