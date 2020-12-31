package user

import (
	ml "github.com/mycontroller-org/backend/v2/pkg/model"
	uml "github.com/mycontroller-org/backend/v2/pkg/model/user"
	stg "github.com/mycontroller-org/backend/v2/pkg/service/storage"
	ut "github.com/mycontroller-org/backend/v2/pkg/utils"
	stgml "github.com/mycontroller-org/backend/v2/plugin/storage"
)

// List by filter and pagination
func List(filters []stgml.Filter, pagination *stgml.Pagination) (*stgml.Result, error) {
	result := make([]uml.User, 0)
	return stg.SVC.Find(ml.EntityUser, &result, filters, pagination)
}

// Get returns a item
func Get(filters []stgml.Filter) (uml.User, error) {
	result := uml.User{}
	err := stg.SVC.FindOne(ml.EntityUser, &result, filters)
	return result, err
}

// GetByID returns a item
func GetByID(ID string) (uml.User, error) {
	result := uml.User{}
	filters := []stgml.Filter{
		{Key: ml.KeyID, Value: ID},
	}
	err := stg.SVC.FindOne(ml.EntityUser, &result, filters)
	return result, err
}

// GetByUsername returns a item
func GetByUsername(username string) (uml.User, error) {
	result := uml.User{}
	filters := []stgml.Filter{
		{Key: ml.KeyUsername, Value: username},
	}
	err := stg.SVC.FindOne(ml.EntityUser, &result, filters)
	return result, err
}

// GetByEmail returns a item
func GetByEmail(email string) (uml.User, error) {
	result := uml.User{}
	filters := []stgml.Filter{
		{Key: ml.KeyEmail, Value: email},
	}
	err := stg.SVC.FindOne(ml.EntityUser, &result, filters)
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
	return stg.SVC.Upsert(ml.EntityUser, user, filters)
}

// Delete items
func Delete(IDs []string) (int64, error) {
	filters := []stgml.Filter{{Key: ml.KeyID, Operator: stgml.OperatorIn, Value: IDs}}
	return stg.SVC.Delete(ml.EntityUser, filters)
}
