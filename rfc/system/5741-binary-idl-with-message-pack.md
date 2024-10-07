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
Literal Type will be `SimpleType.STRUCT`.
`Json Schema` will be stored in `Literal Type's metadata`.

1. Dataclass, Pydantic BaseModel and pure dict in python will all use `SimpleType.STRUCT`.
2. We will put `Json Schema` in Literal Type's `metadata` field, this will be used in flytekit remote api to construct dataclass/Pydantic BaseModel by `Json Schema`.
3. We will use libraries written in golang to compare `Json Schema` to solve this issue: ["[BUG] Union types fail for e.g. two different dataclasses"](https://github.com/flyteorg/flyte/issues/5489).

Note: The `metadata` of `Literal Type` and `Literal Value` are not the same.

## 2 Motivation

Prior to this RFC, in flytekit, when handling dataclasses, Pydantic base models, and dictionaries, we store data using a JSON string within Protobuf struct datatype.

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
        msgpack_bytes = lv.scalar.binary.value
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
   - The library [github.com/vmihailenco/msgpack](https://github.com/vmihailenco/msgpack) doesn't support strict type deserialization (for example, `map[int]string`), but [github.com/shamaton/msgpack/v2](https://github.com/shamaton/msgpack) supports this feature. This is super important for backward compatibility.

3. **Library Popularity**: 
   - While [github.com/shamaton/msgpack/v2](https://github.com/shamaton/msgpack) has fewer stars on GitHub, it has proven to be reliable in various test cases. All cases created by me have passed successfully, which you can find in this [pull request](https://github.com/flyteorg/flytekit/pull/2751).

4. **Project Activity**: 
   - [github.com/shamaton/msgpack/v2](https://github.com/shamaton/msgpack) is still an actively maintained project. The author responds quickly to issues and questions, making it a well-supported choice for projects requiring ongoing maintenance and active support.

5. **Testing Process**: 
   - I initially started with [github.com/vmihailenco/msgpack](https://github.com/vmihailenco/msgpack) but switched to [github.com/shamaton/msgpack/v2](https://github.com/shamaton/msgpack) due to its better support for strict typing and the active support provided by the author.


##### JavaScript
```javascript
import { encode, decode } from '@msgpack/msgpack';

// Example data to encode
const data = { a: 1 };

// Encode the data
const encodedData = encode(data);

// Print the encoded data
console.log(encodedData); // <Buffer 81 a1 61 01>

// Decode the data
const decodedData = decode(encodedData);

// Print the decoded data
console.log(decodedData); // { a: 1 }
```
reference: https://github.com/msgpack/msgpack-javascript 

### FlyteIDL
#### Literal Value

Here is the [IDL definition](https://github.com/flyteorg/flyte/blob/7989209e15600b56fcf0f4c4a7c9af7bfeab6f3e/flyteidl/protos/flyteidl/core/literals.proto#L42-L47).

The `bytes` field is used for serialized data, and the `tag` field specifies the serialization format identifier.
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
// "tag" will default be "msgpack"
func MakeBinaryLiteral(v []byte, tag string) *core.Literal {
	return &core.Literal{
		Value: &core.Literal_Scalar{
			Scalar: &core.Scalar{
				Value: &core.Scalar_Binary{
					Binary: &core.Binary{
						Value: v,
						Tag:   tag,
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

In all transformers, we should implement a function called `from_binary_idl` to convert the Binary IDL Object into the desired type.

A base method can be added to the `TypeTransformer` class, allowing child classes to override it as needed.

During attribute access, Flyte Propeller will deserialize the msgpack bytes into a map object in golang, retrieve the specific attribute, and then serialize it back into msgpack bytes (resulting in a Binary IDL Object containing msgpack bytes).

This implies that when converting a literal to a Python value, we will receive `msgpack bytes` instead of the `expected Python type`.

```python
# In Mashumaro, the default encoder uses strict_map_key=False, while the default decoder uses strict_map_key=True.
# This is relevant for cases like Dict[int, str].
# If strict_map_key=False is not used, the decoder will raise an error when trying to decode keys that are not strictly typed.
def _default_flytekit_decoder(data: bytes) -> Any:
    return msgpack.unpackb(data, raw=False, strict_map_key=False)


def from_binary_idl(self, binary_idl_object: Binary, expected_python_type: Type[T]) -> Optional[T]:
    # Handle msgpack serialization
    if binary_idl_object.tag == "msgpack":
        try:
            # Retrieve the existing decoder for the expected type
            decoder = self._msgpack_decoder[expected_python_type]
        except KeyError:
            # Create a new decoder if not already cached
            decoder = MessagePackDecoder(expected_python_type, pre_decoder_func=_default_flytekit_decoder)
            self._msgpack_decoder[expected_python_type] = decoder
        # Decode the binary IDL object into the expected Python type
        return decoder.decode(binary_idl_object.value)
    else:
        # Raise an error if the binary format is not supported
        raise TypeTransformerFailedError(f"Unsupported binary format {binary_idl_object.tag}")
```

Note: 
1. This base method can handle primitive types, nested typed dictionaries, nested typed lists, and combinations of nested typed dictionaries and lists.

2. Dataclass transformer needs its own `from_binary_idl` method to handle specific cases such as [discriminated classes](https://github.com/flyteorg/flyte/issues/5588).

3. Flyte types (e.g., FlyteFile, FlyteDirectory, StructuredDataset, and FlyteSchema) will need their own `from_binary_idl` methods, as they must handle downloading files from remote object storage when converting literals to Python values.

For example, see the FlyteFile implementation: https://github.com/flyteorg/flytekit/pull/2751/files#diff-22cf9c7153b54371b4a77331ddf276a082cf4b3c5e7bd1595dd67232288594fdR522-R552

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

#### Literal Type's TypeStructure's dataclass_type
This is used for compiling dataclass attribute access. 

With it, we can retrieve the literal type of an attribute and validate it in Flyte's propeller compiler.

For more details, check here: https://github.com/flyteorg/flytekit/blob/fb55841f8660b2a31e99381dd06e42f8cd22758e/flytekit/core/type_engine.py#L454-L525

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

#### Literal Type's TypeStructure's dataclass_type
This is used for compiling Pydantic BaseModel attribute access. 

With it, we can retrieve an attribute's literal type and validate it in Flyte's propeller compiler.

Although this feature is not currently implemented, it will function similarly to the dataclass transformer in the future.

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

When we receive `SimpleType.STRUCT`, we can construct a `Binary IDL Object` using the following logic in `flyteidl/clients/go/coreutils/literals.go`:

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

This is how users can create an execution by using a YAML file:
```bash
flytectl create execution --execFile ./flytectl/create_dataclass_task.yaml -p flytesnacks -d development
```

Example YAML file (`create_dataclass_task.yaml`):
```yaml
iamRoleARN: ""
inputs:
  input:
    a: 1
    b: 3.14
    c: example_string
    d:
      "1": 100
      "2": 200
    e:
      a: 1
      b: 3.14
envs: {}
kubeServiceAcct: ""
targetDomain: ""
targetProject: ""
task: dataclass_example.dataclass_task
version: OSyTikiBTAkjBgrL5JVOVw
```

### FlyteCopilot

When we need to pass an attribute access value to a copilot task, we must modify the code to convert a Binary Literal value with the `msgpack` tag into a primitive value.

(Currently, we will only support primitive values.)

You can reference the relevant section of code here:

[FlyteCopilot - Data Download](https://github.com/flyteorg/flyte/blob/7989209e15600b56fcf0f4c4a7c9af7bfeab6f3e/flytecopilot/data/download.go#L88-L95)

### FlyteConsole
#### How users input into launch form?
When FlyteConsole receives a literal type of `SimpleType.STRUCT`, the input method depends on the availability of a JSON schema:

1. No JSON Schema provided:

Input is expected as `a Javascript Object` (e.g., `{"a": 1}`).

2. JSON Schema provided:

Users can input values based on the schema's expected type and construct an appropriate `Javascript Object`.

Note:

For `dataclass` and Pydantic `BaseModel`, a JSON schema will be provided in their literal type, and the input form will be constructed accordingly.

##### What happens after the user enters data?

Input values -> Javascript Object -> msgpack bytes -> Binary IDL With MessagePack Bytes

#### Displaying Inputs/Outputs in the Console
Use `msgpack` to deserialize bytes into an Object and display it in Flyte Console.

#### Copying Inputs/Outputs in the Console
Allow users to copy the `Object` to the clipboard, as currently implemented.

#### Pasting and Copying from FlyteConsole
Currently, we should support JSON pasting if the content is a JavaScript object. However, there's a question of how we might handle other formats like YAML or MsgPack bytes, especially if copied from a binary file.

For now, focusing on supporting JSON pasting makes sense. However, adding support for YAML and MsgPack bytes could be valuable future enhancements.

## 4 Metrics & Dashboards

None

## 5 Drawbacks  

None

## 6 Alternatives

MsgPack is a good choice because it's more smaller and faster than UTF-8 Encoded JSON String.

You can see the performance comparison here: https://github.com/flyteorg/flyte/pull/5607#issuecomment-2333174325

We will use `msgpack` to do it.

## 7 Potential Impact and Dependencies
None.

## 8. Unresolved Questions
### Conditional Branch
Currently, our support for `DataClass/BaseModel/Dict[type, type]` within conditional branches is incomplete. At present, we only support comparisons of primitive types. However, there are two key challenges when attempting to handle these more complex types:

1. **Primitive Type Comparison vs. Binary IDL Object:**
   - In conditional branches, we receive a `Binary IDL Object` during type comparison, which needs to be converted into a `Primitive IDL Object`. 
   - The issue is that we don't know the expected Python type or primitive type beforehand, making this conversion ambiguous.

2. **MsgPack Incompatibility with `Primitive_Datetime` and `Primitive_Duration`:**
   - MsgPack does not natively support the `Primitive_Datetime` and `Primitive_Duration` types, and instead converts them to strings.
   - This can lead to inconsistencies in comparison logic. One potential workaround is to convert both types to strings for comparison. However, it is uncertain whether this approach is the best solution.

## 9 Conclusion

1. Binary IDL with MessagePack Bytes provides a better representation for dataclasses, Pydantic BaseModels, and untyped dictionaries in Flyte.

2. This approach ensures 100% accuracy of each attribute and enables attribute access.
