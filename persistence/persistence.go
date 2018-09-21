package persistence

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/mongodb/mongo-go-driver/core/connstring"
	//"crypto/tls"

	//	mgo "github.com/globalsign/mgo"
	//	bson "github.com/globalsign/mgo/bson"
	"context"

	bson "github.com/mongodb/mongo-go-driver/bson"
	mongo "github.com/mongodb/mongo-go-driver/mongo"
	clientopt "github.com/mongodb/mongo-go-driver/mongo/clientopt"
)

// Persistable encapsulates the common features of any object that can be generically stored through this layer
type Persistable interface {
	GetID() string
	Decode([]byte) error
	ToJSON() []byte
}

// MongoAbstraction defines the set of DAL functions for accessing this Mongo collection
type MongoAbstraction interface {
	ConnectToMongo() error
	WriteCollection(collectionName string, object Persistable) error
	UpdateCollection(collectionName string, object Persistable) error
	FetchOneFromCollection(collectionName string, id string, object Persistable) error
	DeleteFromCollection(collectionName string, id string) error
}

// MongoSession defines an instantiation of a Mongo DAL. The session maintains a connected state to Mongodb.
type MongoSession struct {
	session        *mongo.Client
	db             *mongo.Database
	mongoURL       string
	connStr        connstring.ConnString
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
// toDuration should be expressed as a multiple of time.Second
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
		ms.logger = log.New(os.Stdout, "mongoLayer", log.Ldate|log.Ltime)
	}
	if len(overrideTo) > 0 {
		ms.timeoutSeconds = time.Duration(overrideTo[0]) * time.Second
	}
	ms.connStr, err = connstring.Parse(ms.mongoURL)
	if err != nil {
		return
	}
/* 	ms.session, err = mongo.NewClientFromConnString(ms.connStr)
	if err != nil {
		return
	}
 	ms.db = ms.session.Database(ms.dbName)
*/
	return
}

// ConnectToMongo creates a connection to the specified mongodb instance
func (ms *MongoSession) ConnectToMongo() (err error) {
	ms.session, err = mongo.Connect(context.Background(), ms.mongoURL, clientopt.ConnectTimeout(ms.timeoutSeconds))
 	if err != nil { return }
	ms.db = ms.session.Database(ms.dbName)
	// if n, checkErr := ms.session.ListDatabaseNames( context.Background(),nil); checkErr != nil {
	//  	err = fmt.Errorf("Validation of connection failed: %s ... And there are %d DBs not there", checkErr.Error(), len(n))
	// }
	return
}

// CheckAndReconnect ensures that there is an active DB connection to mongo. Attempts to reestablish connection if needed
func (ms *MongoSession) CheckAndReconnect() (err error) {
	if ms.session == nil {
		err = ms.ConnectToMongo()
	}
	return
}

// WriteCollection writes the specified Persistable object to a given collection
func (ms *MongoSession) WriteCollection(coll string, obj Persistable) error {
	if err := ms.CheckAndReconnect(); err != nil {
		ms.logger.Printf("WriteCollection: could not establish mongo connection: %s", err)
		return err
	}
	myCollection := ms.db.Collection(coll)
	_, err := myCollection.InsertOne(context.Background(), obj)
	return err
}

// UpdateCollection updates the Persistable object in the specified collection with a matching _id element to the passed in object
func (ms *MongoSession) UpdateCollection(coll string, obj Persistable) (err error) {
	if err = ms.CheckAndReconnect(); err != nil {
		ms.logger.Printf("UpdateCollection: could not establish mongo connection: %s", err)
		return
	}
	/*	if !ms.collectionExists(coll) {
			ms.logger.Printf("UpdateCollection: missing collection: %s", coll)
			return fmt.Errorf("missing collection: %s", coll)
		}
	*/
	myCollection := ms.db.Collection(coll)
	var uResult *mongo.UpdateResult
	var updateDoc *bson.Document
	if updateBSON,err := bson.Marshal(obj); err != nil {
		return fmt.Errorf("Fail to Marshal bson: %s", err)
	} else if updateDoc, err = bson.ReadDocument(updateBSON); err != nil {
		return fmt.Errorf("Fail to create Document: %s", err)
	} else {
		uResult, err = myCollection.ReplaceOne(
			context.Background(),
			bson.NewDocument(
				bson.EC.String("_id", obj.GetID()),
			),
			updateDoc,
		)
		if err != nil {
			ms.logger.Printf("UpdateCollection: topology is closed on update attempt for %s", obj.GetID())
			return fmt.Errorf("Fail to update Document: %s", err)
		}
	}
	if uResult.MatchedCount == 0 {
		err = fmt.Errorf("Update failed. %s not found in collection %s", obj.GetID(), coll)
		ms.logger.Printf("UpdateCollection: %s not found in collection %s", obj.GetID(), coll)
	}
	return
}

// FetchOneFromCollection fetches the Persistable by ID from the specified collection
func (ms *MongoSession) FetchOneFromCollection(coll string, id string, result Persistable) (err error) {
	if err = ms.CheckAndReconnect(); err != nil {
		ms.logger.Printf("FetchOneFromCollection: could not establish mongo connection: %s", err)
		return
	}
	myCollection := ms.db.Collection(coll)
	queryResult := bson.NewDocument()
	err = myCollection.FindOne(
		context.Background(),
		bson.NewDocument(
			bson.EC.String("_id", id),
		),
	).Decode(queryResult)
	if err != nil {
		ms.logger.Printf("FetchOneFromCollection: %s on update attempt for %s", err.Error(), id)
		return
	}
	var bytes []byte 
	bytes, err = queryResult.MarshalBSON()
	if err != nil {
		return
	}

	return result.Decode(bytes)
}

/*
// FetchAllFromCollection fetches all the Persistables from the specified collection
func (ms *MongoSession) FetchAllFromCollection(coll string) (result []bson.Raw, err error) {
	if err = ms.CheckAndReconnect(); err != nil {
		ms.logger.Printf("FetchAllFromCollection: could not establish mongo connection: %s", err)
		return
	}
	myCollection := ms.db.C(coll)
	q := myCollection.Find(bson.D{{}})
	err = q.All(&result)
	return
}
*/

// DeleteFromCollection removes the Loc by ID from the specified collection
func (ms *MongoSession) DeleteFromCollection(coll string, id string) (err error) {
	if err = ms.CheckAndReconnect(); err != nil {
		ms.logger.Printf("DeleteFromCollection: could not establish mongo connection: %s", err)
		return
	}
	myCollection := ms.db.Collection(coll)
	var dResult *mongo.DeleteResult
	dResult, err = myCollection.DeleteOne(
		context.Background(),
		bson.NewDocument(
			bson.EC.String("_id", id),
		),
	)
	if err != nil {
		ms.logger.Printf("DeleteFromCollection: no mongo connection: %s", err)
	} else if dResult.DeletedCount == 0 {
		err = fmt.Errorf("Delete failed. %s not found in collection %s", id, coll)
		ms.logger.Printf("DeleteFromCollection: %s not found in collection %s", id, coll)
	}
	return
}

func (ms *MongoSession) collectionExists(collName string) bool {
	names, err := ms.db.ListCollections(context.Background(), bson.NewDocument())
	if err != nil {
		return false
	}

	for names.Next(context.Background()) {
		elem := bson.NewDocument()
		if err := names.Decode(elem); err != nil {
			log.Fatal(err)
		}
		if elem.Lookup("name").StringValue() == collName {
			return true
		}
	}

	return false
}
