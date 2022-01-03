package memory

import (
	"errors"
	"fmt"
	"reflect"

	cloneUtils "github.com/mycontroller-org/server/v2/pkg/utils/clone"
	filterUtils "github.com/mycontroller-org/server/v2/pkg/utils/filter_sort"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/type"
	"go.uber.org/zap"
)

// Close Implementation
func (s *Store) Close() error {
	// sync memory entities to disk
	zap.L().Info("store memory data into disk started")
	s.writeToDisk()
	zap.L().Info("store memory data into disk completed")
	return nil
}

// Ping Implementation
func (s *Store) Ping() error {
	return nil
}

// Insert Implementation
func (s *Store) Insert(entityName string, data interface{}) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	newID := filterUtils.GetID(data)
	entity := s.getByID(entityName, newID)
	if entity != nil {
		return fmt.Errorf("a entity found with the id: %s", newID)
	}

	clonedData := cloneUtils.Clone(data)
	s.addEntity(entityName, clonedData)
	return nil
}

// Upsert Implementation
func (s *Store) Upsert(entityName string, data interface{}, filters []storageTY.Filter) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	clonedData := cloneUtils.Clone(data)
	return s.updateEntity(entityName, clonedData, filters, true)
}

// Update Implementation
func (s *Store) Update(entityName string, data interface{}, filters []storageTY.Filter) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	clonedData := cloneUtils.Clone(data)
	return s.updateEntity(entityName, clonedData, filters, false)
}

// Find Implementation
func (s *Store) Find(entityName string, out interface{}, filters []storageTY.Filter, pagination *storageTY.Pagination) (*storageTY.Result, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	outVal := reflect.ValueOf(out)
	if outVal.Kind() != reflect.Ptr {
		return nil, errors.New("results argument must be a pointer to a slice")
	}

	sliceVal := outVal.Elem()
	elementType := sliceVal.Type().Elem()
	entities := s.getEntities(entityName)

	filteredEntities := filterUtils.Filter(entities, filters, false)
	entitiesSorted, count := filterUtils.Sort(filteredEntities, pagination)

	for index, entity := range entitiesSorted {
		clonedEntity := cloneUtils.Clone(entity)

		if sliceVal.Len() == index {
			// slice is full
			newElem := reflect.New(elementType)
			sliceVal = reflect.Append(sliceVal, newElem.Elem())
			sliceVal = sliceVal.Slice(0, sliceVal.Cap())
		}
		sliceVal.Index(index).Set(reflect.ValueOf(clonedEntity).Elem())
	}

	outVal.Elem().Set(sliceVal.Slice(0, len(entitiesSorted)))

	offset := int64(0)
	if pagination != nil {
		offset = pagination.Offset
	}

	result := &storageTY.Result{
		Offset: offset,
		Count:  count,
		Data:   out,
	}
	if pagination != nil {
		result.Limit = pagination.Limit
	}
	return result, nil
}

// FindOne Implementation
func (s *Store) FindOne(entityName string, out interface{}, filters []storageTY.Filter) error {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	entities := s.getEntities(entityName)
	entities = filterUtils.Filter(entities, filters, true)

	if len(entities) > 0 {
		clonedEntity := cloneUtils.Clone(entities[0])
		outVal := reflect.ValueOf(out)
		outVal.Elem().Set(reflect.ValueOf(clonedEntity).Elem())
		return nil
	}
	return errors.New("requested data not available")
}

// Delete Implementation
func (s *Store) Delete(entityName string, filters []storageTY.Filter) (int64, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	entities := s.getEntities(entityName)
	filteredEntities := filterUtils.Filter(entities, filters, false)

	if len(filteredEntities) > 0 {
		for _, entity := range filteredEntities {
			id := filterUtils.GetID(entity)
			s.removeEntity(entityName, id)
		}
		return int64(len(filteredEntities)), nil
	}
	return 0, nil
}

func (s *Store) getEntities(entityName string) []interface{} {
	data, ok := s.data[entityName]
	if !ok {
		data = make([]interface{}, 0)
		s.data[entityName] = data
	}
	return data
}

func (s *Store) getByID(entityName, id string) interface{} {
	entities := s.getEntities(entityName)
	for _, entity := range entities {
		eID := filterUtils.GetID(entity)
		if eID == id {
			return entity
		}
	}
	return nil
}

func (s *Store) addEntity(entityName string, entity interface{}) {
	if _, ok := s.data[entityName]; !ok {
		s.data[entityName] = make([]interface{}, 0)
	}
	s.data[entityName] = append(s.data[entityName], entity)
}

func (s *Store) updateEntity(entityName string, entity interface{}, filters []storageTY.Filter, forceUpdate bool) error {
	//zap.L().Info("received data for update", zap.String("entity", entityName), zap.Any("data", entity))
	sourceID := ""
	suppliedID := filterUtils.GetID(entity)
	if suppliedID != "" {
		if s.getByID(entityName, suppliedID) != nil {
			sourceID = suppliedID
		}
	}

	if sourceID == "" && len(filters) > 0 { // with filters find a entity
		entities := s.getEntities(entityName)
		entities = filterUtils.Filter(entities, filters, true)
		if len(entities) > 1 {
			return errors.New("more than one entities found, with the supplied filter")
		} else if len(entities) > 0 {
			sourceID = filterUtils.GetID(entities[0])
		}
	}

	if sourceID != "" {
		for index, entry := range s.data[entityName] {
			eID := filterUtils.GetID(entry)
			if sourceID == eID {
				s.data[entityName][index] = entity
				//	zap.L().Info("Updated on the existing entity", zap.Any("old", entry), zap.Any("new", entity))
				return nil
			}
		}
	}
	if forceUpdate {
		s.data[entityName] = append(s.data[entityName], entity)
		//	zap.L().Info("Entity not available, added", zap.Any("new", entity))
		return nil
	}
	return errors.New("entity not available")
}

func (s *Store) removeEntity(entityName, id string) {
	entities := s.getEntities(entityName)
	for index, entry := range entities {
		eID := filterUtils.GetID(entry)
		if id == eID {
			s.data[entityName] = append(s.data[entityName][:index], s.data[entityName][index+1:]...)
			return
		}
	}
}
