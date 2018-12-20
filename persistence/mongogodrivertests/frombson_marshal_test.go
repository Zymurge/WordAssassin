// Copyright (C) MongoDB, Inc. 2017-present.
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at http://www.apache.org/licenses/LICENSE-2.0

package mongogodrivertests

import (
	"bytes"
	"testing"

//	bson "github.com/mongodb/mongo-go-driver/bson"
	bson "github.com/mongodb/mongo-go-driver/x/bsonx"
	"github.com/stretchr/testify/require"
)

func TestMarshal_roundtripFromBytes(t *testing.T) {
	before := []byte{
		// length
		0x1c, 0x0, 0x0, 0x0,

		// --- begin array ---

		// type - document
		0x3,
		// key - "foo"
		0x66, 0x6f, 0x6f, 0x0,

		// length
		0x12, 0x0, 0x0, 0x0,
		// type - string
		0x2,
		// key - "bar"
		0x62, 0x61, 0x72, 0x0,
		// value - string length
		0x4, 0x0, 0x0, 0x0,
		// value - "baz"
		0x62, 0x61, 0x7a, 0x0,

		// null terminator
		0x0,

		// --- end array ---

		// null terminator
		0x0,
	}

	doc := bson.Doc{}
	require.NoError(t, doc.UnmarshalBSON(before))

	after, err := doc.MarshalBSON()
	require.NoError(t, err)

	require.True(t, bytes.Equal(before, after))
}

func TestMarshal_roundtripFromDoc(t *testing.T) {
//	arr := []string{"hey", "whassup", "?"}
	before := bson.Doc {
		{ "foo", bson.String("bar") },
		{ "baz", bson.Int32(-27) },
//		{ "bing", bson.Array(arr) },
	}

	b, err := before.MarshalBSON()
	require.NoError(t, err)

	after := bson.Doc{}
	require.NoError(t, after.UnmarshalBSON(b))

	require.True(t, before.Equal(after))
} 

func noerr(t *testing.T, err error) {
	require.NoError(t, err)
}