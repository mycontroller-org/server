package helper

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/mycontroller-org/backend/v2/pkg/model"
)

// GetID returns ID from the interface
func GetID(data interface{}) string {
	_, value, err := GetValueByKeyPath(data, model.KeyID)
	if err != nil {
		return ""
	}
	id, ok := value.(string)
	if ok {
		return id
	}
	return fmt.Sprintf("%v", value)
}

// CloneSlice source to destination
func CloneSlice(src []interface{}) ([]interface{}, error) {
	dst := make([]interface{}, len(src))
	copy(dst, src)
	return dst, nil
}

// GetValueByKeyPath returns type and value from the given struct
// returns reflect.Kind, value, error
func GetValueByKeyPath(data interface{}, keyPath string) (reflect.Kind, interface{}, error) {
	dataVal := reflect.ValueOf(data)
	if dataVal.Kind() == reflect.Ptr {
		dataVal = dataVal.Elem()
	}

	if dataVal.Kind() != reflect.Struct {
		return reflect.Invalid, nil, fmt.Errorf("one level pointer or struct interface should be supplied. received:%s", dataVal.Kind())
	}

	// dot key path
	keys := strings.Split(keyPath, ".")

	finalVal := dataVal
	for keyIndex := 0; keyIndex < len(keys); keyIndex++ {
		found := false
		expectedName := strings.ToLower(keys[keyIndex])
		//fmt.Printf("\nin loop[%d], key:%s, kind:%v, \nvalue:%+v\n", keyIndex, expectedName, finalVal.Kind(), finalVal)
		if finalVal.Kind() == reflect.Ptr || finalVal.Kind() == reflect.Interface {
			finalVal = finalVal.Elem()
			//fmt.Printf("\nin loop[%d], key:%s, kind:%v, \nvalue:%+v\n", keyIndex, expectedName, finalVal.Kind(), finalVal)
		}

		switch finalVal.Kind() {
		case reflect.Struct:
			if finalVal.Kind() == reflect.Struct {
				for subKeyIndex := 0; subKeyIndex < finalVal.NumField(); subKeyIndex++ {
					receivedName := strings.ToLower(finalVal.Type().Field(subKeyIndex).Name)
					//fmt.Printf("comparing names: %s vs %s\n", expectedName, receivedName)
					if expectedName == receivedName {
						finalVal = finalVal.FieldByIndex([]int{subKeyIndex})
						//fmt.Printf("found: %s, %+v\n", receivedName, finalVal)
						found = true
						break
					}
				}
			}

		case reflect.Map:
			for _, mapKey := range finalVal.MapKeys() {
				receivedName := strings.ToLower(mapKey.String())
				//fmt.Printf("comparing names: %s vs %s\n", expectedName, receivedName)
				if expectedName == receivedName {
					finalVal = finalVal.MapIndex(mapKey)
					//fmt.Printf("found: %s, %+v\n", receivedName, finalVal)
					found = true
					break
				}
			}

		default:
			return reflect.Invalid, nil, fmt.Errorf("not supported kind:%s", finalVal.Kind())
		}

		if !found {
			return reflect.Invalid, nil, fmt.Errorf("key not found:%s", expectedName)
		}
	}
	//fmt.Printf("returning, kind:%v, \nvalue:%+v\n\n", finalVal.Kind(), finalVal)

	if finalVal.Kind() == reflect.Ptr || finalVal.Kind() == reflect.Interface {
		if finalVal.IsNil() {
			return finalVal.Kind(), nil, nil
		}
		finalVal = finalVal.Elem()
	}
	return finalVal.Kind(), finalVal.Interface(), nil
}
