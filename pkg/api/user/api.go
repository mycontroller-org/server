package user

import (
	ml "github.com/mycontroller-org/backend/v2/pkg/model"
	uml "github.com/mycontroller-org/backend/v2/pkg/model/user"
	svc "github.com/mycontroller-org/backend/v2/pkg/service"
	ut "github.com/mycontroller-org/backend/v2/pkg/utils"
	stgml "github.com/mycontroller-org/backend/v2/plugin/storage"
)

// List by filter and pagination
func List(filters []stgml.Filter, pagination *stgml.Pagination) (*stgml.Result, error) {
	result := make([]uml.User, 0)
	return svc.STG.Find(ml.EntityUser, &result, filters, pagination)
}

// Get returns a item
func Get(filters []stgml.Filter) (uml.User, error) {
	result := uml.User{}
	err := svc.STG.FindOne(ml.EntityUser, &result, filters)
	return result, err
}

// GetByID returns a item
func GetByID(ID string) (uml.User, error) {
	result := uml.User{}
	filters := []stgml.Filter{
		{Key: ml.KeyID, Value: ID},
	}
	err := svc.STG.FindOne(ml.EntityUser, &result, filters)
	return result, err
}

// GetByUsername returns a item
func GetByUsername(username string) (uml.User, error) {
	result := uml.User{}
	filters := []stgml.Filter{
		{Key: ml.KeyUsername, Value: username},
	}
	err := svc.STG.FindOne(ml.EntityUser, &result, filters)
	return result, err
}

// GetByEmail returns a item
func GetByEmail(email string) (uml.User, error) {
	result := uml.User{}
	filters := []stgml.Filter{
		{Key: ml.KeyEmail, Value: email},
	}
	err := svc.STG.FindOne(ml.EntityUser, &result, filters)
	return result, err
}

// Save config into disk
func Save(user *uml.User) error {
	if user.ID == "" {
		user.ID = ut.RandUUID()
	}
	filters := []stgml.Filter{
		{Key: ml.KeyID, Value: user.ID},
	}
	return svc.STG.Upsert(ml.EntityUser, user, filters)
}

// Delete items
func Delete(IDs []string) (int64, error) {
	filters := []stgml.Filter{{Key: ml.KeyID, Operator: stgml.OperatorIn, Value: IDs}}
	return svc.STG.Delete(ml.EntityUser, filters)
}
