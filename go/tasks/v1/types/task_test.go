package types

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

type myCustomState struct {
	String    string                    `json:"str"`
	Recursive *myCustomState            `json:"recursive"`
	Map       map[string]*myCustomState `json:"map"`
}

func marshalIntoMap(x interface{}) (map[string]interface{}, error) {
	raw, err := json.Marshal(x)
	if err != nil {
		return nil, err
	}

	res := map[string]interface{}{}
	return res, json.Unmarshal(raw, &res)
}

func TestState_marshaling(t *testing.T) {
	expectedMessage := "TestMessage"
	expectedCustom := myCustomState{
		String: expectedMessage,
		Recursive: &myCustomState{
			String: expectedMessage + expectedMessage,
		},
		Map: map[string]*myCustomState{
			"key1": {
				String: expectedMessage,
			},
		},
	}

	rawCustom, err := marshalIntoMap(expectedCustom)
	assert.NoError(t, err)

	raw, err := json.Marshal(&rawCustom)
	assert.NoError(t, err)

	newState := &myCustomState{}
	err = json.Unmarshal(raw, newState)
	assert.NoError(t, err)
	assert.True(t, reflect.DeepEqual(expectedCustom, *newState), "%v != %v", rawCustom, *newState)
	assert.Equal(t, expectedMessage, newState.Map["key1"].String)
}
