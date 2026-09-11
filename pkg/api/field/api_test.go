package field

import (
	"context"
	"testing"

	fieldTY "github.com/mycontroller-org/server/v2/pkg/types/field"
	busTY "github.com/mycontroller-org/server/v2/plugin/bus/types"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type stubStorage struct {
	storageTY.Plugin
	getErr  error
	upserts int
}

func (s *stubStorage) FindOne(entityName string, out interface{}, filter []storageTY.Filter) error {
	return s.getErr
}

func (s *stubStorage) Upsert(entityName string, data interface{}, filter []storageTY.Filter) error {
	s.upserts++
	return nil
}

type stubBus struct{}

func (stubBus) Name() string                      { return "stub" }
func (stubBus) Close() error                      { return nil }
func (stubBus) Publish(string, interface{}) error { return nil }
func (stubBus) Subscribe(string, busTY.CallBackFunc) (int64, error) {
	return 0, nil
}
func (stubBus) Unsubscribe(string, int64) error { return nil }
func (stubBus) QueueSubscribe(string, string, busTY.CallBackFunc) (int64, error) {
	return 0, nil
}
func (stubBus) QueueUnsubscribe(string, string, int64) error { return nil }
func (stubBus) UnsubscribeAll(string) error                  { return nil }
func (stubBus) PausePublish()                                {}
func (stubBus) ResumePublish()                               {}
func (stubBus) TopicPrefix() string                          { return "" }

func TestSaveRetainValueMissingDocumentCreates(t *testing.T) {
	storage := &stubStorage{getErr: storageTY.ErrNoDocuments}
	api := New(context.Background(), zap.NewNop(), storage, stubBus{})
	err := api.Save(&fieldTY.Field{
		ID:        "missing-id",
		GatewayID: "gw1",
		NodeID:    "n1",
		SourceID:  "s1",
		FieldID:   "temp",
		Name:      "Temperature",
	}, true)
	require.NoError(t, err)
	require.Equal(t, 1, storage.upserts)
}

func TestSaveRetainValueOtherGetError(t *testing.T) {
	storage := &stubStorage{getErr: context.DeadlineExceeded}
	api := New(context.Background(), zap.NewNop(), storage, stubBus{})
	err := api.Save(&fieldTY.Field{ID: "id-1"}, true)
	require.Error(t, err)
	require.Equal(t, 0, storage.upserts)
}
