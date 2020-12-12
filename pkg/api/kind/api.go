package kind

import (
	ml "github.com/mycontroller-org/backend/v2/pkg/model"
	kml "github.com/mycontroller-org/backend/v2/pkg/model/kind"
	pml "github.com/mycontroller-org/backend/v2/pkg/model/pagination"
	stgml "github.com/mycontroller-org/backend/v2/pkg/model/storage"
	svc "github.com/mycontroller-org/backend/v2/pkg/service"
	ut "github.com/mycontroller-org/backend/v2/pkg/utils"
)

// List by filter and pagination
func List(f []pml.Filter, p *pml.Pagination) (*pml.Result, error) {
	out := make([]kml.Kind, 0)
	return svc.STG.Find(ml.EntityKind, &out, f, p)
}

// Get a kind
func Get(f []pml.Filter) (kml.Kind, error) {
	out := kml.Kind{}
	err := svc.STG.FindOne(ml.EntityKind, &out, f)
	return out, err
}

// Save kind config
func Save(node *kml.Kind) error {
	if node.ID == "" {
		node.ID = ut.RandUUID()
	}
	f := []pml.Filter{
		{Key: ml.KeyID, Value: node.ID},
	}
	return svc.STG.Upsert(ml.EntityKind, node, f)
}

// GetByTypeName returns a kind by type and name
func GetByTypeName(kindType, name string) (*kml.Kind, error) {
	f := []pml.Filter{
		{Key: ml.KeyKindType, Value: kindType},
		{Key: ml.KeyKindName, Value: name},
	}
	out := &kml.Kind{}
	err := svc.STG.FindOne(ml.EntityKind, out, f)
	return out, err
}

// Delete kind
func Delete(IDs []string) (int64, error) {
	f := []pml.Filter{{Key: ml.KeyID, Operator: stgml.OperatorIn, Value: IDs}}
	return svc.STG.Delete(ml.EntityKind, f)
}
