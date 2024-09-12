# Binary IDL With MessagePack Bytes

**Authors:**

- [@Han-Ru](https://github.com/future-outlier)
- [@Yee Hing Tong](https://github.com/wild-endeavor)
- [@Ping-Su](https://github.com/pingsutw)
- [@Eduardo Apolinario](https://github.com/eapolinario)
- [@Haytham Abuelfutuh](https://github.com/EngHabu)
- [@Ketan Umare](https://github.com/kumare3)

## 1 Executive Summary
### Literal Value
Literal Value will be `Binary`.
Use `bytes` in `Binary` instead of `Protobuf struct`.

- To Literal

| Before                            | Now                                          |
|-----------------------------------|----------------------------------------------|
| Python Val -> JSON String -> Protobuf Struct | Python Val -> (Dict ->) Bytes -> Binary (value: MessagePack Bytes, tag: msgpack) IDL Object |

- To Python Value

| Before                            | Now                                          |
|-----------------------------------|----------------------------------------------|
| Protobuf Struct -> JSON String -> Python Val | Binary (value: MessagePack Bytes, tag: msgpack) IDL Object -> Bytes -> (Dict ->) -> Python Val |

Note: if a python value can't directly be converted to `MessagePack Bytes`, we can convert it to `Dict`, and then convert it to `MessagePack Bytes`.

For example, the pydantic to literal function will be `BaseModel` -> `dict` -> `MessagePack Bytes` -> `Binary (value: MessagePack Bytes, tag: msgpack) IDL Object`.

For pure `dict` in python, the to literal function will be `dict` -> `MessagePack Bytes` -> `Binary (value: MessagePack Bytes, tag: msgpack) IDL Object`.

### Literal Type
Literal Type will be `Protobuf struct`.
`Json Schema` will be stored in `Literal Type's metadata`.

1. Dataclass, Pydantic BaseModel and pure dict in python will all use `Protobuf Struct`.
2. We will put `Json Schema` in Literal Type's `metadata` field, this will be used in flytekit remote api to construct dataclass/Pydantic BaseModel by `Json Schema`.
3. We will use libraries written in golang to compare `Json Schema` to solve this issue: ["[BUG] Union types fail for e.g. two different dataclasses"](https://github.com/flyteorg/flyte/issues/5489).


Note: The `metadata` of `Literal Type` and `Literal Value` are not the same.

## 2 Motivation

In Flytekit, when handling dataclasses, Pydantic base models, and dictionaries, we store data using a JSON string within Protobuf struct datatype.
This approach causes issues with integers, as Protobuf struct does not support int types, leading to their conversion to floats.
This results in performance issues since we need to recursively iterate through all attributes/keys in dataclasses and dictionaries to ensure floats types are converted to int. In addition to performance issues, the required code is complicated and error prone.

Note: We have more than 10 issues about dict, dataclass and Pydantic.

This feature can solve them all.

## 3 Proposed Implementation
### Before
```python
@task
def t1() -> dict:
  ...
  return {"a": 1} # Protobuf Struct {"a": 1.0}

@task
def t2(a: dict):
  print(a["integer"]) # wrong, will be a float
```
### After
```python
@task
def t1() -> dict: # Literal(scalar=Scalar(binary=Binary(value=b'msgpack_bytes', tag="msgpack")))
  ...
  return {"a": 1}  # Protobuf Binary value=b'\x81\xa1a\x01', produced by msgpack

@task
def t2(a: dict):
  print(a["integer"]) # correct, it will be a integer
```

#### Note
- We will use implement `to_python_value` to every type transformer to ensure backward compatibility.
For example, `Binary IDL Object` -> python value and `Protobuf Struct IDL Object` -> python value are both supported.

### How to turn a value to bytes?
#### Use MsgPack to convert a value into bytes
##### Python
```python
import msgpack

# Encode
def to_literal():
  msgpack_bytes = msgpack.dumps(python_val)
  return Literal(scalar=Scalar(binary=Binary(value=b'msgpack_bytes', tag="msgpack")))

# Decode
def to_python_value():
    # lv: literal value
    if lv.scalar.binary.tag == "msgpack":
        msgpack_bytes = lv.scalar.json.value
    else:
        raise ValueError(f"{tag} is not supported to decode this Binary Literal: {lv.scalar.binary}.")
    return msgpack.loads(msgpack_bytes)
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
#### Literal Value
```proto
// A simple byte array with a tag to help different parts of the system communicate about what is in the byte array.
// It's strongly advisable that consumers of this type define a unique tag and validate the tag before parsing the data.
message Binary {
    bytes value = 1; // Serialized data (MessagePack) for supported types like Dataclass, Pydantic BaseModel, and dict.
    string tag = 2; // The serialization format identifier (e.g., MessagePack). Consumers must define unique tags and validate them before deserialization.
}
```
#### Literal Type
```proto
import "google/protobuf/struct.proto";

enum SimpleType {
    NONE = 0;
    INTEGER = 1;
    FLOAT = 2;
    STRING = 3;
    BOOLEAN = 4;
    DATETIME = 5;
    DURATION = 6;
    BINARY = 7;
    ERROR = 8;
    STRUCT = 9; // Use this one.
}
message LiteralType {
    SimpleType simple = 1; // Use this one.
    google.protobuf.Struct metadata = 6; // Store Json Schema to differentiate different dataclass.
}
```

### FlytePropeller
1. Attribute Access for dictionary, Dataclass, and Pydantic in workflow.
Dict[type, type] is supported already, we have to support Dataclass, Pydantic and dict now.
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
  case *core.Scalar_JSON:
		literalType = &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_JSON}}
  ...
  return literalType 
}
```
3. Support input and default input.
```go
// Literal Input
func ExtractFromLiteral(literal *core.Literal) (interface{}, error) {
    switch literalValue := literal.Value.(type) {
        case *core.Literal_Scalar:
        ...
            case *core.Scalar_JSON:
                return scalarValue.Json.GetValue(), nil
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
                                    Value: &core.Scalar_JSON{
                                        JSON: &core.JSON{
                                            Value: []byte(""),
                                        },
										SerializationFormat: "msgpack",
                                    },
                                },
                            },
                        }, nil
    }
}
```
4. Compiler (Backward Compatibility with `Struct` type)
```go
if upstreamTypeCopy.GetSimple() == flyte.SimpleType_STRUCT && downstreamTypeCopy.GetSimple() == flyte.SimpleType_JSON {
		return true
	}
```
### FlyteKit
#### pyflyte run
The behavior will remain unchanged. 
We will pass the value to our class, which inherits from `click.ParamType`, and use the corresponding type transformer to convert the input to the correct type.

### Dict Transformer

| **Stage** | **Conversion** | **Description**                                                                                                                                                                             |
| --- | --- |---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **Before** | Python Value to Literal | 1. `Dict[type, type]` uses type hints to construct a LiteralMap. <br> 2. `dict` uses `JSON.dumps` to turn a `dict` value to a JSON string, and store it to Protobuf Struct.                 |
| | Literal to Python Value | 1. `Dict[type, type]` uses type hints to convert LiteralMap to Python Value. <br> 2. `dict` uses `JSON.loads` to turn a JSON string into a dict value and store it to Protobuf Struct.     |
| **After** | Python Value to Literal | 1. `Dict[type, type]` stays the same. <br> 2. `dict` uses `msgpack.dumps` to turn a dict into msgpack bytes, and store it to Protobuf JSON.                                            |
| | Literal to Python Value | 1. `Dict[type, type]` uses type hints to convert LiteralMap to Python Value. <br> 2.  `dict` conversion: msgpack bytes -> dict value, method: `msgpack.loads`. |

### Dataclass Transformer

| **Stage** | **Conversion** | **Description**                                                                                                                                                                                                                                                                                                                                                                                                                                                                   |
| --- | --- |-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **Before** | Python Value to Literal | Uses `mashumaro JSON Encoder` to turn a dataclass value to a JSON string, and store it to Protobuf `Struct`. |
| | Literal to Python Value | Uses `mashumaro JSON Decoder` to turn a JSON string to a python value, and recursively fixed int attributes to int (it will be float because we stored it in to `Struct`). |
| **After** | Python Value to Literal | Uses `mashumaro MessagePackEncoder` to convert a dataclass value into msgpack bytes, storing them in the Protobuf `JSON` field. |
| | Literal to Python Value | Uses `mashumaro MessagePackDecoder` to convert msgpack bytes back into a Python value. |

### Pydantic Transformer

| **Stage** | **Conversion** | **Description** |
| --- | --- | --- |
| **Before** | Python Value to Literal | Convert `BaseModel` to a JSON string, and then convert it to a Protobuf `Struct`. |
| | Literal to Python Value | Convert Protobuf `Struct` to a JSON string and then convert it to a `BaseModel`. |
| **After** | Python Value to Literal | Converts the Pydantic `BaseModel` to a dictionary, then serializes it into msgpack bytes using `msgpack.dumps`. |
| | Literal to Python Value | Deserializes `msgpack` bytes into a dictionary, then converts it back into a Pydantic `BaseModel`. |

Note: Pydantic BaseModel can't be serialized directly by `msgpack`, but this implementation will still ensure 100% correct.

### FlyteCtl
In FlyteCtl, we can construct input for the execution, so we have to make sure the values we passed to FlyteAdmin 
can all be constructed to Literal.

reference: https://github.com/flyteorg/flytectl/blob/131d6a20c7db601ca9156b8d43d243bc88669829/cmd/create/serialization_utils.go#L48 

### FlyteConsole
#### Show input/output on FlyteConsole
We will get the node's input and output literal values by FlyteAdmin’s API, and obtain the JSON IDL bytes from the literal value.

We can use MsgPack dumps the MsgPack into a dictionary, and shows it to the flyteconsole.
#### Construct Input
We should use `msgpack.encode` to encode input value and store it to the literal’s JSON field.

## 4 Metrics & Dashboards

None

## 5 Drawbacks  

None

## 6 Alternatives

None, it's doable.


## 7 Potential Impact and Dependencies
We should check whether `serialization_format` is specified and supported in the Flyte backend, Flytekit, and Flyteconsole. Currently, we use `msgpack` as our default serialization format.

In the future, we might want to support different JSON types such as "eJSON" or "ndJSON." We can add `json_type` to the JSON IDL to accommodate this.

There are 3 reasons why we add `serialization_format` to the JSON IDL rather than the literal's `metadata`:
1. Metadata use cases are more related to when the data is created, where the data is stored, etc.
2. This is required information for all JSON IDLs, and it will seem more important if we include it as a field in the IDL.
3. If we want to add `json_type` or other JSON IDL-specific use cases in the future, we can include them in the JSON IDL field, making it more readable.

## 8 Unresolved questions
None.

## 9 Conclusion
MsgPack is better because it's more smaller and faster.
You can see the performance comparison here: https://github.com/flyteorg/flyte/pull/5607#issuecomment-2333174325
We will use `msgpack` to do it.