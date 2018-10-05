package persistence

import (
	"strings"
	"fmt"
	"log"
	"os"
	"time"
	"context"

	"github.com/mongodb/mongo-go-driver/core/connstring"
	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/mongodb/mongo-go-driver/mongo"
	"github.com/mongodb/mongo-go-driver/mongo/clientopt"

//	"bytes"
	mgobson "gopkg.in/mgo.v2/bson"
)

func getBytes(key interface{}) ([]byte, error) {
    //var buf bytes.Buffer
	buf, err := mgobson.Marshal(key); if err != nil {
        return nil, err
    }
    return buf, nil
}

// Persistable encapsulates the common features of any object that can be generically stored through this layer
type Persistable interface {
	GetID() string
//	Decode([]byte) error
}

// MongoAbstraction defines the set of DAL functions for accessing this Mongo collection
type MongoAbstraction interface {
	ConnectToMongo() error
	WriteCollection(collectionName string, object Persistable) error
	UpdateCollection(collectionName string, object Persistable) error
	FetchOneFromCollection(collectionName string, id string) ([]byte,error)
	FetchAllFromCollection(collectionName string) ([][]byte, error)
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
		ms.logger = log.New(os.Stdout, "mongoLayer", log.Ldate|log.Ltime)
	}
	if len(overrideTo) > 0 {
		ms.timeoutSeconds = time.Duration(overrideTo[0]) * time.Second
	}
	ms.connStr, err = connstring.Parse(ms.mongoURL)
	if err != nil {
		return
	}
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

// UpdateCollection updates the Persistable object in the specified collection with a matching _id element to the passed in object
// If the ID is not found, logs and then returns an error containing the message "no documents"
func (ms *MongoSession) UpdateCollection(coll string, obj Persistable) (err error) {
	if err = ms.CheckAndReconnect(); err != nil {
		ms.logger.Printf("UpdateCollection: could not establish mongo connection: %s", err)
		return
	}

	myCollection := ms.db.Collection(coll)
	var uResult *mongo.UpdateResult
	var updateDoc *bson.Document
	if updateBSON,err := getBytes(obj); err != nil {
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
// If the ID is not found, logs and then returns an error containing the message "no documents"
func (ms *MongoSession) FetchOneFromCollection(coll string, id string) (result []byte, err error) {
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
		ms.logger.Printf("FetchOneFromCollection: %s on Decode attempt for %s", err.Error(), id)
		return
	}
	result, err = queryResult.MarshalBSON()
	return
}


// FetchAllFromCollection fetches all the Persistables from the specified collection
// They are returned in an array of the specified type in sample, which is supplied only for typing purposes
func (ms *MongoSession) FetchAllFromCollection(coll string) (results [][]byte, err error) {
	if err := ms.CheckAndReconnect(); err != nil {
		ms.logger.Printf("FetchAllFromCollection: could not establish mongo connection: %s", err)
		return nil, err
	}

	results = make([][]byte,0)
	myCollection := ms.db.Collection(coll)
	cur, err := myCollection.Find(context.Background(), nil)
	if err != nil { 
		ms.logger.Printf("FetchAllFromCollection: %s on Find attempt from %s", err.Error(), coll)
		return nil, err
	}
	defer cur.Close(context.Background())
	
	for cur.Next(context.Background()) {
		doc := bson.NewDocument()
		err = cur.Decode(doc)
		if err != nil { 
			ms.logger.Printf("FetchAllFromCollection: %s on Decode attempt for %s", err.Error(), doc)
			return nil, err
		}
		var elem []byte
		elem, err = doc.MarshalBSON()
		if err != nil {
		 	return nil, err
		}
		results = append(results, elem)
	}

	return results, cur.Err()
}

// // FetchAllFromCollection fetches all the Persistables from the specified collection
// // They are returned in an array of the specified type in sample, which is supplied only for typing purposes
// func (ms *MongoSession) FetchAllFromCollection(coll string, sample Persistable) ([]Persistable, error) {
// 	if err := ms.CheckAndReconnect(); err != nil {
// 		ms.logger.Printf("FetchAllFromCollection: could not establish mongo connection: %s", err)
// 		return nil, err
// 	}

// 	// Create result set after validating type is Persistable
// 	// make the concrete type here???
// 	result := make([]Persistable,0)

// 	myCollection := ms.db.Collection(coll)
// 	//var cur mongo.Cursor
// 	cur, err := myCollection.Find(context.Background(), nil)
// 	if err != nil { 
// 		ms.logger.Printf("FetchAllFromCollection: %s on Find attempt from %s", err.Error(), coll)
// 		return nil, err
// 	}
// 	defer cur.Close(context.Background())
	
// 	for cur.Next(context.Background()) {
// 		doc := bson.NewDocument()
// 		err = cur.Decode(doc)
// 		if err != nil { 
// 			ms.logger.Printf("FetchAllFromCollection: %s on Decode attempt for %s", err.Error(), doc)
// 			return nil, err
// 		}
// 		var bytes []byte
// 		bytes, err = doc.MarshalBSON()
// 		if err != nil {
// 			ms.logger.Printf("FetchAllFromCollection: %s on MarshalBSON attempt for %s", err.Error(), doc)
// 			return nil, err
// 		}
// 		elem, err := sample.Decode(bytes)
// 		if err != nil {
// 			ms.logger.Printf("FetchAllFromCollection: %s on Persistable.Decode to type %T attempt for %s", err.Error(), sample, bytes)
// 			return nil, err
// 		}
// 		result = append(result, elem)
// 	}

// 	return result, cur.Err()
// }

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
		bson.NewDocument(
			bson.EC.String("_id", id),
		),
	)
	if err != nil {
		ms.logger.Printf("DeleteFromCollection: no mongo connection: %s", err)
	} else if dResult.DeletedCount == 0 {
		err = fmt.Errorf("Delete failed: no documents for id=%s in collection %s", id, coll)
		ms.logger.Printf("DeleteFromCollection: no documents for id=%s in collection %s", id, coll)
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
