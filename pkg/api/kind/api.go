package kind

import (
	ml "github.com/mycontroller-org/backend/v2/pkg/model"
	kml "github.com/mycontroller-org/backend/v2/pkg/model/kind"
	svc "github.com/mycontroller-org/backend/v2/pkg/service"
	ut "github.com/mycontroller-org/backend/v2/pkg/utils"
	stgml "github.com/mycontroller-org/backend/v2/plugin/storage"
)

// List by filter and pagination
func List(f []stgml.Filter, p *stgml.Pagination) (*stgml.Result, error) {
	out := make([]kml.Kind, 0)
	return svc.STG.Find(ml.EntityKind, &out, f, p)
}

// Get a kind
func Get(f []stgml.Filter) (kml.Kind, error) {
	out := kml.Kind{}
	err := svc.STG.FindOne(ml.EntityKind, &out, f)
	return out, err
}

// Save kind config
func Save(node *kml.Kind) error {
	if node.ID == "" {
		node.ID = ut.RandUUID()
	}
	f := []stgml.Filter{
		{Key: ml.KeyID, Value: node.ID},
	}
	return svc.STG.Upsert(ml.EntityKind, node, f)
}

// GetByTypeName returns a kind by type and name
func GetByTypeName(kindType, name string) (*kml.Kind, error) {
	f := []stgml.Filter{
		{Key: ml.KeyKindType, Value: kindType},
		{Key: ml.KeyKindName, Value: name},
	}
	out := &kml.Kind{}
	err := svc.STG.FindOne(ml.EntityKind, out, f)
	return out, err
}

// Delete kind
func Delete(IDs []string) (int64, error) {
	f := []stgml.Filter{{Key: ml.KeyID, Operator: stgml.OperatorIn, Value: IDs}}
	return svc.STG.Delete(ml.EntityKind, f)
}
