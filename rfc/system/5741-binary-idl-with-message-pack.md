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


Note:

1. If a Python value can't be directly converted to `MessagePack Bytes`, we can first convert it to a `Dict`, and then convert it to `MessagePack Bytes`.

   - **For example:** The Pydantic-to-literal function workflow will be:
     `BaseModel` -> `dict` -> `MessagePack Bytes` -> `Binary (value: MessagePack Bytes, tag: msgpack) IDL Object`.

   - **For pure `dict` in Python:** The to-literal function workflow will be:
     `dict` -> `MessagePack Bytes` -> `Binary (value: MessagePack Bytes, tag: msgpack) IDL Object`.

2. There is **NO JSON** involved in the new type at all. Only **JSON Schema** is used to construct `DataClass` or `Pydantic BaseModel`.


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

This results in performance issues since we need to recursively iterate through all attributes/keys in dataclasses and dictionaries to ensure floats types are converted to int.

In addition to performance issues, the required code is complicated and error prone.

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
    "github.com/shamaton/msgpack/v2"
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

reference: [shamaton/msgpack GitHub Repository](https://github.com/shamaton/msgpack)

Notes:

1. **MessagePack Implementations**: 
   - We can explore all MessagePack implementations for Golang at the [MessagePack official website](https://msgpack.org/index.html).

2. **Library Comparison**: 
   - The library [github.com/vmihailenco/msgpack](https://github.com/vmihailenco/msgpack) doesn't support strict type deserialization (for example, `map[int]string`), but `"github.com/shamaton/msgpack/v2"` supports this feature. This is super important for backward copmatibility.

3. **Library Popularity**: 
   - While `"github.com/shamaton/msgpack/v2"` has fewer stars on GitHub, it has proven to be reliable in various test cases. All cases created by me have passed successfully, which you can find in this [pull request](https://github.com/flyteorg/flytekit/pull/2751).

4. **Project Activity**: 
   - `"github.com/shamaton/msgpack/v2"` is still an actively maintained project. The author responds quickly to issues and questions, making it a well-supported choice for projects requiring ongoing maintenance and active support.

5. **Testing Process**: 
   - I initially started with [github.com/vmihailenco/msgpack](https://github.com/vmihailenco/msgpack) but switched to `"github.com/shamaton/msgpack/v2"` due to its better support for strict typing and the active support provided by the author.


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
  case *core.Scalar_Binary:
		literalType = &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_BINARY}}
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
            case *core.Scalar_Binary:
			    return scalarValue.Binary, nil
    }
}
// Default Input
func MakeDefaultLiteralForType(typ *core.LiteralType) (*core.Literal, error) {
	switch t := typ.GetType().(type) {
	case *core.LiteralType_Simple:
        case core.SimpleType_BINARY:
			return MakeLiteral([]byte{})
    }
}
// Use Message Pack as Default Tag for deserialization.
func MakeBinaryLiteral(v []byte) *core.Literal {
	return &core.Literal{
		Value: &core.Literal_Scalar{
			Scalar: &core.Scalar{
				Value: &core.Scalar_Binary{
					Binary: &core.Binary{
						Value: v,
						Tag:   "msgpack",
					},
				},
			},
		},
	}
}
```
4. Compiler
```go
func (t trivialChecker) CastsFrom(upstreamType *flyte.LiteralType) bool {
	if upstreamType.GetEnumType() != nil {
		if t.literalType.GetSimple() == flyte.SimpleType_STRING {
			return true
		}
	}

    if t.literalType.GetEnumType() != nil {
        if upstreamType.GetSimple() == flyte.SimpleType_STRING {
            return true
        }
    }

	if GetTagForType(upstreamType) != "" && GetTagForType(t.literalType) != GetTagForType(upstreamType) {
		return false
	}

    // Here is the new way to check if dataclass/pydantic BaseModel are castable or not.
    if upstreamTypeCopy.GetSimple() == flyte.SimpleType_STRUCT &&\
         downstreamTypeCopy.GetSimple() == flyte.SimpleType_STRUCT {
        // Json Schema is stored in Metadata
        upstreamMetadata := upstreamTypeCopy.GetMetadata()
        downstreamMetadata := downstreamTypeCopy.GetMetadata()

        // There's bug in flytekit's dataclass Transformer to generate JSON Scheam before,
        // in some case, we the JSON Schema will be nil, so we can only pass it to support
        // backward compatible. (reference task should be supported.)
        if upstreamMetadata == nil || downstreamMetadata == nil {
            return true
        }

        return isSameTypeInJSON(upstreamMetadata, downstreamMetadata) ||\
                 isSuperTypeInJSON(upstreamMetadata, downstreamMetadata)
    }

	upstreamTypeCopy := *upstreamType
	downstreamTypeCopy := *t.literalType
	upstreamTypeCopy.Structure = &flyte.TypeStructure{}
	downstreamTypeCopy.Structure = &flyte.TypeStructure{}
	upstreamTypeCopy.Metadata = &structpb.Struct{}
	downstreamTypeCopy.Metadata = &structpb.Struct{}
	upstreamTypeCopy.Annotation = &flyte.TypeAnnotation{}
	downstreamTypeCopy.Annotation = &flyte.TypeAnnotation{}
	return upstreamTypeCopy.String() == downstreamTypeCopy.String()
}
```
### FlyteKit
#### Attribute Access
In most transformers, we should create a function `from_binary_idl` to convert the Binary IDL Object into the desired type.

When performing attribute access, Propeller will deserialize the msgpack bytes into a map object, retrieve the attribute, and then serialize it back into msgpack bytes (a Binary IDL Object containing msgpack bytes).

This means that when converting a literal to a Python value, we will receive `msgpack bytes` instead of our `expected Python type`.

#### pyflyte run
The behavior will remain unchanged.

We will pass the value to our class, which inherits from `click.ParamType`, and use the corresponding type transformer to convert the input to the correct type.

### Dict Transformer
There are 2 cases in Dict Transformer, `Dict[type, type]` and `dict`.

For `Dict[type, type]`, we will stay everything the same as before.

#### Literal Value
For `dict`, the life cycle of it will be as below.

Before:
- `to_literal`: `dict` -> `JSON String` -> `Protobuf Struct`
- `to_python_val`: `Protobuf Struct` -> `JSON String` -> `dict`

After:
- `to_literal`: `dict` -> `msgpack bytes` -> `Binary(value=b'msgpack_bytes', tag="msgpack")`
- `to_python_val`: `Binary(value=b'msgpack_bytes', tag="msgpack")` -> `msgpack bytes` -> `dict`

#### JSON Schema
The JSON Schema of `dict` will be empty.
### Dataclass Transformer
#### Literal Value
Before:
- `to_literal`: `dataclass` -> `JSON String` -> `Protobuf Struct`
- `to_python_val`: `Protobuf Struct` -> `JSON String` -> `dataclass`

After:
- `to_literal`: `dataclass` -> `msgpack bytes` -> `Binary(value=b'msgpack_bytes', tag="msgpack")`
- `to_python_val`: `Binary(value=b'msgpack_bytes', tag="msgpack")` -> `msgpack bytes` -> `dataclass`

Note: We will use mashumaro's `MessagePackEncoder` and `MessagePackDecoder` to serialize and deserialize dataclass value in python.
```python
from mashumaro.codecs.msgpack import MessagePackDecoder, MessagePackEncoder
```

#### JSON Schema
The JSON Schema of `dataclass` will be generated by `marshmallow` or `mashumaro`.
Check here: https://github.com/flyteorg/flytekit/blob/8c6f6f0f17d113447e1b10b03e25a34bad79685c/flytekit/core/type_engine.py#L442-L474

### Pydantic Transformer
#### Literal Value
Pydantic can't be serialized to `msgpack bytes` directly.
But `dict` can.

- `to_literal`: `BaseModel` -> `dict` -> `msgpack bytes` -> `Binary(value=b'msgpack_bytes', tag="msgpack")`
- `to_python_val`: `Binary(value=b'msgpack_bytes', tag="msgpack")` -> `msgpack bytes` -> `dict` -> `BaseModel`

Note: Pydantic BaseModel can't be serialized directly by `msgpack`, but this implementation will still ensure 100% correct.

```python
@dataclass
class DC_inside:
    a: int
    b: float

@dataclass
class DC:
    a: int
    b: float
    c: str
    d: Dict[str, int]
    e: DC_inside

class MyDCModel(BaseModel):
    dc: DC

my_dc = MyDCModel(dc=DC(a=1, b=2.0, c="3", d={"4": 5}, e=DC_inside(a=6, b=7.0)))
# {'dc': {'a': 1, 'b': 2.0, 'c': '3', 'd': {'4': 5}, 'e': {'a': 6, 'b': 7.0}}}
```

#### JSON Schema
The JSON Schema of `BaseModel` will be generated by Pydantic's API.
```python
@dataclass
class DC_inside:
    a: int
    b: float

@dataclass
class DC:
    a: int
    b: float
    c: str
    d: Dict[str, int]
    e: DC_inside

class MyDCModel(BaseModel):
    dc: DC

my_dc = MyDCModel(dc=DC(a=1, b=2.0, c="3", d={"4": 5}, e=DC_inside(a=6, b=7.0)))
my_dc.model_json_schema()
"""
{'$defs': {'DC': {'properties': {'a': {'title': 'A', 'type': 'integer'}, 'b': {'title': 'B', 'type': 'number'}, 'c': {'title': 'C', 'type': 'string'}, 'd': {'additionalProperties': {'type': 'integer'}, 'title': 'D', 'type': 'object'}, 'e': {'$ref': '#/$defs/DC_inside'}}, 'required': ['a', 'b', 'c', 'd', 'e'], 'title': 'DC', 'type': 'object'}, 'DC_inside': {'properties': {'a': {'title': 'A', 'type': 'integer'}, 'b': {'title': 'B', 'type': 'number'}}, 'required': ['a', 'b'], 'title': 'DC_inside', 'type': 'object'}}, 'properties': {'dc': {'$ref': '#/$defs/DC'}}, 'required': ['dc'], 'title': 'MyDCModel', 'type': 'object'}
"""
```

### FlyteCtl
In FlyteCtl, we can construct input for the execution.

We can construct a `Binary IDL Object` when we receive `Literal Type Struct`.

In `flyteidl/clients/go/coreutils/literals.go`:
```go
if newT.Simple == core.SimpleType_STRUCT {
    if _, isValueStringType := v.(string); !isValueStringType {
        byteValue, err := msgpack.Marshal(v)
        if err != nil {
            return nil, fmt.Errorf("unable to marshal to json string for struct value %v", v)
        }
        strValue = string(byteValue)
    }
}
``` 

### FlyteConsole
#### Show input/output on FlyteConsole

1. Get Bytes from the Binary IDL Object.
2. Get Tag from the Binary IDL Object.
3. Use Tag to determine how to deserialize `Bytes`, and show the deserialized value in FlyteConsole.

#### Copy Input
Return `MessagePack Bytes` to your clipboard, it can be used in `Input Bytes` below.

#### Construct Input
##### Input Bytes
1. Input `MessagePack Bytes` to the console, for example, `{"a": 1}` in python will be `b'\x81\xa1a\x01'`.
2. Tag can only use `msgpack`, it's forced now.

##### Launch Form
1. Use `JSON Schema` to know what every input field, and make users input it 1 by 1.
2. Serialize the input value into `MessagePack Bytes`
3. Create a `Binary IDL Object`, value is `MessagePack Bytes` and tag is `msgpack`.

## 4 Metrics & Dashboards

None

## 5 Drawbacks  

None

## 6 Alternatives

None, it's doable.


## 7 Potential Impact and Dependencies
None.

## 8 Unresolved questions
None.

## 9 Conclusion
MsgPack is a good choice because it's more smaller and faster than UTF-8 Encoded JSON String.

You can see the performance comparison here: https://github.com/flyteorg/flyte/pull/5607#issuecomment-2333174325
We will use `msgpack` to do it.
