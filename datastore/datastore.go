package datastore

import (
	"bytes"
	"encoding/gob"
	"errors"
	"reflect"
	"strings"

	"github.com/docker/swarm/pkg/store"
)

//DataStore exported
type DataStore interface {
	// PutObject adds a new Record based on an object into the datastore
	PutObject(kvObject KV) error
	// PutObjectAtomic provides an atomic add and update operation for a Record
	PutObjectAtomic(kvObject KV) error
	// KVStore returns access to the KV Store
	KVStore() store.Store
}

type datastore struct {
	store  store.Store
	config *StoreConfiguration
}

//StoreConfiguration exported
type StoreConfiguration struct {
	Addrs    []string
	Provider string
}

//KV Key Value interface used by objects to be part of the DataStore
type KV interface {
	Key() []string
	Value() []byte
	Index() uint64
	SetIndex(uint64)
}

//Key provides convenient method to create a Key
func Key(key ...string) string {
	keychain := []string{"docker", "libnetwork"}
	keychain = append(keychain, key...)
	str := strings.Join(keychain, "/")
	return str + "/"
}

var errNewDatastore = errors.New("Error creating new Datastore")
var errInvalidConfiguration = errors.New("Invalid Configuration passed to Datastore")

// newClient used to connect to KV Store
func newClient(kv string, addrs []string) (DataStore, error) {
	store, err := store.CreateStore(kv, addrs, store.Config{})
	if err != nil {
		return nil, err
	}
	ds := &datastore{store: store}
	return ds, nil
}

// NewDataStore creates a new instance of LibKV data store
func NewDataStore(config *StoreConfiguration) (DataStore, error) {
	if config == nil {
		return nil, errInvalidConfiguration
	}
	return newClient(config.Provider, config.Addrs)
}

func (ds *datastore) KVStore() store.Store {
	return ds.store
}

// PutObjectAtomic adds a new Record based on an object into the datastore
func (ds *datastore) PutObjectAtomic(kvObject KV) error {
	if kvObject == nil {
		return errors.New("kvObject is nil")
	}
	kvObjValue := kvObject.Value()

	if kvObjValue == nil {
		return ErrInvalidAtomicRequest
	}
	_, err := ds.store.AtomicPut(Key(kvObject.Key()...), []byte{}, kvObjValue, kvObject.Index())
	if err != nil {
		return err
	}

	_, index, err := ds.store.Get(Key(kvObject.Key()...))
	if err != nil {
		return err
	}
	kvObject.SetIndex(index)
	return nil
}

// PutObject adds a new Record based on an object into the datastore
func (ds *datastore) PutObject(kvObject KV) error {
	if kvObject == nil {
		return errors.New("kvObject is nil")
	}
	return ds.putObjectWithKey(kvObject, kvObject.Key()...)
}

func (ds *datastore) putObjectWithKey(kvObject KV, key ...string) error {
	kvObjValue := kvObject.Value()

	// If the KVObject provides a Value, use it as is. No further processing required.
	if kvObjValue != nil {
		return ds.store.Put(Key(key...), kvObjValue)
	}

	// Else, DataStore will rely on `kv:"<tag>"` Tag to form the KV pair(s) for the object
	var keychain []string
	keychain = append(keychain, key...)
	value := reflect.ValueOf(kvObject).Elem()

	for i := 0; i < value.NumField(); i++ {
		field := value.Type().Field(i)
		fieldValue := value.Field(i)
		derefValue := fieldValue
		if fieldValue.Kind() == reflect.Ptr {
			derefValue = fieldValue.Elem()
		}
		tag := field.Tag.Get("kv")
		switch tag {
		case "leaf": // "leaf" field is encoded into []byte and used as value
			childKeychain := append(keychain, field.Name)
			b, err := GetBytes(derefValue.Interface())
			if err != nil {
				return err
			}
			err = ds.store.Put(Key(childKeychain...), b)
			if err != nil {
				return err
			}
		case "iterative": // "iterative" fields are iterated and expanded into multiple kv pairs
			err := ds.iterateObject(field, derefValue, keychain)
			if err != nil {
				return err
			}
		case "recursive": // "recursive" fields are supported only for Structs
			err := ds.recurseObject(fieldValue, keychain)
			if err != nil {
				return err
			}
		}

	}
	return nil
}

// Supports just a Map iteration as of now
func (ds *datastore) iterateObject(field reflect.StructField, value reflect.Value, keychain []string) error {
	switch value.Kind() {
	case reflect.Map:
		childKeychain := append(keychain, field.Name)
		for _, k := range value.MapKeys() {
			mapVal := value.MapIndex(k)
			derefMapVal := mapVal
			if mapVal.Kind() == reflect.Ptr {
				derefMapVal = mapVal.Elem()
			}
			if recursiveObj, ok := derefMapVal.Interface().(KV); ok {
				if recursiveObj.Value() != nil {
					key := Key(append(childKeychain, recursiveObj.Key()...)...)
					err := ds.store.Put(key, recursiveObj.Value())
					if err != nil {
						return err
					}
				}
			}
			b, err := GetBytes(derefMapVal.Interface())
			if err != nil {
				return err
			}
			err = ds.store.Put(Key(append(childKeychain, k.String())...), b)
			if err != nil {
				return err
			}
		}
	default:
		return errors.New("Map can be the only iterative type")
	}
	return nil
}

// Supports just a KV-Type Structure as of now
func (ds *datastore) recurseObject(value reflect.Value, keychain []string) error {
	derefValue := value
	if value.Kind() == reflect.Ptr {
		derefValue = value.Elem()
	}
	switch derefValue.Kind() {
	case reflect.Struct:
		if recursiveObj, ok := value.Interface().(KV); ok {
			ds.putObjectWithKey(recursiveObj, append(keychain, recursiveObj.Key()...)...)
		} else {
			return errors.New("A recursive Struct must implement KV")
		}
	default:
		return errors.New("Struct can be the only recursive type")
	}
	return nil
}

//GetBytes provides a convenience method to convert any data into []byte
func GetBytes(key interface{}) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(key)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
