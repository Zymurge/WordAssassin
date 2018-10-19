package mongogodrivertests

import (
	"github.com/mongodb/mongo-go-driver/bson"
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
		docToBytes(bson.NewDocument(bson.EC.Boolean("foo", false))),
	},
	{
		"bigger struct",
		nil,
		struct {
			aString string
			anInt64 int64
		}{aString: "string value", anInt64: 131313},
		docToBytes(bson.NewDocument(
			bson.EC.String("aString", "string value"),
			bson.EC.Int64("anInt64", 131313),
			//EC.Time()
		)),
	},
}

func docToBytes(doc *bson.Document) []byte {
	result, err := doc.MarshalBSON()
	if err != nil { panic(err) }
	return result
}

