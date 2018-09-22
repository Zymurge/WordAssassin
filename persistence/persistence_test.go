package persistence

import (
	"fmt"
	"bytes"
	"context"
	"log"
	"encoding/json"
	"os"
	"time"
	"testing"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	mgo "github.com/globalsign/mgo"
	bson "github.com/globalsign/mgo/bson"
)

const (
	TestMongoURL   string = "mongodb://localhost:27017"
	TestDbName     string = "testDB"
	TestCollection string = "TestCollection"
)

type MongoSessionSuite struct {
	suite.Suite
	session *mgo.Session
	logger  *log.Logger
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
	result, err := NewMongoSession("mongodb://testURL", "", m.logger)
	m.NoError(err, "Test failed in creating MongoSession. Err: %s", err)
	m.EqualValues(DefaultDbName, result.dbName, "DB name should be the default")
	m.EqualValues(DefaultTimeout, result.timeoutSeconds, "Timeout value should default when not specified")
}

func (m *MongoSessionSuite) TestCtorInvalidConnectionStringThrowsError() {
	_, err := NewMongoSession("/Bad URL", TestDbName, m.logger, 3)
	require.Error(m.T(), err, "Should return an error about bad connection string")
	require.Containsf(m.T(), err.Error(), "error parsing uri", "Looking for err message saying it can't handle non-mongodb:// fronted URIs. Instead got %s", err)
}

func (m *MongoSessionSuite) TestConnectToMongo() {
	testMS, err := NewMongoSession(TestMongoURL, TestDbName, m.logger)
	require.NoError(m.T(), err, "Test failed in creating MongoSession. Err: %s", err)
	err = testMS.ConnectToMongo()
	require.NoError(m.T(), err, "Sucessful connect throws no error. Instead we got %s", err)
}

/* TODO: figure out why a non-mongo URI doesn't error on connect and test accordingly
// func (m *MongoSessionSuite) TestConnectToMongoNoConnectionThrowsError() {
// 	testMS, err := NewMongoSession("mongodb://Bad.URL", TestDbName, m.logger, 3)
// 	require.NoError(m.T(), err, "Test failed in creating MongoSession. Err: %s", err)
// 	err = testMS.ConnectToMongo()
// 	require.Error(m.T(), err, "Should return an error when the mongo server can't be found")
// 	require.Containsf(m.T(), err.Error(), "No DB found", "Looking for err message saying it can't find the server. Instead got %s", err)
// }
*/

// NOTE: FIXED
func (m *MongoSessionSuite) TestWriteCollection() {
	var err error
	var testMS *MongoSession
	testEvent := &testGenericPersistable{
		ID:   "13",
		Name: "Fred",
	}
	testMS, err = NewMongoSession(TestMongoURL, TestDbName, m.logger, 3)
	require.NoError(m.T(), err, "Test failed in creating MongoSession. Err: %s", err)

	m.T().Run("Positive", func(t *testing.T) {
		err = testMS.WriteCollection(TestCollection, testEvent)
		require.NoError(t, err, "Successful write throws no error. Instead we got %s", err)

		actual := &testGenericPersistable{}
		err = testMS.FetchOneFromCollection(TestCollection, testEvent.GetID(), actual)
		require.NoError(t, err, "Failed to validate write for id=%s", testEvent.GetID())
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
	m.T().Run("Dropped connection", func(t *testing.T) {
		brokenSession, logBuf := GetMongoSessionWithLogger(t)
		err = brokenSession.session.Disconnect(context.Background())
		err = brokenSession.WriteCollection(TestCollection, testEvent)
		require.Error(t, err, "Should get an error if changed to unreachable URL")
		require.Contains(t, err.Error(), "topology is closed", "Should complain about lack of connectivity")
		require.Contains(t, logBuf.String(), "topology is closed", "Log message should complain about lack of connectivity")
		require.Contains(t, logBuf.String(), "WriteCollection", "Log message should inform on source of issue")
	})
}

// NOTE: FIXED
func (m *MongoSessionSuite) TestDeleteFromCollection() {
	var err error
	var testMS *MongoSession
	testEvent := &testGenericPersistable{
		ID:          "-13",
		TimeCreated: time.Unix(63667134976, 53).UTC(),
		Name:        "@wilma.f",
	}
	testMS, err = NewMongoSession(TestMongoURL, TestDbName, m.logger, 3)
	require.NoError(m.T(), err, "Test failed in creating MongoSession. Err: %s", err)

	m.T().Run("Positive", func(t *testing.T) {
		err = AddToMongoCollection(t, m.session, TestCollection, testEvent)
		require.NoError(t, err, "Test failed in setup adding to collection. Err: %s", err)

		err = testMS.DeleteFromCollection(TestCollection, testEvent.GetID())
		require.NoError(t, err, "Successful deletions throw no errors. But this threw: %s", err)
		// validate that it's actually deleted now
		actual := &testGenericPersistable{}
		// TODO: fix fetchone for more complete mesaging
		// expectedMsg := fmt.Sprintf("no documents for id=%s", testEvent.GetID())
		expectedMsg := fmt.Sprintf("no documents")
		err = testMS.FetchOneFromCollection(TestCollection, testEvent.GetID(), actual)
		require.Error(t, err, "A deleted id should throw an error on fetch")
		require.Contains(t, err.Error(), expectedMsg, "The error should be a 'not found' message. Instead it is %s", err)
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
	m.T().Run("Dropped connection", func(t *testing.T) {
		brokenSession, logBuf := GetMongoSessionWithLogger(t)
		err = brokenSession.session.Disconnect(context.Background())
		err = brokenSession.DeleteFromCollection(TestCollection, "any ID will do")
		require.Error(t, err, "Should get an error if changed to unreachable URL")
		require.Contains(t, err.Error(), "topology is closed", "Should complain about lack of connectivity")
		require.Contains(t, logBuf.String(), "topology is closed", "Log message should complain about lack of connectivity")
		require.Contains(t, logBuf.String(), "DeleteFromCollection", "Log message should inform on source of issue")
	})
}

// NOTE: All func tests working. Use setup as model.
// TODO: solve diconnect test
func (m *MongoSessionSuite) TestUpdateCollection() {
	var err error
	var testMS *MongoSession
	testEvent := &testGenericPersistable{
		ID:          "-13",
		TimeCreated: time.Unix(63667134985, 13).UTC(),
		Name:        "@wilma.f",
	}
	testMS, err = NewMongoSession(TestMongoURL, TestDbName, m.logger, 3)
	require.NoError(m.T(), err, "Test failed in creating MongoSession. Err: %s", err)

	m.T().Run("Positive", func(t *testing.T) {
		err = AddToMongoCollection(t, m.session, TestCollection, testEvent)
		require.NoError(m.T(), err, "Test failed in setup adding to collection. Err: %s", err)
		updateEvent := testEvent
		expected := "@new.name"
		updateEvent.Name = expected

		err = testMS.UpdateCollection(TestCollection, updateEvent)
		require.NoError(t, err, "Successful update throws no error. Instead we got %s", err)
		actual := &testGenericPersistable{}
		if err := testMS.FetchOneFromCollection(TestCollection, updateEvent.GetID(), actual); err != nil {
			require.NoError(t, err, "Failed to fetch result: %s", err.Error())
		}
		require.Equal(t, expected, actual.Name)
	})
	m.T().Run("MissingID", func(t *testing.T) {
		badIDEvent := testEvent
		badIDEvent.ID = "I b missing"

		err = testMS.UpdateCollection(TestCollection, badIDEvent)
		require.Error(t, err, "Missing ID should error on update")
		require.Contains(t, err.Error(), "not found", "Looking for message about ID missing, but got: %s", err)
	})
	m.T().Run("CollectionNotExist", func(t *testing.T) {
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
		require.Contains(t, err.Error(), "not found", "Looking for message about not found, but got: %s", err)
	})
	m.T().Run("Dropped connection", func(t *testing.T) {
		brokenSession, logBuf := GetMongoSessionWithLogger(t)
		err = brokenSession.session.Disconnect(context.Background())
		err = brokenSession.UpdateCollection(TestCollection, testEvent)
		require.Error(t, err, "Should get an error if changed to unreachable URL")
		require.Contains(t, err.Error(), "topology is closed", "Should complain about lack of connectivity")
		require.Contains(t, logBuf.String(), "topology is closed", "Log message should complain about lack of connectivity")
		require.Contains(t, logBuf.String(), "UpdateCollection", "Log message should inform on source of issue")
	})
}

// NOTE: FIXED
func (m *MongoSessionSuite) TestFetchOneFromCollection() {
	var err error
	var testMS *MongoSession
	testEvent := &testGenericPersistable{
		ID:          "31",
		Name:        "Barney",
		TimeCreated: time.Unix(63667135112, 0).UTC(),
	}
	// Shared setup
	err = AddToMongoCollection(m.T(), m.session, TestCollection, testEvent)
	require.NoError(m.T(), err, "Test failed in setup adding to collection. Err: %s", err)
	testMS, err = NewMongoSession(TestMongoURL, TestDbName, m.logger, 3)
	require.NoError(m.T(), err, "Test failed in creating MongoSession. Err: %s", err)
	err = testMS.ConnectToMongo()
	require.NoError(m.T(), err, "Test failed in connecting MongoSession. Err: %s", err)

	m.T().Run("Positive", func(t *testing.T) {
		result := &testGenericPersistable{}
		err = testMS.FetchOneFromCollection(TestCollection, testEvent.GetID(), result)
		require.NoError(t, err, "Successful lookup throws no error. Instead we got %s", err)
		require.NotNil(t, result, "Successful lookup has to actually return something")
		require.Equal(t, testEvent.GetID(), result.GetID())
		require.Equal(t, testEvent.TimeCreated, result.TimeCreated)
		require.Equal(t, testEvent.Name, result.Name)
	})
	m.T().Run("Missing ID", func(t *testing.T) {
		var result Persistable
		badID := "I an I bad, mon"
		err = testMS.FetchOneFromCollection(TestCollection, badID, result)
		require.Error(t, err, "Missing id should throw an error")
		require.Contains(t, err.Error(), "no documents", "Message should give a clue. Instead it is %s", err)
	})
	m.T().Run("Dropped connection", func(t *testing.T) {
		var result Persistable
		brokenSession, logBuf := GetMongoSessionWithLogger(t)
		err = brokenSession.session.Disconnect(context.Background())
		err = brokenSession.FetchOneFromCollection(TestCollection, testEvent.GetID(), result)
		require.Error(t, err, "Should get an error if changed to unreachable URL")
		require.Contains(t, err.Error(), "topology is closed", "Should complain about lack of connectivity")
		require.Contains(t, logBuf.String(), "topology is closed", "Log message should complain about lack of connectivity")
		require.Contains(t, logBuf.String(), "FetchOneFromCollection", "Log message should inform on source of issue")
	})
}

func (m *MongoSessionSuite) TestFetchAllFromCollection() {
	// Shared setup to populate a set of persistables
	var err error
	testEvents := []testGenericPersistable{
		{
			ID:          "31",
			Name:        "Barney",
			TimeCreated: time.Unix(63667135112, 0).UTC(),
		},
		{
			ID:          "41",
			Name:        "Betty",
			TimeCreated: time.Unix(63667135112, 0).UTC(),
		},
		{
			ID:          "51",
			Name:        "Bam Bam",
			TimeCreated: time.Unix(63667135112, 0).UTC(),
		},
	}
	for _, v := range testEvents {
		err = AddToMongoCollection(m.T(), m.session, TestCollection, v)
		require.NoError(m.T(), err, "Test failed in setup adding to collection. Err: %s", err)
	}
	var testMS *MongoSession
	testMS, err = NewMongoSession(TestMongoURL, TestDbName, m.logger, 3)
	require.NoError(m.T(), err, "Test failed in creating MongoSession. Err: %s", err)
	require.NotNil(m.T(), testMS)
	/*
		m.T().Run("Positive", func(t *testing.T) {
			results := []bson.Raw{}
			results, err = testMS.FetchAllFromCollection(TestCollection)
			require.NoError(t, err, "Successful lookup throws no error. Instead we got %s", err)
			require.NotNil(t, results, "Successful lookup has to actually return something")
			require.Equal(t, len(testEvents), len(results))
		})
	*/
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

//*** Test Assets ***//

// testGenericPersistable is created for each time a player is added to the game
type testGenericPersistable struct {
	//GameEvent
	ID          string    `json:"id" bson:"_id"`
	TimeCreated time.Time `json:"timeCreated" bson:"timecreated"`
	Name        string    `json:"name" bson:"name"`
}

// GetID returns the unique identifer for this event
func (e *testGenericPersistable) GetID() string {
	return e.ID
}

// GetTimeCreated returns the unique identifer for this event
func (e *testGenericPersistable) GetTimeCreated() time.Time {
	return e.TimeCreated
}

func (e *testGenericPersistable) Decode(raw []byte) error {
	if err := bson.Unmarshal(raw, e); err != nil {
		return err
	}
	return nil
}

func (e *testGenericPersistable) ToJSON() []byte {
	result, _ := json.Marshal(e)
	return result
}
