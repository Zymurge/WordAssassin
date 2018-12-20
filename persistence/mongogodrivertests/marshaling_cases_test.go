package mongogodrivertests

import (
	bson "github.com/mongodb/mongo-go-driver/x/bsonx"
	"github.com/mongodb/mongo-go-driver/bson/bsoncodec"
)

type marshalingTestCase struct {
	name string
	reg  *bsoncodec.Registry
	val  interface{}
	want []byte
}

var marshalingTestCases = []marshalingTestCase{
	{
		"small struct",
		nil,
		struct {
			Foo bool
		}{Foo: false},
		docToBytes( &bson.Doc {
			{ "foo", bson.Boolean(false) },
		}),
	},
	{
		"bigger struct",
		nil,
		struct {
			aString string
			anInt64 int64
		}{aString: "string value", anInt64: 131313},
		docToBytes( &bson.Doc {
			{ "aString", bson.String("string value") },
			{ "anInt64", bson.Int64( 131313 ) },
		},),
	},
}

func docToBytes(doc *bson.Doc) []byte {
	result, err := doc.MarshalBSON()
	if err != nil { panic(err) }
	return result
}

