package mongogodrivertests

/*
	This test is in place while mongo-go-driver is in alpha. The goal is to verify that the bsoncodec
	logic is working or not absent of any wordassassin implementation
*/

/*
import (
//	bson "github.com/mongodb/mongo-go-driver/bson"
	bson "github.com/mongodb/mongo-go-driver/x/bsonx"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

type TestStruct struct {
	ID      string    `bson:"_id"`
	aString string    `bson:"astring"`
	anInt   int32     `bson:"anint"`
	aTime   time.Time `bson:"atime"`
}


func TestShowTimeUnmarshalError(t *testing.T) {
	testTime := time.Now()
	myStruct := TestStruct{"myID", "a string", 13, testTime}
	var result TestStruct

	myDoc := bson.Doc {
		{ "_id", bson.String(myStruct.ID) },
		{ "astring", bson.String(myStruct.aString) },
		{ "anInt", bson.Int32(myStruct.anInt) },
		{ "aTime", bson.Time(myStruct.aTime) },
	}
	docBytes := docToBytes( &myDoc )

	myBSON := bson.Doc{}

	t.Run("Marshal from Document and interface match", func(t *testing.T) {
		require.Equal(t, docBSON, marshalBSON)
	})
	t.Run("Unmarshal works for string and int", func(t *testing.T) {
		require.Equal(t, myStruct.ID, result.ID)           // passes
		require.Equal(t, myStruct.anInt, result.anInt)     // passes
		require.Equal(t, myStruct.aString, result.aString) // passes
	})
	t.Run("Unmarshal works for time", func(t *testing.T) {
		require.Equal(t, myStruct.aTime, result.aTime) // fails due to rounding on nsec
	})
}


//// Helpers ////

func docToBytes(doc *bson.Doc) []byte {
	result, err := doc.MarshalBSON()
	if err != nil { panic(err) }
	return result
}

*/