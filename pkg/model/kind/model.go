package kind

import (
	"github.com/mycontroller-org/backend/v2/pkg/model/cmap"
)

// kind types
const (
	TypeUser     = "user"
	TypeExporter = "exporter"
)

// Kind model
type Kind struct {
	ID     string               `json:"id"`
	Type   string               `json:"type"`
	Labels cmap.CustomStringMap `json:"labels"`
	Spec   cmap.CustomMap       `json:"spec"`
}

// Clone kind resource
func (k *Kind) Clone() Kind {
	newKind := Kind{
		ID:     k.ID,
		Type:   k.Type,
		Labels: k.Labels.Clone(),
		Spec:   k.Spec.Clone(),
	}
	return newKind
}
