// Copyright the Hyperledger Fabric contributors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"encoding/json"
	"log"
	"reflect"
	"testing"

	"github.com/goledgerdev/cc-tools/mock"
	"github.com/stretchr/testify/assert"
)

func TestMockStateRangeQueryIterator(t *testing.T) {
	stub := mock.NewMockStub("rangeTest", nil)
	stub.MockTransactionStart("init")
	stub.PutState("1", []byte{61})
	stub.PutState("0", []byte{62})
	stub.PutState("5", []byte{65})
	stub.PutState("3", []byte{63})
	stub.PutState("4", []byte{64})
	stub.PutState("6", []byte{66})
	stub.MockTransactionEnd("init")

	expectKeys := []string{"3", "4"}
	expectValues := [][]byte{{63}, {64}}

	rqi := mock.NewMockStateRangeQueryIterator(stub, "2", "5")

	// log.Println("Running loop")
	for i := 0; i < 2; i++ {
		response, err := rqi.Next()
		if err != nil {
			log.Println("Loop", i, "got", response.Key, response.Value, err)
		}
		if expectKeys[i] != response.Key {
			log.Println("Expected key", expectKeys[i], "got", response.Key)
			t.FailNow()
		}
		if expectValues[i][0] != response.Value[0] {
			log.Println("Expected value", expectValues[i], "got", response.Value)
		}
	}
}

// TestMockStateRangeQueryIterator_openEnded tests running an open-ended query
// for all keys on the MockStateRangeQueryIterator
func TestMockStateRangeQueryIterator_openEnded(t *testing.T) {
	stub := mock.NewMockStub("rangeTest", nil)
	stub.MockTransactionStart("init")
	stub.PutState("1", []byte{61})
	stub.PutState("0", []byte{62})
	stub.PutState("5", []byte{65})
	stub.PutState("3", []byte{63})
	stub.PutState("4", []byte{64})
	stub.PutState("6", []byte{66})
	stub.MockTransactionEnd("init")

	rqi := mock.NewMockStateRangeQueryIterator(stub, "", "")

	count := 0
	for rqi.HasNext() {
		rqi.Next()
		count++
	}

	if count != rqi.Stub.Keys.Len() {
		t.FailNow()
	}
}

type Marble struct {
	ObjectType string `json:"docType"` //docType is used to distinguish the various types of objects in state database
	Name       string `json:"name"`    //the fieldtags are needed to keep case from bouncing around
	Color      string `json:"color"`
	Size       int    `json:"size"`
	Owner      string `json:"owner"`
}

// JSONBytesEqual compares the JSON in two byte slices.
func jsonBytesEqual(expected []byte, actual []byte) bool {
	var infExpected, infActual interface{}
	if err := json.Unmarshal(expected, &infExpected); err != nil {
		return false
	}
	if err := json.Unmarshal(actual, &infActual); err != nil {
		return false
	}
	return reflect.DeepEqual(infActual, infExpected)
}

func TestGetStateByPartialCompositeKey(t *testing.T) {
	stub := mock.NewMockStub("GetStateByPartialCompositeKeyTest", nil)
	stub.MockTransactionStart("init")

	marble1 := &Marble{"marble", "set-1", "red", 5, "tom"}
	// Convert marble1 to JSON with Color and Name as composite key
	compositeKey1, _ := stub.CreateCompositeKey(marble1.ObjectType, []string{marble1.Name, marble1.Color})
	marbleJSONBytes1, _ := json.Marshal(marble1)
	// Add marble1 JSON to state
	stub.PutState(compositeKey1, marbleJSONBytes1)

	marble2 := &Marble{"marble", "set-1", "blue", 5, "jerry"}
	compositeKey2, _ := stub.CreateCompositeKey(marble2.ObjectType, []string{marble2.Name, marble2.Color})
	marbleJSONBytes2, _ := json.Marshal(marble2)
	stub.PutState(compositeKey2, marbleJSONBytes2)

	marble3 := &Marble{"marble", "set-2", "red", 5, "tom-jerry"}
	compositeKey3, _ := stub.CreateCompositeKey(marble3.ObjectType, []string{marble3.Name, marble3.Color})
	marbleJSONBytes3, _ := json.Marshal(marble3)
	stub.PutState(compositeKey3, marbleJSONBytes3)

	stub.MockTransactionEnd("init")
	// should return in sorted order of attributes
	expectKeys := []string{compositeKey2, compositeKey1}
	expectKeysAttributes := [][]string{{"set-1", "blue"}, {"set-1", "red"}}
	expectValues := [][]byte{marbleJSONBytes2, marbleJSONBytes1}

	rqi, _ := stub.GetStateByPartialCompositeKey("marble", []string{"set-1"})
	// log.Println("Running loop")
	for i := 0; i < 2; i++ {
		response, err := rqi.Next()
		if err != nil {
			log.Println("Loop", i, "got", response.Key, response.Value, err)
		}
		if expectKeys[i] != response.Key {
			log.Println("Expected key", expectKeys[i], "got", response.Key)
			t.FailNow()
		}
		objectType, attributes, _ := stub.SplitCompositeKey(response.Key)
		if objectType != "marble" {
			log.Println("Expected objectType", "marble", "got", objectType)
			t.FailNow()
		}
		// log.Println(attributes)
		for index, attr := range attributes {
			if expectKeysAttributes[i][index] != attr {
				log.Println("Expected keys attribute", expectKeysAttributes[index][i], "got", attr)
				t.FailNow()
			}
		}
		if jsonBytesEqual(expectValues[i], response.Value) != true {
			log.Println("Expected value", expectValues[i], "got", response.Value)
			t.FailNow()
		}
	}
}

func TestGetStateByPartialCompositeKeyCollision(t *testing.T) {
	stub := mock.NewMockStub("GetStateByPartialCompositeKeyCollisionTest", nil)
	stub.MockTransactionStart("init")

	vehicle1Bytes := []byte("vehicle1")
	compositeKeyVehicle1, _ := stub.CreateCompositeKey("Vehicle", []string{"VIN_1234"})
	stub.PutState(compositeKeyVehicle1, vehicle1Bytes)

	vehicleListing1Bytes := []byte("vehicleListing1")
	compositeKeyVehicleListing1, _ := stub.CreateCompositeKey("VehicleListing", []string{"LIST_1234"})
	stub.PutState(compositeKeyVehicleListing1, vehicleListing1Bytes)

	stub.MockTransactionEnd("init")

	// Only the single "Vehicle" object should be returned, not the "VehicleListing" object
	rqi, _ := stub.GetStateByPartialCompositeKey("Vehicle", []string{})
	i := 0
	// log.Println("Running loop")
	for rqi.HasNext() {
		i++
		response, err := rqi.Next()
		if err != nil {
			log.Println("Loop", i, "got", response.Key, response.Value, err)
		}
	}
	// Only the single "Vehicle" object should be returned, not the "VehicleListing" object
	if i != 1 {
		log.Println("Expected 1, got", i)
		t.FailNow()
	}
}

func TestGetTxTimestamp(t *testing.T) {
	stub := mock.NewMockStub("GetTxTimestamp", nil)
	stub.MockTransactionStart("init")

	timestamp, err := stub.GetTxTimestamp()
	if timestamp == nil || err != nil {
		t.FailNow()
	}

	stub.MockTransactionEnd("init")
}

// TestPutEmptyState confirms that setting a key value to empty or nil in the mock state deletes the key
// instead of storing an empty key.
func TestPutEmptyState(t *testing.T) {
	stub := mock.NewMockStub("FAB-12545", nil)

	// Put an empty and nil state value
	stub.MockTransactionStart("1")
	err := stub.PutState("empty", []byte{})
	assert.NoError(t, err)
	err = stub.PutState("nil", nil)
	assert.NoError(t, err)
	stub.MockTransactionEnd("1")

	// Confirm both are nil
	stub.MockTransactionStart("2")
	val, err := stub.GetState("empty")
	assert.NoError(t, err)
	assert.Nil(t, val)
	val, err = stub.GetState("nil")
	assert.NoError(t, err)
	assert.Nil(t, val)
	// Add a value to both empty and nil
	err = stub.PutState("empty", []byte{0})
	assert.NoError(t, err)
	err = stub.PutState("nil", []byte{0})
	assert.NoError(t, err)
	stub.MockTransactionEnd("2")

	// Confirm the value is in both
	stub.MockTransactionStart("3")
	val, err = stub.GetState("empty")
	assert.NoError(t, err)
	assert.Equal(t, val, []byte{0})
	val, err = stub.GetState("nil")
	assert.NoError(t, err)
	assert.Equal(t, val, []byte{0})
	stub.MockTransactionEnd("3")

	// Set both back to empty / nil
	stub.MockTransactionStart("4")
	err = stub.PutState("empty", []byte{})
	assert.NoError(t, err)
	err = stub.PutState("nil", nil)
	assert.NoError(t, err)
	stub.MockTransactionEnd("4")

	// Confirm both are nil
	stub.MockTransactionStart("5")
	val, err = stub.GetState("empty")
	assert.NoError(t, err)
	assert.Nil(t, val)
	val, err = stub.GetState("nil")
	assert.NoError(t, err)
	assert.Nil(t, val)
	stub.MockTransactionEnd("5")

}
