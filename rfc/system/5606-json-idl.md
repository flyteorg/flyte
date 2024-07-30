# JSON IDL

**Authors:**

- [@Han-Ru](https://github.com/future-outlier)
- [@Ping-Su](https://github.com/pingsutw)

## 1 Executive Summary

Use json byte string in protobuf instead of json string to fix int is not supported in protobuf struct.


## 2 Motivation

In Flytekit, when handling dataclasses, Pydantic base models, and dictionaries, we store data using JSON strings within struct protobuf. This approach causes issues with integers, as protobuf does not support int types, leading to their conversion to floats. This results in performance issues since we need to recursively iterate through all attributes/keys in dataclasses and dictionaries to ensure floats types are converted to int.

Note: We have more than 10 issues about dict, dataclass and Pydantic.

This feature can solve them all.

## 3 Proposed Implementation
### Flytekit Example
#### Before
```python
@task
def t1() -> dict:
  ...
  return {"a": 1} -> protobuf Struct {"a": 1.0}

@task
def t2(a: <dict>):
  print(a["integer"]) # wrong, will be float point
```
#### After
```python
Json = "json"

@task
def t1() -> Annotated[dict, Json]: // Json Byte Strings
  ...
  return {"a": 1} -> protobuf Json b'{"a": 1}'

@task
def t2(a: <dict>):
  print(a["integer"]) # wrong
```

#### Note
- We use Annotated[dict, Json]  instead of dict to ensure backward compatibility.
    - This helps us avoid breaking changes.
- It makes it easier for the frontend to support JSON IDL after these features are merged.
- If users only upgrade Flytekit, we can ensure they won’t face error when using dict only.
(Since we have to upgrade both flytepropeller, flyteidl and flytekit to support JSON IDL.)


### How to create a byte string?
#### Use MsgPack to convert value to a byte string
##### Python
```python
import msgpack
# Encode
def to_literal():
  json_bytes = msgpack.dumps(v)
  return Literal(scalar=Scalar(json=Json(json_bytes)), metadata={"format": "json"})
# Decode
def to_python_value():
  json_bytes = lv.scalar.json.value
  return msgpack.loads(json_bytes)
```
reference: https://github.com/msgpack/msgpack-python 

##### Golang
```go
package main

import (
    "fmt"
    "github.com/vmihailenco/msgpack/v5"
)

func main() {
    // Example data to encode
    data := map[string]int{"a": 1}

    // Encode the data
    encodedData, err := msgpack.Marshal(data)
    if err != nil {
        panic(err)
    }

    // Print the encoded data
    fmt.Printf("Encoded data: %x\n", encodedData) // Output: 81a16101

    // Decode the data
    var decodedData map[string]int
    err = msgpack.Unmarshal(encodedData, &decodedData)
    if err != nil {
        panic(err)
    }

    // Print the decoded data
    fmt.Printf("Decoded data: %+v\n", decodedData) // Output: map[a:1]
}
```

reference: https://github.com/vmihailenco/msgpack 

##### JavaScript
```javascript
import msgpack5 from 'msgpack5';

// Create a MessagePack instance
const msgpack = msgpack5();

// Example data to encode
const data = { a: 1 };

// Encode the data
const encodedData = msgpack.encode(data);

// Print the encoded data
console.log(encodedData); // <Buffer 81 a1 61 01>

// Decode the data
const decodedData = msgpack.decode(encodedData);

// Print the decoded data
console.log(decodedData); // { a: 1 }
```
reference: https://github.com/msgpack/msgpack-javascript 


### FlyteIDL
```proto
message Json {
    bytes value = 1;
}

message Scalar {
    oneof value {
        Primitive primitive = 1;
        Blob blob = 2;
        Binary binary = 3;
        Schema schema = 4;
        Void none_type = 5;
        Error error = 6;
        google.protobuf.Struct generic = 7;
        StructuredDataset structured_dataset = 8;
        Union union = 9;
        Json json = 10; // New Type
    }
}
```

### FlytePropeller
1. Attribute Access for dictionary, Datalcass, and Pydantic in workflow.
Dict[type, type] is supported already, we have to support Datalcass and Pydantic now.
```python
from flytekit import task, workflow
from dataclasses import dataclass

@dataclass
class DC:
    a: int

@task
def t1() -> DC:
    return DC(a=1)

@task
def t2(x: int):
    print("x:", x)
    return

@workflow
def wf():
  o = t1()
  t2(x=o.a)
```
2. Create a Literal Type for Scalar when doing type validation.
```go
func literalTypeForScalar(scalar *core.Scalar) *core.LiteralType {
  ...
  case *core.Scalar_Json:
		literalType = &core.LiteralType\
			{Type: &core.LiteralType_Simple{Simple: core.SimpleType_JSON}}
  ...
  return literalType 
}
```
3. Support input and default input
```go
// Literal Input
func ExtractFromLiteral(literal *core.Literal) (interface{}, error) {
    switch literalValue := literal.Value.(type) {
        case *core.Literal_Scalar:
        ...
            case *core.Scalar_Json:
                return scalarValue.Json.Value, nil
    }
}
// Default Input
func MakeDefaultLiteralForType(typ *core.LiteralType) (*core.Literal, error) {
    switch t := typ.GetType().(type) {
	case *core.LiteralType_Simple:
        ...
        case core.SimpleType_JSON:
                        return &core.Literal{
                            Value: &core.Literal_Scalar{
                                Scalar: &core.Scalar{
                                    Value: &core.Scalar_Json{
                                        Json: &core.Json{
                                            Value: []byte("{}"),
                                        },
                                    },
                                },
                            },
                        }, nil
    }
}
```
### FlyteKit
#### Add class Json as FlyteIdlEntity
```python
class Json(_common.FlyteIdlEntity):
    def __init__(self, value: bytes):
        self._value = value

    @property
    def value(self):
        """
        :rtype: bytes
        """
        return self._value

    def to_flyte_idl(self):
        """
        :rtype: flyteidl.core.literals_pb2.Json
        """
        return _literals_pb2.Json(value=self.value)

    @classmethod
    def from_flyte_idl(cls, proto):
        """
        :param flyteidl.core.literals_pb2.Json proto:
        :rtype: Json
        """
        return cls(value=proto.value)
```
#### Dict Transformer
##### Before
###### Convert Python Value to Literal
- Dict[type, type] uses type hints to construct a LiteralMap.
- `dict` uses `json.dumps` to turn a `dict` value to a JSON string, and store it to protobuf Struct .
###### Convert Literal to Python Value
- `Dict[type, type]` uses type hints to convert LiteralMap to Python Value.
- `dict` uses `json.loads` to turn a JSON string to a dict value and store it to protobuf Struct .
##### After
###### Convert Python Value to Literal
- `Dict[type, type]` stays the same.
- `dict` uses `msgpack.dumps` to turn a dict value to a JSON byte string, and store is to protobuf Json .

###### Convert Literal to Python Value

`Dict[type, type]` uses type hints to convert LiteralMap to Python Value.

`dict` uses `msgpack.loads` to turn a JSON byte string to a `dict` value and store it to protobuf `Struct` .

#### Dataclass Transformer
##### Before
###### Convert Python Value to Literal
Uses `mashumaro JSON Encoder` to turn a dataclass value to a JSON string, and store it to protobuf `Struct` .
Note: For `FlyteTypes`, we will inherit mashumaro `SerializableType` to define our own serialization behvaior, which includes upload file to remote storage.

###### Convert Literal to Python Value
Uses `mashumaro JSON Decoder` to turn a JSON string to a python value, and recursively fixed int attributes to int (it will be float because we stored it in to `Struct`).

Note: For `FlyteTypes`, we will inherit the mashumaro `SerializableType` to define our own serialization behavior, which includes uploading files to remote storage.
##### After
###### Convert Python Value to Literal
Uses `msgpack.dumps()` to turn a Python value into a byte string.
Note: For `FlyteTypes`, we will need to customize serialization behavior by msgpack reference here.

https://github.com/msgpack/msgpack-python?tab=readme-ov-file#packingunpacking-of-custom-data-type 

###### Convert Literal to Python Value

Uses `msgpack.loads()` to turn a byte string into a Python value.

Note: For `FlyteTypes`, we will need to customize deserialization behavior by `msgpack` reference here.

https://github.com/msgpack/msgpack-python?tab=readme-ov-file#packingunpacking-of-custom-data-type 


#### Pydantic Transformer
##### Before
###### Convert Python Value to Literal
Convert `BaseModel` to a JSON string, and then convert it to a Protobuf `Struct`.
###### Convert Literal to Python Value
Convert Protobuf `Struct` to a JSON string and then convert it to a `BaseModel`.
##### After
###### Convert Python Value to Literal
Convert the Pydantic `BaseModel` to a JSON string, then convert the JSON string to a `dictionary`, and finally, convert it to a `byte string` using msgpack.
###### Convert Literal to Python Value
Convert `byte string` to a `dictionary` using `msgpack`, then convert dictionary to a JSON string, and finally, convert it to Pydantic `BaseModel`.

### FlyteCtl
In Flytectl, we can construct input for the execution, so we have to make sure the values we passed to FlyteAdmin can all be constructed to Literal.

https://github.com/flyteorg/flytectl/blob/131d6a20c7db601ca9156b8d43d243bc88669829/cmd/create/serialization_utils.go#L48 

### FlyteConsole
#### Show input/output on FlyteConsole
We will get node’s input output literal value by FlyteAdmin’s API, and get the json byte string in the literal value.

We can use MsgPack dumps the json byte string to a dictionary, and shows it to the flyteconsole.
#### Construct Input
We should use `msgpack.encode` to encode input value and store it to the literal’s json field.



## 4 Metrics & Dashboards

None

## 5 Drawbacks
There's no breaking changes if we use `Annotated[dict, Json]`, but we need to be very careful about will there be any breaking changes.

## 6 Alternatives
Use UTF-8 format to encode and decode, this will be more easier for implementation, but maybe will cause performance issue when using Pydantic Transformer.

## 7 Potential Impact and Dependencies

*Here, we aim to be mindful of our environment and generate empathy towards others who may be impacted by our decisions.*

- *What other systems or teams are affected by this proposal?*
- *How could this be exploited by malicious attackers?*

## 8 Unresolved questions
I am not sure use UTF-8 format or msgpack to encode and decode is a better option.

## 9 Conclusion
Whether use UTF-8 format or msgpack to encode and decode, we will definitely do it.
