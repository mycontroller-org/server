package utils

import (
	"errors"
	"testing"
	"time"

	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"
)

func TestStampCreatedOnLookup(t *testing.T) {
	existing := time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC)

	t.Run("preserve existing", func(t *testing.T) {
		got := time.Time{}
		if err := StampCreatedOnLookup(&got, existing, nil, false); err != nil {
			t.Fatal(err)
		}
		if !got.Equal(existing) {
			t.Fatalf("got %v", got)
		}
	})

	t.Run("new id", func(t *testing.T) {
		got := time.Time{}
		if err := StampCreatedOnLookup(&got, time.Time{}, nil, true); err != nil {
			t.Fatal(err)
		}
		if got.IsZero() {
			t.Fatal("expected now")
		}
	})

	t.Run("missing document is create", func(t *testing.T) {
		got := time.Time{}
		if err := StampCreatedOnLookup(&got, time.Time{}, storageTY.ErrNoDocuments, false); err != nil {
			t.Fatal(err)
		}
		if got.IsZero() {
			t.Fatal("expected now")
		}
	})

	t.Run("lookup error is returned", func(t *testing.T) {
		got := time.Time{}
		boom := errors.New("storage down")
		if err := StampCreatedOnLookup(&got, existing, boom, false); !errors.Is(err, boom) {
			t.Fatalf("got %v", err)
		}
		if !got.IsZero() {
			t.Fatal("should not stamp on lookup error")
		}
	})

	t.Run("keep incoming", func(t *testing.T) {
		got := existing
		if err := StampCreatedOnLookup(&got, time.Time{}, nil, true); err != nil {
			t.Fatal(err)
		}
		if !got.Equal(existing) {
			t.Fatalf("overwrote incoming: %v", got)
		}
	})
}
