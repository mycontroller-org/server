package utils

import (
	"errors"
	"time"

	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"
)

// StampCreatedOn keeps an existing createdOn, or sets now for a new resource.
// Existing records with a zero createdOn stay zero.
func StampCreatedOn(incoming *time.Time, existing time.Time, isNew bool) {
	if incoming == nil || !incoming.IsZero() {
		return
	}
	if !existing.IsZero() {
		*incoming = existing
		return
	}
	if isNew {
		*incoming = time.Now()
	}
}

// StampCreatedOnLookup stamps createdOn using a GetByID result.
// lookupErr == nil preserves existing; ErrNoDocuments treats the row as new;
// any other error is returned so Save does not wipe timestamps.
func StampCreatedOnLookup(incoming *time.Time, existing time.Time, lookupErr error, isNewID bool) error {
	if incoming == nil || !incoming.IsZero() {
		return nil
	}
	if isNewID {
		*incoming = time.Now()
		return nil
	}
	if lookupErr == nil {
		if !existing.IsZero() {
			*incoming = existing
		}
		return nil
	}
	if errors.Is(lookupErr, storageTY.ErrNoDocuments) {
		*incoming = time.Now()
		return nil
	}
	return lookupErr
}
