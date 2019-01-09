package persistence

import (
//	"github.com/mongodb/mongo-go-driver/mongo"
	bson "github.com/mongodb/mongo-go-driver/bson"
//	bsonx "github.com/mongodb/mongo-go-driver/x/bsonx"
//	mongo "github.com/mongodb/mongo-go-driver/x/mongo/driver"
	"fmt"

	mgo "gopkg.in/mgo.v2"
)

// MockMongoSession provides a mock abstraction to mongo
type MockMongoSession struct {
	MongoAbstraction
	ConnectMode  string
	QueryMode    string
	WriteMode    string
	FetchResult  Persistable
	FetchResults []Persistable
}

// NewMockMongoSession provides a mock with default 'positive' behaviors
func NewMockMongoSession() *MockMongoSession {
	mm := MockMongoSession{}
	mm.ConnectMode = "positive"
	mm.QueryMode   = "positive"
	mm.WriteMode   = "positive"
	return &mm
}

// ConnectToMongo mock. Controlled by mm.ConnectMode values 'positive' and 'no connect'
func (mm *MockMongoSession) ConnectToMongo() error {
	switch {
	case mm.ConnectMode == "positive":
		return nil
	case mm.ConnectMode == "no connect":
		return fmt.Errorf("mocked connection failure: no reachable servers")
	}
	return fmt.Errorf("Unknown mode for ConnectToMongo: %s", mm.ConnectMode)
}

// WriteCollection mock. Controlled by mm.WriteMode values 'positive', 'fail' and 'duplicate'
func (mm *MockMongoSession) WriteCollection(collectionName string, object Persistable) error {
	if err := mm.ConnectToMongo(); err != nil {
		return err
	}
	switch {
	case mm.WriteMode == "positive":
		return nil
	case mm.WriteMode == "fail":
		return fmt.Errorf("Mock error on write")
	case mm.WriteMode == "duplicate":
		err := mgo.QueryError{
			Code:    11000,
			Message: "Mock duplicate on write",
		}
		return &err
	}
	return fmt.Errorf("Unknown mode for WriteCollection: %s", mm.WriteMode)
}

// UpdateCollection mock. Controlled by mm.WriteMode values 'positive', 'fail' and 'missing'
func (mm *MockMongoSession) UpdateCollection(collectionName string, object Persistable) error {
	if err := mm.ConnectToMongo(); err != nil {
		return err
	}
	switch {
	case mm.WriteMode == "positive":
		return nil
	case mm.WriteMode == "fail":
		return fmt.Errorf("Mock error on update")
	case mm.WriteMode == "missing":
		err := mgo.QueryError{
			Code:    11000, // TODO: find the right error Code and type
			Message: "Mock not found on update",
		}
		return &err
	}
	return fmt.Errorf("Unknown mode for UpdateCollection: %s", mm.WriteMode)
}

// FetchIDFromCollection mock. Controlled by mm.QueryMode values 'positive' and 'fail'
func (mm *MockMongoSession) FetchIDFromCollection(collectionName string, id string) (result []byte, err error) {
	if err := mm.ConnectToMongo(); err != nil {
		return nil, err
	}
	switch {
	case mm.QueryMode == "positive":
		result, err = bson.Marshal(mm.FetchResult)
		return
	case mm.QueryMode == "fail":
		return nil, fmt.Errorf("Mock error on get")
	}
	return nil, fmt.Errorf("Unknown mode for FetchFromCollection: %s", mm.QueryMode)
}

// FetchFromCollection mock. Controlled by mm.QueryMode values 'positive' and 'fail'
func (mm *MockMongoSession) FetchFromCollection(collectionName string, query bson.M) (results [][]byte, err error) {
	//TODO: implement FetchFromCollection
	panic("ack!")
}

// FetchAllFromCollection mock. Controlled by mm.QueryMode values 'positive' and 'fail'
func (mm *MockMongoSession) FetchAllFromCollection(collectionName string) (results [][]byte, err error) {
	if err := mm.ConnectToMongo(); err != nil {
		return nil, err
	}
	switch {
	case mm.QueryMode == "positive":
		results = make([][]byte,len(mm.FetchResults))
		for i :=0; i < len(mm.FetchResults); i++ {
			if results[i], err = bson.Marshal(mm.FetchResults[i]); err != nil {
				return nil, err
			}
		}
		return
	case mm.QueryMode == "fail":
		return nil, fmt.Errorf("Mock error on get")
	}
	return nil, fmt.Errorf("Unknown mode for FetchAllFromCollection: %s", mm.QueryMode)
}

// DeleteFromCollection mock. Controlled by mm.QueryMode values 'positive' and 'fail'
func (mm *MockMongoSession) DeleteFromCollection(collectionName string, id string) error {
	if err := mm.ConnectToMongo(); err != nil {
		return err
	}
	switch {
	case mm.WriteMode == "positive":
		return nil
	case mm.WriteMode == "fail":
		return fmt.Errorf("Mock error on delete")
	case mm.WriteMode == "missing":
		err := mgo.QueryError{
			Code:    11000, // TODO: find the right error Code and type
			Message: "Mock not found on delete",
		}
		return &err
	}
	return fmt.Errorf("Unknown mode for DeleteFromCollection: %s", mm.QueryMode)
}
