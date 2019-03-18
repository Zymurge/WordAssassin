package persistence

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/globalsign/mgo"
	bson "go.mongodb.org/mongo-driver/bson"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

const (
	TestMongoURL   string = "mongodb://localhost:27017"
	TestDbName     string = "testDB"
	TestCollection string = "TestCollection"
	// Static error messasges from mongo
	ClientDisconnectErrText string = "client is disconnected"
	TopologyClosedErrText   string = "topology is closed"
)

type MongoSessionSuite struct {
	suite.Suite
	session *mgo.Session
	logger  *log.Logger
}

//*** Test Assets ***//

// GenericPersistable is created for each time a player is added to the game
type GenericPersistable struct {
	ID          string    `json:"id" bson:"_id"`
	TimeCreated time.Time `json:"timeCreated" bson:"timecreated"`
	Name        string    `json:"name" bson:"name"`
	ANumber		int       `json:"anumber" bson:"anumber"`	  
}

// GetID returns the unique identifer for this event
func (e GenericPersistable) GetID() string {
	return e.ID
}

// GetTimeCreated returns the unique identifer for this event
func (e GenericPersistable) GetTimeCreated() time.Time {
	return time.Now()
}

func (e *GenericPersistable) Decode(raw []byte) error {
	if err := bson.Unmarshal(raw, e); err != nil {
		return err
	}
	return nil
}

func (e GenericPersistable) ToJSON() []byte {
	result, _ := json.Marshal(e)
	return result
}

// Runner for the test suite. Ensures that mongo can be reached at the default location or aborts the suite. The suite provides a
// pre-connected session for its tests to use for setting the DB state via the SetupTest() call.
func TestMongoSessionSuite(t *testing.T) {
	// precondition is that Mongo must be connectable at the default URL for the suite to run
	session, err := mgo.Dial(TestMongoURL)
	if session != nil {
		defer session.Close()
	}
	require.NoErrorf(t, err, "Mongo must be available at %s for this suite to function", TestMongoURL)
	suite.Run(t, new(MongoSessionSuite))
}

func (m *MongoSessionSuite) SetupSuite() {
	m.session = GetMongoClearedCollection(m.T(), TestCollection)
	m.logger = log.New(os.Stderr, "persistence_test: ", log.Ldate|log.Ltime)
}

func (m *MongoSessionSuite) SetupTest() {
	err := ClearMongoCollection(m.T(), m.session, TestCollection)
	m.NoError(err, "Test failed in setup clearing collection. Err: %s", err)
}

func (m *MongoSessionSuite) TestCtorDefaults() {
	result, err := NewMongoSession("mongodb://testURL", "", nil)
	m.NoError(err, "Test failed in creating MongoSession. Err: %s", err)
	m.Equal(DefaultDbName, result.dbName, "DB name should be the default")
	m.Equal(DefaultTimeout, result.timeoutSeconds, "Timeout value should default when not specified")
	m.Equal("MongoSession: ", result.logger.Prefix(), "Logger should default the prefix when no logger passed in")
}

func (m *MongoSessionSuite) TestCtorNonDefaults() {
	result, err := NewMongoSession("mongodb://testURL", "aDB", m.logger, 13)
	m.NoError(err, "Test failed in creating MongoSession. Err: %s", err)
	m.Equal("aDB", result.dbName, "DB name should be as passed")
	m.Equal(time.Duration(13000000000), result.timeoutSeconds, "Timeout value should be as passed")
	m.Equal(m.logger.Prefix(), result.logger.Prefix(), "Logger should use the prefix of passed in logger")
}

func (m *MongoSessionSuite) TestCtorInvalidConnectionStringThrowsError() {
	_, err := NewMongoSession("/Bad URL", TestDbName, m.logger, 3)
	require.Error(m.T(), err, "Should return an error about bad connection string")
	require.Containsf(m.T(), err.Error(), "error parsing uri", "Looking for err message saying it can't handle non-mongodb:// fronted URIs. Instead got %s", err)
}

// TODO: default logger test

func (m *MongoSessionSuite) TestConnectToMongo() {
	testMS, err := NewMongoSession(TestMongoURL, TestDbName, m.logger)
	require.NoError(m.T(), err, "Test failed in creating MongoSession. Err: %s", err)
	err = testMS.ConnectToMongo()
	require.NoError(m.T(), err, "Sucessful connect throws no error. Instead we got %s", err)
}

func (m *MongoSessionSuite) TestConnectToMongoNoConnectionThrowsError() {
 	testMS, err := NewMongoSession("mongodb://Bad.URL", TestDbName, m.logger, 3)
 	require.NoError(m.T(), err, "Test failed in creating MongoSession. Err: %s", err)
	err = testMS.ConnectToMongo()
	require.Error(m.T(), err, "Should return an error when the mongo server can't be found")
	require.Containsf(m.T(), err.Error(), "No DB found", "Looking for err message saying it can't find the server. Instead got %s", err)
}

func (m *MongoSessionSuite) TestWriteCollection() {
	var err error
	var testMS *MongoSession
	testEvent := &GenericPersistable{
		ID:   "13",
		Name: "Fred",
		ANumber: 13,
	}
	testMS, err = NewMongoSession(TestMongoURL, TestDbName, m.logger, 3)
	require.NoError(m.T(), err, "Test failed in creating MongoSession. Err: %s", err)
	err = testMS.ConnectToMongo()
	require.NoError(m.T(), err, "Test failed in connecting MongoSession. Err: %s", err)

	m.T().Run("Positive", func(t *testing.T) {
		actual := GenericPersistable{}
		err = testMS.WriteCollection(TestCollection, testEvent)
		require.NoError(t, err, "Successful write throws no error. Instead we got %s", err)

		actualBytes, fErr := testMS.FetchIDFromCollection(TestCollection, testEvent.GetID())
		require.NoError(t, fErr, "Failed to validate write for id=%s", testEvent.GetID())
		err = actual.Decode(actualBytes)
	})
	m.T().Run("DuplicateInsertShouldError", func(t *testing.T) {
		dupTestEvent := testEvent
		dupTestEvent.ID = "14"
		err = AddToMongoCollection(t, m.session, TestCollection, dupTestEvent)
		require.NoError(t, err, "Test failed in setup adding to collection. Err: %s", err)

		err = testMS.WriteCollection(TestCollection, testEvent)
		require.Error(t, err, "Attempt to insert duplicate event should throw")
		require.Contains(t, err.Error(), "duplicate", "Expect error text to mention this")
	})
	m.T().Run("CollectionNotExistShouldStillWrite", func(t *testing.T) {
		testBadCollection := "garbage"
		ClearMongoCollection(t, m.session, testBadCollection)

		err = testMS.WriteCollection(testBadCollection, testEvent)
		require.NoErrorf(t, err, "Writes should create collection on the fly. Got err: %s", err)
		writeCount, _ := m.session.DB(TestDbName).C(testBadCollection).Count()
		require.True(t, writeCount == 1, "Record should have been written as only entry")
	})
	m.T().Run("Dropped connection should recover", func(t *testing.T) {
		brokenSession, _ := GetMongoSessionWithLogger(t)
		err = brokenSession.session.Disconnect(context.Background())
		err = brokenSession.WriteCollection(TestCollection, 
			&GenericPersistable {
				ID:   "-999",
				Name: "Someone",
				ANumber: 13,
			},
		)
		require.NoError(t, err, "Should get an error if changed to unreachable URL")
	})
}

func (m *MongoSessionSuite) TestDeleteFromCollection() {
	var err error
	var testMS *MongoSession
	testEvent := &GenericPersistable{
		ID:          "-13",
	//	TimeCreated: time.Unix(63667134976, 53).UTC(),
		Name:        "@wilma.f",
		ANumber:	-13,
	}
	testMS, err = NewMongoSession(TestMongoURL, TestDbName, m.logger, 3)
	require.NoError(m.T(), err, "Test failed in creating MongoSession. Err: %s", err)
	err = testMS.ConnectToMongo()
	require.NoError(m.T(), err, "Test failed in connecting MongoSession. Err: %s", err)

	m.T().Run("Positive", func(t *testing.T) {
		err = AddToMongoCollection(t, m.session, TestCollection, testEvent)
		require.NoError(t, err, "Test failed in setup adding to collection. Err: %s", err)

		err = testMS.DeleteFromCollection(TestCollection, testEvent.GetID())
		require.NoError(t, err, "Successful deletions throw no errors. But this threw: %s", err)
		// validate that it's actually deleted now
		// TODO: fix fetchone for more complete mesaging
		// expectedMsg := fmt.Sprintf("no documents for id=%s", testEvent.GetID())
		expectedMsg := fmt.Sprintf("no documents")
		_, fErr := testMS.FetchIDFromCollection(TestCollection, testEvent.GetID())
		require.Error(t, fErr, "A deleted id should throw an error on fetch")
		require.Contains(t, fErr.Error(), expectedMsg, "The error should be a 'not found' message. Instead it is %s", fErr)
	})
	m.T().Run("Missing ID", func(t *testing.T) {
		testID := "I don't exist"
		expectedMsg := fmt.Sprintf("no documents for id=%s", testID)

		err = testMS.DeleteFromCollection(TestCollection, testID)
		require.Error(t, err, "Delete on missing ID should throw error")
		require.Containsf(t, err.Error(), expectedMsg, "should specify why it threw on missing ID")
	})
	m.T().Run("CollectionNotExist", func(t *testing.T) {
		testID := "it matters not"
		expectedMsg := fmt.Sprintf("no documents for id=%s", testID)
		testBadCollection := "garbage"
		err = testMS.DeleteFromCollection(testBadCollection, testID)
		require.Error(t, err, "Should get error message when attempt to access non-existent collection")
		require.Contains(t, err.Error(), expectedMsg, "Looking for the not found phrase, but got: %s", err)
	})
	m.T().Run("Dropped connection should recover", func(t *testing.T) {
		testID := "any ID will do"
		expectedMsg := fmt.Sprintf("no documents for id=%s", testID)

		brokenSession, _ := GetMongoSessionWithLogger(t)
		err = brokenSession.session.Disconnect(context.Background())
		err = brokenSession.DeleteFromCollection(TestCollection, testID)
		require.Error(t, err, "Should get an error for missing ID to prove it hit the server")
		require.Contains(t, err.Error(), expectedMsg, "Should complain about missing ID")
	})
}

func (m *MongoSessionSuite) TestUpdateCollection() {
	var err error
	var testMS *MongoSession
	testEvent := &GenericPersistable{
		ID:          "-13",
		TimeCreated: time.Unix(63667134985, 13).UTC(),
		Name:        "@wilma.f",
		ANumber:	-13,
	}
	testMS, err = NewMongoSession(TestMongoURL, TestDbName, m.logger, 3)
	require.NoError(m.T(), err, "Test failed in creating MongoSession. Err: %s", err)
	err = testMS.ConnectToMongo()
	require.NoError(m.T(), err, "Test failed in connecting MongoSession. Err: %s", err)

	m.T().Run("Positive", func(t *testing.T) {
		err = AddToMongoCollection(t, m.session, TestCollection, testEvent)
		require.NoError(m.T(), err, "Test failed in setup adding to collection. Err: %s", err)
		updateEvent := testEvent
		expected := "@new.name"
		updateEvent.Name = expected
		var actual GenericPersistable

		err = testMS.UpdateCollection(TestCollection, updateEvent)
		require.NoError(t, err, "Successful update throws no error. Instead we got %s", err)
		var resultBytes []byte
		resultBytes, err = testMS.FetchIDFromCollection(TestCollection, updateEvent.GetID())
		require.NoError(t, err, "Failed to fetch result: %s", err)
		err = actual.Decode(resultBytes)
		require.NoError(t, err, "Failed to decode result bytes: %s", err)
		require.Equal(t, expected, actual.Name)
	})
	m.T().Run("MissingID", func(t *testing.T) {
		badIDEvent := &GenericPersistable{}
		*badIDEvent = *testEvent
		badIDEvent.ID = "I b missing"

		err = testMS.UpdateCollection(TestCollection, badIDEvent)
		require.Error(t, err, "Missing ID should error on update")
		require.Contains(t, err.Error(), "no documents", "Looking for message about ID missing, but got: %s", err)
	})
	m.T().Run("CollectionNotExist", func(t *testing.T) {
		// ensure that the target collection does not exist by iterating through all and deleting if found
		testBadCollection := "garbage"
		collections, _ := m.session.DB(TestDbName).CollectionNames()
		for _, v := range collections {
			if v == testBadCollection {
				err := m.session.DB(TestDbName).C(testBadCollection).DropCollection()
				require.NoError(t, err, "Test failed in setup dropping test collection. Err: %s", err)
			}
		}

		err = testMS.UpdateCollection(testBadCollection, testEvent)
		require.Error(t, err, "Should get error message when attempt to access non-existent collection")
		require.Contains(t, err.Error(), "no documents", "Looking for message about not found, but got: %s", err)
	})
	m.T().Run("Dropped connection recovers", func(t *testing.T) {
		brokenSession, _ := GetMongoSessionWithLogger(t)
		err = brokenSession.session.Disconnect(context.Background())
		err = brokenSession.UpdateCollection(TestCollection, testEvent)
		require.NoError(t, err, "Should not get an error on a valid update after reconnect")
	})
}

func (m *MongoSessionSuite) TestFetchIDFromCollection() {
	var err error
	var testMS *MongoSession
	testEvent := &GenericPersistable{
		ID:          "31",
		TimeCreated: time.Unix(63667135112, 0),
		Name:        "Barney",
		ANumber:	 31,
	}
	// Shared setup
	err = AddToMongoCollection(m.T(), m.session, TestCollection, testEvent)
	require.NoError(m.T(), err, "Test failed in setup adding to collection. Err: %s", err)
	testMS, err = NewMongoSession(TestMongoURL, TestDbName, m.logger, 3)
	require.NoError(m.T(), err, "Test failed in creating MongoSession. Err: %s", err)
	err = testMS.ConnectToMongo()
	require.NoError(m.T(), err, "Test failed in connecting MongoSession. Err: %s", err)

	m.T().Run("Positive", func(t *testing.T) {
		result := GenericPersistable{}
		resultBytes, fErr := testMS.FetchIDFromCollection(TestCollection, testEvent.GetID())
		require.NoError(t, fErr, "Successful lookup throws no error. Instead we got %s", err)
		require.NotNil(t, resultBytes, "Successful lookup has to actually return something")
		err = result.Decode(resultBytes)
		require.NoError(t, err, "Successful decode throws no error. Instead we got %s", err)
		require.Equal(t, testEvent.GetID(), result.GetID())
		require.Equal(t, testEvent.TimeCreated, result.TimeCreated)
		require.Equal(t, testEvent.Name, result.Name)
	})
	m.T().Run("Missing ID", func(t *testing.T) {
		badID := "I an I bad, mon"
		_, fErr := testMS.FetchIDFromCollection(TestCollection, badID)
		require.Error(t, fErr, "Missing id should throw an error")
		require.Contains(t, fErr.Error(), "no documents", "Message should give a clue. Instead it is %s", fErr)
	})
	m.T().Run("Dropped connection should recover", func(t *testing.T) {
		brokenSession, _ := GetMongoSessionWithLogger(t)
		err = brokenSession.session.Disconnect(context.Background())
		_, fErr := brokenSession.FetchIDFromCollection(TestCollection, testEvent.GetID())
		require.NoError(t, fErr)
	})
}

func (m *MongoSessionSuite) TestFetchAllFromCollection() {
	// Shared setup to populate a set of persistables
	var err error
	testEvents := []GenericPersistable{
		{
			ID:          "31",
			Name:        "Barney",
			TimeCreated: time.Unix(0, 0),
			ANumber:	 21,
		},
		{
			ID:          "41",
			Name:        "Betty",
			TimeCreated: time.Unix(50000, 0),
			ANumber:	 21,
		},
		{
			ID:          "51",
			Name:        "Bam Bam",
			TimeCreated: time.Unix(3000000, 0),
			ANumber:	 44,
		},
	}
	for _, v := range testEvents {
		err = AddToMongoCollection(m.T(), m.session, TestCollection, v)
		require.NoError(m.T(), err, "Test failed in setup adding to collection. Err: %s", err)
	}
	var testMS *MongoSession
	testMS, err = NewMongoSession(TestMongoURL, TestDbName, m.logger, 3)
	require.NoError(m.T(), err, "Test failed in creating MongoSession. Err: %s", err)
	err = testMS.ConnectToMongo()
	require.NoError(m.T(), err, "Test failed in connecting MongoSession. Err: %s", err)

	m.T().Run("Positive", func(t *testing.T) {
		results, fErr := testMS.FetchAllFromCollection(TestCollection)
		require.NoError(t, fErr, "Successful lookup throws no error. Instead we got %s", fErr)
		require.NotNil(t, results, "Successful lookup has to actually return something")
		require.Equal(t, len(testEvents), len(results))
		// convert to original type then test further
		for i,b := range results {
			actual := GenericPersistable{}
			dErr := actual.Decode(b)
			require.NoError(t, dErr, "Successful decode throws no error. Instead we got %s", dErr)
			require.IsType(t, GenericPersistable{}, actual)
			require.Equal(t, testEvents[i].GetID(), actual.GetID())
			require.Equal(t, testEvents[i].TimeCreated, actual.TimeCreated)
		}
	})
	// TODO: UT failure scenarios

}

/*** Helper functions ***/

// MongoClearCollection drops the specified collection. Depends on constants for TestMongoURL and DbName.
// Hardcodes 2 second timeout on connect, since it expects local mongo to work
func MongoClearCollection(collName string) error {
	to := 2 * time.Second
	session, err := mgo.DialWithTimeout(TestMongoURL, to)
	defer session.Close()
	if err != nil {
		return err
	}
	myCollection := session.DB(TestDbName).C(collName)
	_, err = myCollection.RemoveAll(nil)
	return err
}

// GetMongoClearedCollection clears the specified collection and returns an active session pointing to it
func GetMongoClearedCollection(t *testing.T, collName string) (session *mgo.Session) {
	var err error
	session, err = mgo.Dial(TestMongoURL)
	if err != nil {
		t.Errorf("GetMongoClearedCollection failed to connect to Mongo")
	}
	myCollection := session.DB(TestDbName).C(collName)
	_, err = myCollection.RemoveAll(nil)
	if err != nil {
		t.Errorf("GetMongoClearedCollection failed to clear collection %s: %s", collName, err)
	}
	return session
}

func ClearMongoCollection(t *testing.T, session *mgo.Session, collName string) error {
	var err error
	clearMe := session.DB(TestDbName).C(collName)
	_, err = clearMe.RemoveAll(nil)
	if err != nil {
		t.Errorf("ClearMongoCollection failed to clear collection %s: %s", collName, err)
	}
	return err
}

func AddToMongoCollection(t *testing.T, session *mgo.Session, collName string, obj interface{}) error {
	myCollection := session.DB(TestDbName).C(collName)
	return myCollection.Insert(obj)
}

func GetMongoSessionWithLogger(t *testing.T) (ms *MongoSession, logBuf *bytes.Buffer) {
	logBuf = &bytes.Buffer{}
	logLabel := "persistence_test: "
	blog := log.New(logBuf, logLabel, 0)
	var err error
	ms, err = NewMongoSession(TestMongoURL, TestDbName, blog, 3)
	require.NoError(t, err, "Test failed in creating MongoSession. Err: %s", err)
	err = ms.ConnectToMongo()
	require.NoError(t, err, "Test failed in connecting MongoSession. Err: %s", err)
	require.NotNil(t, ms, "Test failed in setting up session")
	return
}

