package persistence

import (
	"strings"
	"fmt"
	"log"
	"os"
	"time"
	"context"

	mongo "go.mongodb.org/mongo-driver/mongo"
	bson "go.mongodb.org/mongo-driver/bson"
	conn "go.mongodb.org/mongo-driver/x/mongo/driver/connstring"
	options "go.mongodb.org/mongo-driver/mongo/options"
)

// Persistable encapsulates the common features of any object that can be generically stored through this layer
type Persistable interface {
	GetID() string
//	Decode([]byte) error
}

// MongoAbstraction defines the set of DAL functions for accessing this Mongo collection
type MongoAbstraction interface {
	ConnectToMongo() error
	CountInCollection(collectionName string, query bson.M) (int64, error)
	DeleteFromCollection(collectionName string, id string) error
	FetchAllFromCollection(collectionName string) ([][]byte, error)
	FetchFromCollection(collectionName string, query bson.M) ([][]byte, error)
	FetchIDFromCollection(collectionName string, id string) ([]byte,error)
	UpdateCollection(collectionName string, object Persistable) error
	WriteCollection(collectionName string, object Persistable) error
}

// MongoSession defines an instantiation of a Mongo DAL. The session maintains a connected state to Mongodb.
type MongoSession struct {
	session        *mongo.Client
	db             *mongo.Database
	mongoURL       string
	connStr        conn.ConnString
	dbName         string
	timeoutSeconds time.Duration
	logger         *log.Logger
}

// DbName designates the default DB name in mongo
const (
	DefaultDbName  string        = "defaultDB"
	DefaultTimeout time.Duration = 10 * time.Second
)

// NewMongoSession is a factory method to create a fresh MongoSession for a given connection string and DB
// overrideTo should be expressed as a multiple of time.Second
func NewMongoSession(mongoURL string, dbName string, logger *log.Logger, overrideTo ...int64) (ms *MongoSession, err error) {
	ms = &MongoSession{
		mongoURL:       mongoURL,
		dbName:         dbName,
		timeoutSeconds: DefaultTimeout,
		logger:         logger,
	}
	if ms.dbName == "" {
		ms.dbName = DefaultDbName
	}
	if ms.logger == nil {
		ms.logger = log.New(os.Stdout, "MongoSession: ", log.Ldate|log.Ltime)
	}
	if len(overrideTo) > 0 {
		ms.timeoutSeconds = time.Duration(overrideTo[0]) * time.Second
	}
	ms.connStr, err = conn.Parse(ms.mongoURL)
	if err != nil {
		return
	}
	ms.logger.Printf("New MongoSession established for %s", ms.mongoURL)
	if  err = ms.ConnectToMongo(); err != nil {
		err = fmt.Errorf("MongoSession connect failure: %v", err)
	}
	return
}

// ConnectToMongo creates a connection to the specified mongodb instance
func (ms *MongoSession) ConnectToMongo() (err error) {
	opts := options.Client().SetConnectTimeout(ms.timeoutSeconds).SetAppName("wordassassin").ApplyURI(ms.mongoURL)
	ms.session, err = mongo.Connect(context.Background(), opts)
	if err != nil { return }
	ms.db = ms.session.Database(ms.dbName)
	err = ms.CheckConnection()
	return
}	

// CheckAndReconnect ensures that there is an active DB connection to mongo. Attempts to reestablish connection if needed
func (ms *MongoSession) CheckAndReconnect() (err error) {
	if checkErr := ms.CheckConnection(); checkErr != nil {
		err = ms.ConnectToMongo()
	}	
	return
}	

// CheckConnection validates if the server can be successfully pinged. Provides the error from the mongo client on false.
// Uses a hardcoded timeout of 2 seconds. Might want to change that at some point.
func (ms *MongoSession) CheckConnection() (err error) {
	if ms.session == nil {
		panic("CheckConnection called with nil session")
	}	
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	err = ms.session.Ping( ctx, nil )
	if err != nil && err.Error() == "context deadline exceeded" {
		err = fmt.Errorf("Ping timed out. No DB found")
	}	

	return
}	

// CountInCollection provides a count of the documents in the specified collection that match the query Document
func (ms *MongoSession) CountInCollection(coll string, query bson.M) (result int64, err error) {
	if err := ms.CheckAndReconnect(); err != nil {
		ms.logger.Printf("CountInCollection: could not establish mongo connection: %s", err)
		return 0, err
	}	
	
	myCollection := ms.db.Collection(coll)
	count, err := myCollection.CountDocuments(context.Background(), &query)
	if err != nil { 
		ms.logger.Printf("CountInCollection: %s on Count attempt from %s", err.Error(), coll)
		return 0, err
	}	
	return count, nil
}	

// DeleteFromCollection removes the Loc by ID from the specified collection
// If the ID is not found, logs and then returns an error containing the message "no documents"
func (ms *MongoSession) DeleteFromCollection(coll string, id string) (err error) {
	if err = ms.CheckAndReconnect(); err != nil {
		ms.logger.Printf("DeleteFromCollection: could not establish mongo connection: %s", err)
		return
	}

	myCollection := ms.db.Collection(coll)
	var dResult *mongo.DeleteResult
	dResult, err = myCollection.DeleteOne(
		context.Background(),
		bson.M{ "_id": id },
	)
	if err != nil {
		ms.logger.Printf("DeleteFromCollection: no mongo connection: %s", err)
	} else if dResult.DeletedCount == 0 {
		err = fmt.Errorf("Delete failed: no documents for id=%s in collection %s", id, coll)
		ms.logger.Printf("DeleteFromCollection: no documents for id=%s in collection %s", id, coll)
	}
	
	return
}

// FetchAllFromCollection fetches all the Persistables from the specified collection
// They are returned in an array of the specified type in sample, which is supplied only for typing purposes
func (ms *MongoSession) FetchAllFromCollection(coll string) (results [][]byte, err error) {
	return ms.FetchFromCollection(coll, bson.M{})
}	

// FetchFromCollection fetches the Persistables from the specified collection that match the query Document
// They are returned in an array of the specified type in sample, which is supplied only for typing purposes
func (ms *MongoSession) FetchFromCollection(coll string, query bson.M) (results [][]byte, err error) {
	if err := ms.CheckAndReconnect(); err != nil {
		ms.logger.Printf("FetchFromCollection: could not establish mongo connection: %s", err)
		return nil, err
	}	

	results = make([][]byte,0)
	myCollection := ms.db.Collection(coll)
	cur, err := myCollection.Find(context.Background(), &query)
	if err != nil { 
		ms.logger.Printf("FetchFromCollection: %s on Find attempt from %s", err.Error(), coll)
		return nil, err
	}	
	defer cur.Close(context.Background())
	
	for cur.Next(context.Background()) {
		var doc bson.M
		err = cur.Decode(&doc)
		if err != nil { 
			ms.logger.Printf("FetchFromCollection: %s on Decode attempt for %s", err.Error(), doc)
			return nil, err
		}	
		var elem []byte
		elem, err = bson.Marshal(doc)
		if err != nil {
			return nil, err
		}	
		results = append(results, elem)
	}	
	
	return results, cur.Err()
}	

// FetchIDFromCollection fetches the Persistable by ID from the specified collection
// If the ID is not found, logs and then returns an error containing the message "no documents"
func (ms *MongoSession) FetchIDFromCollection(coll string, id string) (result []byte, err error) {
	if err = ms.CheckAndReconnect(); err != nil {
		ms.logger.Printf("FetchIDFromCollection: could not establish mongo connection: %s", err)
		return
	}	
	
	myCollection := ms.db.Collection(coll)
	var queryResult bson.M
	err = myCollection.FindOne(
		context.Background(),
		bson.M{ "_id": id },
	).Decode(&queryResult)	
	if err != nil {
		ms.logger.Printf("FetchIDFromCollection: %s on Decode attempt for %s", err.Error(), id)
		return
	}	
	result, err = bson.Marshal(queryResult)
	return
}	

// UpdateCollection updates the Persistable object in the specified collection with a matching _id element to the passed in object
// If the ID is not found, logs and then returns an error containing the message "no documents"
func (ms *MongoSession) UpdateCollection(coll string, obj Persistable) (err error) {
	if err = ms.CheckAndReconnect(); err != nil {
		ms.logger.Printf("UpdateCollection: could not establish mongo connection: %s", err)
		return
	}	
	
	myCollection := ms.db.Collection(coll)
	uResult, replaceErr := myCollection.ReplaceOne(
		context.Background(),
		bson.M{ "_id": obj.GetID() },
		obj,
	)	
	if replaceErr != nil {
		ms.logger.Printf("UpdateCollection: topology is closed on update attempt for %s", obj.GetID())
		err = fmt.Errorf("Fail to update Document: %s", replaceErr)
	}	
	if uResult.MatchedCount == 0 {
		ms.logger.Printf("UpdateCollection: no documents for ID=%s in collection %s", obj.GetID(), coll)
		err = fmt.Errorf("Update failed. no documents for ID=%s in collection %s", obj.GetID(), coll)
	}	
	
	return
}	

// WriteCollection writes the specified Persistable object to a given collection
func (ms *MongoSession) WriteCollection(coll string, obj Persistable) (err error) {
	if err = ms.CheckAndReconnect(); err != nil {
		ms.logger.Printf("WriteCollection: could not establish mongo connection: %s", err)
		return
	}

	myCollection := ms.db.Collection(coll)
	wResult, insErr := myCollection.InsertOne(context.Background(), obj)
	if insErr != nil {
		if strings.Contains(insErr.Error(), "duplicate") {
			err = fmt.Errorf("Write failed: duplicate key on insert for %s", obj.GetID())
			ms.logger.Printf("WriteCollection: duplicate key on insert for %s", obj.GetID())
		} else {
			err = fmt.Errorf("Write failed. topology is closed on insert: %s", err)
			ms.logger.Printf("WriteCollection: topology is closed on insert attempt for %s", obj.GetID())
		}
	} else if wResult.InsertedID == nil {
		err = fmt.Errorf("Write failed. %s not inserted in collection %s", obj.GetID(), coll)
		ms.logger.Printf("WriteCollection: %s not inserted in collection %s", obj.GetID(), coll)
	}

	return
}

// func (ms *MongoSession) collectionExists(collName string) bool {
	// 	names, err := ms.db.ListCollections(context.Background(), bson.Doc{})
// 	if err != nil {
	// 		return false
// 	}

// 	for names.Next(context.Background()) {
	// 		var elem bson.Doc
	// 		if err := names.Decode(elem); err != nil {
// 			log.Fatal(err)
// 		}
// 		if elem.Lookup("name").StringValue() == collName {
// 			return true
// 		}
// 	}

// 	return false
// }
