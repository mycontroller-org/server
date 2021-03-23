package mysensors

import (
	"bytes"
	"encoding/binary"
	hexENC "encoding/hex"
	"sync"
	"time"
)

var (
	fwStore = firmwareStore{store: make(map[string]*firmwareRaw)}
)

type firmwareStore struct {
	store map[string]*firmwareRaw
	mutex sync.Mutex
}

// add firmware to the store
func (fws *firmwareStore) add(key string, fw *firmwareRaw) {
	fws.mutex.Lock()
	defer fws.mutex.Unlock()
	fws.store[key] = fw
}

// get a firmware with availability status from the store
func (fws *firmwareStore) get(key string) (*firmwareRaw, bool) {
	fws.mutex.Lock()
	defer fws.mutex.Unlock()
	fw, ok := fws.store[key]
	return fw, ok
}

// remove a firmware from the store
// func (fws *firmwareStore) remove(key string) {
// 	fws.mutex.Lock()
// 	defer fws.mutex.Unlock()
// 	delete(fws.store, key)
// }

// check the firmware access data and remove it if it is too old
func (fws *firmwareStore) purge() {
	fws.mutex.Lock()
	defer fws.mutex.Unlock()
	for k, fw := range fws.store {
		if time.Since(fw.LastAccess) >= firmwarePurgeInactiveTime { // eligible for purging
			delete(fws.store, k)
		}
	}
}

// toHex returns hex string
func toHex(in interface{}) (string, error) {
	var bBuf bytes.Buffer
	err := binary.Write(&bBuf, binary.LittleEndian, in)
	if err != nil {
		return "", err
	}
	return hexENC.EncodeToString(bBuf.Bytes()), nil
}

// toStruct updates struct from hex string
func toStruct(hex string, out interface{}) error {
	hb, err := hexENC.DecodeString(hex)
	if err != nil {
		return err
	}
	r := bytes.NewReader(hb)
	return binary.Read(r, binary.LittleEndian, out)
}
