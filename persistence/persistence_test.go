package persistence

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	mgo "github.com/globalsign/mgo"
	bson "github.com/globalsign/mgo/bson"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

const (
	TestMongoURL   string = "localhost:27017"
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
	result := NewMongoSession("testURL", "", m.logger)
	m.EqualValues(DefaultDbName, result.dbName, "DB name should be the default")
	m.EqualValues(DefaultTimeout, result.timeoutSeconds, "Timeout value should default when not specified")
}

func (m *MongoSessionSuite) TestConnectToMongo() {
	ms := MongoSession{
		mongoURL:       TestMongoURL,
		timeoutSeconds: 3 * time.Second,
	}
	err := ms.ConnectToMongo()
	m.NoError(err, "Sucessful connect throws no error. Instead we got %s", err)
	m.IsType(MongoSession{}, ms, "Wrong type on connect: %T", ms)
}

func (m *MongoSessionSuite) TestConnectToMongoNoConnectionThrowsError() {
	ms := MongoSession{
		mongoURL:       "i.am.abad.url:12345",
		timeoutSeconds: 100 * time.Millisecond,
	}
	err := ms.ConnectToMongo()
	m.Error(err, "Should return an error when the mongo server can't be found")
	m.Containsf(err.Error(), "no reachable", "Looking for err message saying it can't find the server. Instead got %s", err)
}

func (m *MongoSessionSuite) TestWriteCollection() {
	var err error
	testEvent := &testGenericPersistable{
		ID:   "13",
		Name: "Fred",
	}
	m.T().Run("Positive", func(t *testing.T) {
		testMS := NewMongoSession(TestMongoURL, TestDbName, m.logger, 3)
		err = testMS.WriteCollection(TestCollection, testEvent)
		require.NoError(t, err, "Successful write throws no error. Instead we got %s", err)
	})
	m.T().Run("DuplicateInsertShouldError", func(t *testing.T) {
		dupTestEvent := testEvent
		dupTestEvent.ID = "14"
		err = AddToMongoCollection(t, m.session, TestCollection, dupTestEvent)
		require.NoError(t, err, "Test failed in setup adding to collection. Err: %s", err)

		// write the same event again
		testMS := NewMongoSession(TestMongoURL, TestDbName, m.logger, 3)
		err = testMS.WriteCollection(TestCollection, testEvent)
		require.Error(t, err, "Attempt to insert duplicate event should throw")
		require.Contains(t, err.Error(), "duplicate", "Expect error text to mention this")
	})
	m.T().Run("CollectionNotExistShouldStillWrite", func(t *testing.T) {
		testBadCollection := "garbage"
		ClearMongoCollection(t, m.session, testBadCollection)

		testMS := NewMongoSession(TestMongoURL, TestDbName, m.logger, 3)
		err = testMS.WriteCollection(testBadCollection, testEvent)
		require.NoErrorf(t, err, "Writes should create collection on the fly. Got err: %s", err)
		writeCount, _ := m.session.DB(TestDbName).C(testBadCollection).Count()
		require.True(t, writeCount == 1, "Record should have been written as only entry")
	})
	m.T().Run("Dropped connection", func(t *testing.T) {
		testMS, logBuf := GetMongoSessionWithLogger()
		testMS.mongoURL = "yo"

		err = testMS.WriteCollection(TestCollection, testEvent)
		require.Error(t, err, "Should get an error if changed to unreachable URL")
		require.Contains(t, err.Error(), "no reachable servers", "Return value should complain about lack of connectivity")
		require.Contains(t, logBuf.String(), "no reachable servers", "Log message should complain about lack of connectivity")
		require.Contains(t, logBuf.String(), "WriteCollection", "Log message should inform on source of issue")
	})
}

func (m *MongoSessionSuite) TestDeleteFromCollection() {
	var err error
	testEvent := testGenericPersistable{
		ID:          "-13",
		TimeCreated: time.Unix(63667134976, 53).UTC(),
		Name:        "@wilma.f",
	}
	m.T().Run("Positive", func(t *testing.T) {
		err = AddToMongoCollection(t, m.session, TestCollection, testEvent)
		require.NoError(t, err, "Test failed in setup adding to collection. Err: %s", err)

		testMS := NewMongoSession(TestMongoURL, TestDbName, m.logger, 3)
		err = testMS.DeleteFromCollection(TestCollection, testEvent.GetID())
		require.NoError(t, err, "Successful deletions throw no errors. But this threw: %s", err)
	})
	m.T().Run("Missing ID", func(t *testing.T) {
		testID := "I don't exist"

		testMS := NewMongoSession(TestMongoURL, TestDbName, m.logger, 3)
		err = testMS.DeleteFromCollection(TestCollection, testID)
		require.Error(t, err, "Delete on missing ID should throw error")
		require.Containsf(t, err.Error(), "not found", "mgo should specify why it threw on missing ID")
	})
	m.T().Run("CollectionNotExist", func(t *testing.T) {
		testBadCollection := "garbage"
		testMS := NewMongoSession(TestMongoURL, TestDbName, m.logger, 3)
		err = testMS.DeleteFromCollection(testBadCollection, "it matters not")
		require.Error(t, err, "Should get error message when attempt to access non-existent collection")
		require.Contains(t, err.Error(), "not found", "Looking for the not found phrase, but got: %s", err)
	})
	m.T().Run("Dropped connection", func(t *testing.T) {
		testMS, logBuf := GetMongoSessionWithLogger()
		testMS.mongoURL = "yo"
		err = testMS.DeleteFromCollection(TestCollection, "it matters not")
		require.Error(t, err, "Should get an error if changed to unreachable URL")
		require.Contains(t, err.Error(), "no reachable servers", "Should complain about lack of connectivity")
		require.Contains(t, logBuf.String(), "no reachable servers", "Log message should complain about lack of connectivity")
		require.Contains(t, logBuf.String(), "DeleteFromCollection", "Log message should inform on source of issue")
	})
}

func (m *MongoSessionSuite) TestUpdateCollection() {
	var err error
	testEvent := &testGenericPersistable{
		ID:          "-13",
		TimeCreated: time.Unix(63667134985, 13).UTC(),
		Name:        "@wilma.f",
	}
	// Shared setup
	err = AddToMongoCollection(m.T(), m.session, TestCollection, testEvent)
	require.NoError(m.T(), err, "Test failed in setup adding to collection. Err: %s", err)

	m.T().Run("Positive", func(t *testing.T) {
		updateEvent := testEvent
		expected := "@wilma.f"
		updateEvent.Name = expected
		testMS := NewMongoSession(TestMongoURL, TestDbName, m.logger, 3)

		err = testMS.UpdateCollection(TestCollection, updateEvent)
		require.NoError(t, err, "Successful update throws no error. Instead we got %s", err)
		actual := &testGenericPersistable{}
		if err := testMS.FetchFromCollection(TestCollection, updateEvent.GetID(), actual); err != nil {
			require.NoError(t, err, "Failed to fetch result: %s", err.Error())
		}
		require.Equal(t, expected, actual.Name)
	})
	m.T().Run("MissingID", func(t *testing.T) {
		badIDEvent := testEvent
		badIDEvent.ID = "I b missing"
		testMS := NewMongoSession(TestMongoURL, TestDbName, m.logger, 3)

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
		testMS := NewMongoSession(TestMongoURL, TestDbName, m.logger, 3)

		err = testMS.UpdateCollection(testBadCollection, testEvent)
		require.Error(t, err, "Should get error message when attempt to access non-existent collection")
		require.Contains(t, err.Error(), "Non-existent collection for update", "Looking for missing collection, but got: %s", err)
	})
	m.T().Run("Dropped connection", func(t *testing.T) {
		testMS, logBuf := GetMongoSessionWithLogger()
		testMS.mongoURL = "yo"

		err = testMS.UpdateCollection(TestCollection, testEvent)
		require.Error(t, err, "Should get an error if changed to unreachable URL")
		require.Contains(t, err.Error(), "no reachable servers", "Should complain about lack of connectivity")
		require.Contains(t, logBuf.String(), "no reachable servers", "Log message should complain about lack of connectivity")
		require.Contains(t, logBuf.String(), "UpdateCollection", "Log message should inform on source of issue")
	})
}

func (m *MongoSessionSuite) TestFetchFromCollection() {
	var err error
	testEvent := testGenericPersistable{
		ID:          "31",
		Name:        "Barney",
		TimeCreated: time.Unix(63667135112, 0).UTC(),
	}
	// Shared setup
	err = AddToMongoCollection(m.T(), m.session, TestCollection, testEvent)
	require.NoError(m.T(), err, "Test failed in setup adding to collection. Err: %s", err)
	testMS := NewMongoSession(TestMongoURL, TestDbName, m.logger, 3)

	m.T().Run("Positive", func(t *testing.T) {
		result := &testGenericPersistable{}
		err = testMS.FetchFromCollection(TestCollection, testEvent.GetID(), result)
		require.NoError(t, err, "Successful lookup throws no error. Instead we got %s", err)
		require.NotNil(t, result, "Successful lookup has to actually return something")
		require.Equal(t, testEvent.GetID(), result.GetID())
		require.Equal(t, testEvent.TimeCreated, result.TimeCreated)
		require.Equal(t, testEvent.Name, result.Name)
	})
	m.T().Run("Missing ID", func(t *testing.T) {
		var result Persistable
		badID := "I an I bad, mon"
		err = testMS.FetchFromCollection(TestCollection, badID, result)
		require.Error(t, err, "Missing id should throw an error")
		require.Contains(t, err.Error(), "not found", "Message should give a clue. Instead it is %s", err)
	})
	m.T().Run("Dropped connection", func(t *testing.T) {
		var result Persistable
		testMS, logBuf := GetMongoSessionWithLogger()
		testMS.mongoURL = "yo"
		err = testMS.FetchFromCollection(TestCollection, testEvent.GetID(), result)
		require.Error(t, err, "Should get an error if changed to unreachable URL")
		require.Contains(t, err.Error(), "no reachable servers", "Should complain about lack of connectivity")
		require.Contains(t, logBuf.String(), "no reachable servers", "Log message should complain about lack of connectivity")
		require.Contains(t, logBuf.String(), "FetchFromCollection", "Log message should inform on source of issue")
	})
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

func GetMongoSessionWithLogger() (ms *MongoSession, logBuf *bytes.Buffer) {
	logBuf = &bytes.Buffer{}
	logLabel := "persistence_test: "
	blog := log.New(logBuf, logLabel, 0)
	ms = NewMongoSession(TestMongoURL, TestDbName, blog, 3)
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

// Decode populates this instance from the supplied bson
func (e *testGenericPersistable) Decode(b bson.M) error {
	if val, ok := b["_id"]; ok {
		if e.ID, ok = val.(string); !ok {
			return fmt.Errorf("Cast issue for ID")
		}
	} else {
		return fmt.Errorf("Missing tag: _id")
	}
	if val, ok := b["timecreated"]; ok {
		if e.TimeCreated, ok = val.(time.Time); !ok {
			return fmt.Errorf("Cast issue for TimeCreated")
		}
		e.TimeCreated = e.TimeCreated.UTC()
	} else {
		return fmt.Errorf("Missing tag: timecreated")
	}
	if val, ok := b["name"]; ok {
		if e.Name, ok = val.(string); !ok {
			return fmt.Errorf("Cast issue for Name")
		}
	} else {
		return fmt.Errorf("Missing tag: name")
	}

	return nil
}
