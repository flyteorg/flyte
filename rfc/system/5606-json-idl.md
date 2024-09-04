# JSON IDL

**Authors:**

- [@Han-Ru](https://github.com/future-outlier)
- [@Ping-Su](https://github.com/pingsutw)
- [@Fabio M. Graetz](https://github.com/fg91)
- [@Yee Hing Tong](https://github.com/wild-endeavor)

## 1 Executive Summary
- To Literal

| Before                            | Now                                          |
|-----------------------------------|----------------------------------------------|
| JSON String -> Protobuf Struct    | JSON String -> Byte String -> Protobuf Json  |

- To Python Value

| Before                            | Now                                          |
|-----------------------------------|----------------------------------------------|
| Protobuf Struct -> JSON String | Protobuf Json -> Byte String -> JSON String  |

Use byte string in protobuf instead of json string to fix case that int is not supported in protobuf struct.

## 2 Motivation

In Flytekit, when handling dataclasses, Pydantic base models, and dictionaries, we store data using JSON strings within protobuf struct datatype.
This approach causes issues with integers, as protobuf struct does not support int types, leading to their conversion to floats.
This results in performance issues since we need to recursively iterate through all attributes/keys in dataclasses and dictionaries to ensure floats types are converted to int. In addition to performance issues, the required code is complicated and error prone.

Note: We have more than 10 issues about dict, dataclass and Pydantic.

This feature can solve them all.

## 3 Proposed Implementation
### Flytekit Example
#### Before
```python
@task
def t1() -> dict:
  ...
  return {"a": 1} # protobuf Struct {"a": 1.0}

@task
def t2(a: dict):
  print(a["integer"]) # wrong, will be a float
```
#### After
```python
@task
def t1() -> dict: # Json Byte Strings
  ...
  return {"a": 1}  # protobuf Json b'{"a": 1}'

@task
def t2(a: dict):
  print(a["integer"]) # correct, it will be a integer
```

#### Note
- We will use the same type interface and ensure the backward compatibility.

### How to create a byte string?
#### Use MsgPack to convert value to a byte string
##### Python
```python
import msgpack
import json

# Encode
def to_literal():
  json_str = json.dumps(python_val)
  json_bytes = msgpack.dumps(json_str)
  return Literal(scalar=Scalar(json=Json(json_bytes)), metadata={"format": "json"})

# Decode
def to_python_value():
  json_bytes = lv.scalar.json.value
  json_str = msgpack.loads(json_bytes)
  return json.loads(json_str)
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
1. Attribute Access for dictionary, Dataclass, and Pydantic in workflow.
Dict[type, type] is supported already, we have to support Dataclass and Pydantic now.
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
		literalType = &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_JSON}}
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
#### pyflyte run
The behavior will remain unchanged. 
We will pass the value to our class, which inherits from `click.ParamType`, and use the corresponding type transformer to convert the input to the correct type.

### Dict Transformer

| **Stage** | **Conversion** | **Description**                                                                                                                                                                                                                                                                                                               |
| --- | --- |-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **Before** | Python Value to Literal | 1. `Dict[type, type]` uses type hints to construct a LiteralMap. <br> 2. `dict` uses `json.dumps` to turn a `dict` value to a JSON string, and store it to protobuf Struct.                                                                                                                                                   |
| | Literal to Python Value | 1. `Dict[type, type]` uses type hints to convert LiteralMap to Python Value. <br> 2. `dict` uses `json.loads` to turn a JSON string to a dict value and store it to protobuf Struct.                                                                                                                                          |
| **After** | Python Value to Literal | 1. `Dict[type, type]` stays the same. <br> 2. `dict` uses `msgpack.dumps` to turn a JSON string to a byte string, and store is to protobuf Json. |
| | Literal to Python Value | 1. `Dict[type, type]` uses type hints to convert LiteralMap to Python Value. <br> 2.  `dict` conversion: byte string -> JSON string -> dict value, method: `msgpack.loads` -> `json.loads`. |

### Dataclass Transformer

| **Stage** | **Conversion** | **Description**                                                                                                                                                                                                                                                                                                                                                                                                                                                                   |
| --- | --- |-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **Before** | Python Value to Literal | Uses `mashumaro JSON Encoder` to turn a dataclass value to a JSON string, and store it to protobuf `Struct`. |
| | Literal to Python Value | Uses `mashumaro JSON Decoder` to turn a JSON string to a python value, and recursively fixed int attributes to int (it will be float because we stored it in to `Struct`). |
| **After** | Python Value to Literal | Uses `mashumaro JSON Encoder` to turn a dataclass value to a JSON string, and uses `msgpack.dumps()` to turn the JSON string into a byte string, and store it to protobuf `Json`. |
| | Literal to Python Value | Uses `msgpack.loads()` to turn a byte string into a JSON string, and uses `mashumaro JSON Decoder` to turn the JSON string into a Python value. |

### Pydantic Transformer

| **Stage** | **Conversion** | **Description** |
| --- | --- | --- |
| **Before** | Python Value to Literal | Convert `BaseModel` to a JSON string, and then convert it to a Protobuf `Struct`. |
| | Literal to Python Value | Convert Protobuf `Struct` to a JSON string and then convert it to a `BaseModel`. |
| **After** | Python Value to Literal | Convert the Pydantic `BaseModel` to a JSON string, then convert the JSON string to a `byte string` using msgpack. |
| | Literal to Python Value | Convert `byte string` to a JSON string using `msgpack`, then convert it to Pydantic `BaseModel`. |


### FlyteCtl
In FlyteCtl, we can construct input for the execution, so we have to make sure the values we passed to FlyteAdmin 
can all be constructed to Literal.

reference: https://github.com/flyteorg/flytectl/blob/131d6a20c7db601ca9156b8d43d243bc88669829/cmd/create/serialization_utils.go#L48 

### FlyteConsole
#### Show input/output on FlyteConsole
We will get node’s input output literal value by FlyteAdmin’s API, and get the json byte string in the literal value.

We can use MsgPack dumps the json byte string to a dictionary, and shows it to the flyteconsole.
#### Construct Input
We should use `msgpack.encode` to encode input value and store it to the literal’s json field.


## 4 Metrics & Dashboards

None

## 5 Drawbacks
None

## 6 Alternatives
None, it's doable.

## 7 Potential Impact and Dependencies

*Here, we aim to be mindful of our environment and generate empathy towards others who may be impacted by our decisions.*

- *What other systems or teams are affected by this proposal?*
- *How could this be exploited by malicious attackers?*

## 8 Unresolved questions
MsgPack is better because it's more smaller and faster.

## 9 Conclusion
We will use msgpack to do it.
