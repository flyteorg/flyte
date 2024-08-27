# Tuple IDL

**Authors:**

- [@Chun-Mao (Michael) Lai](https://github.com/Mecoli1219)

## 1 Executive Summary

Goals:

- Implement support for tuple types (list of different types) in FlyteIDL
  - Example: `a : Tuple[int, str, float]` must be assigned with integer for first element, string for second element, and floating value for third element.
- Implement support for new Python types in Flytekit
  - `typing.NamedTuple`
    - `typing.NamedTuple("NAME", ("K1": T1), ("K2": T2))`
  - `typing.Tuple`:
    - `typing.Tuple[T1, T2, T3]` (Simplify case for NamedTuple)
    - `typing.Tuple[T1, ...]` (Not the goal of this tuple IDL, since it is quite similar to `typing.List[T1]`)

Implementation:

- Each element in the tuple should be considered independently, while the order of elements should be preserved. In order to support NamedTuple, we need to store the name of each element.
  ```proto
  message LiteralField {
      string name = 1;
      Literal literal = 2;
  }
  ```

## 2 Motivation

Flytekit restricts the support of tuple types currently, which are commonly used in Python. This RFC aims to add support for tuple types in FlyteIDL and Flytekit.

Note: We have several issues about NamedTuple and tuple support in Flytekit.

- [#1337](https://github.com/flyteorg/flyte/issues/1337)
- [#3158](https://github.com/flyteorg/flyte/issues/3158)
- [#4358](https://github.com/flyteorg/flyte/issues/4358)

This feature can solve them all.

## 3 Proposed Implementation

### FlyteIDL

#### Literal of NamedTuple & Tuple

- Each field in the tuple should be considered independently, while the order of fields should be preserved. In order to support NamedTuple, we need to store the name of each field. Each field of the tuple should be stored in a `LiteralField` message.

  ```proto
   message LiteralField {
       string name = 1;
       Literal literal = 2;
   }
  ```

- The tuple should be stored in a `LiteralFieldCollection` message.

  ```proto
  message LiteralFieldCollection {
      string name = 1;
      repeated LiteralField fields = 1;
  }

  message Literal {
      oneof value {
          // ...
          LiteralFieldCollection tuple = 4;
      }
      // ...
  }
  ```

- Similarly, we have to add it to binding data for reference in the output.

  ```proto
  message BindingDataField {
      string name = 1;
      BindingData binding = 2;
  }

  message BindingDataFieldCollection {
      string name = 1;
      repeated BindingDataField fields = 2;
  }

  message BindingData {
      oneof data {
          // ...
          BindingDataFieldCollection tuple = 4;
      }
      // ...
  }
  ```

- These messages types should be added to [flyteidl/protos/flyteidl/core/types.proto](https://github.com/flyteorg/flyte/blob/master/flyteidl/protos/flyteidl/core/literals.proto)

#### Type of NamedTuple and Tuple

- The type of NamedTuple and Tuple should be stored in a `TupleType` message.

  ```proto
  message TupleType {
      repeated LiteralType types = 1;
  }

  message LiteralType {
      oneof type {
          // ...
          TupleType tuple_type = 11;
      }
      // ...
  }
  ```

- The above codes should be added to [flyteidl/protos/flyteidl/core/types.proto](https://github.com/flyteorg/flyte/blob/master/flyteidl/protos/flyteidl/core/types.proto). However, the proposed `TupleType` is exactly the same as `UnionType` in the current FlyteIDL except for the name and the variable name. We may need to consider whether to merge them or not.

### FlytePropeller

1. Implement a new type checker in [flytepropeller/pkg/compiler/validators/typing.go](https://github.com/flyteorg/flyte/blob/master/flytepropeller/pkg/compiler/validators/typing.go)

   ```go
    type tupleTypeChecker struct {
        literalType *flyte.LiteralType
    }

    func (t tupleTypeChecker) CastsFrom(upstreamType *flyte.LiteralType) bool {
        tupleType := t.literalType.GetTupleType()
        upstreamTupleType := upstreamType.GetTupleType()

        if upstreamTupleType != nil && tupleType != nil {
            // For each field in the tuple, check if the upstream type can be casted to the downstream type
            tupleTypes := tupleType.GetTypes()
                upstreamTupleTypes := upstreamTupleType.GetTypes()
                if len(tupleTypes) == len(upstreamTupleTypes) {
                    for i, downstreamType := range tupleTypes {
                        if !getTypeChecker(downstreamType).CastsFrom(upstreamTupleTypes[i]) {
                            return false
                        }
                    }
                    return true
                }

        }
        return false
    }


    func getTypeChecker(t *flyte.LiteralType) typeChecker {
        switch t.GetType().(type) {
        // ...
        case *flyte.LiteralType_TupleType:
            return tupleTypeChecker{
                literalType: t,
            }
        }
    }
   ```

2. Update the bindings type validation code in [flytepropeller/pkg/compiler/validators/bindings.go](https://github.com/flyteorg/flyte/blob/master/flytepropeller/pkg/compiler/validators/bindings.go)

   ```go
    func validateBinding(w c.WorkflowBuilder, nodeID c.NodeID, nodeParam string, binding *flyte.BindingData,
        expectedType *flyte.LiteralType, errs errors.CompileErrors, validateParamTypes bool) (
        resolvedType *flyte.LiteralType, upstreamNodes []c.NodeID, ok bool) {

        switch binding.GetValue().(type) {
        // ...
       case *flyte.BindingData_Tuple:
            // Goes through union-aware AreTypesCastable
   	    break
        default:
            // ...
        }

        switch val := binding.GetValue().(type) {
        // ...
        case *flyte.BindingData_Tuple:
            if val.Tuple == nil {
                errs.Collect(errors.NewValueRequiredErr(nodeID, nodeParam))
                return nil, nil, !errs.HasErrors()
            }

            if expectedType.GetTupleType() != nil {
                allNodeIds := make([]c.NodeID, 0, len(val.Tuple.GetFields()))
                var subTypes []*flyte.LiteralType
                for i, f := range val.Tuple.GetFields() {
                    if resolvedType, nodeIds, ok := validateBinding(w, nodeID, nodeParam, f.GetBinding(), expectedType.GetTupleType().GetTypes()[i], errs.NewScope(), validateParamTypes); ok {
                        allNodeIds = append(allNodeIds, nodeIds...)
                        subTypes = append(subTypes, resolvedType)
                    }
                }

                return &flyte.LiteralType{
                    Type: &flyte.LiteralType_TupleType{
                        TupleType: &flyte.TupleType{
                            Types: subTypes,
                    },
                }, allNodeIds, !errs.HasErrors()
            }

            errs.Collect(errors.NewMismatchingBindingsErr(nodeID, nodeParam, expectedType.String(), val.Collection.String()))
        }

    }
   ```

### FlyteKit

- Supporting the following types in Flytekit:
  - `NamedTuple("a", [("x", int), ("y", str)])` could be stored as
    ```proto
    LiteralFieldCollection {
        name: "a"
        fields: [
            LiteralField {
                name: "x"
                literal: Literal { scalar: Scalar { primitive: Primitive { integer: 1 } } }
            }
            LiteralField {
                name: "y"
                literal: Literal { scalar: Scalar { primitive: Primitive { string_value: "hello" } } }
            }
        ]
    }
    ```
  - `Tuple[int, str]` could be stored as
    ```proto
    LiteralFieldCollection {
        name: ""
        fields: [
            LiteralField {
                name: ""
                literal: Literal { scalar: Scalar { primitive: Primitive { integer: 1 } } }
            }
            LiteralField {
                name: ""
                literal: Literal { scalar: Scalar { primitive: Primitive { string_value: "hello" } } }
            }
        ]
    }
    ```

---

#### pyflyte run

The behavior will remain unchanged.
We will pass the value to our class, which inherits from `click.ParamType`, and use the corresponding type transformer to convert the input to the correct type.

### Dict Transformer

| **Stage**  | **Conversion**          | **Description**                                                                                                                                                                                                                                                                                                               |
| ---------- | ----------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Before** | Python Value to Literal | 1. `Dict[type, type]` uses type hints to construct a LiteralMap. <br> 2. `dict` uses `json.dumps` to turn a `dict` value to a JSON string, and store it to protobuf Struct.                                                                                                                                                   |
|            | Literal to Python Value | 1. `Dict[type, type]` uses type hints to convert LiteralMap to Python Value. <br> 2. `dict` uses `json.loads` to turn a JSON string to a dict value and store it to protobuf Struct.                                                                                                                                          |
| **After**  | Python Value to Literal | 1. `Dict[type, type]` stays the same. <br> 2. ~~`dict` uses `msgpack.dumps` to turn a dict value to a JSON byte string, and store is to protobuf Json.~~ <br> 3. `dict` uses `msgpack.dumps` to turn a JSON string to a byte string, and store is to protobuf Json.                                                           |
|            | Literal to Python Value | 1. `Dict[type, type]` uses type hints to convert LiteralMap to Python Value. <br> 2. ~~`dict` uses `msgpack.loads` to turn a byte string to a JSON string and `dict` value and store it to protobuf `Struct`.~~ <br> 3. `dict` conversion: byte string -> JSON string -> dict value, method: `msgpack.loads` -> `json.loads`. |

### Dataclass Transformer

| **Stage**  | **Conversion**          | **Description**                                                                                                                                                                                                                                                                                                                                                                                                                                                                    |
| ---------- | ----------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Before** | Python Value to Literal | Uses `mashumaro JSON Encoder` to turn a dataclass value to a JSON string, and store it to protobuf `Struct`. <br><br> ~~Note: For `FlyteTypes`, we will inherit mashumaro `SerializableType` to define our own serialization behavior, which includes uploading files to remote storage.~~                                                                                                                                                                                         |
|            | Literal to Python Value | Uses `mashumaro JSON Decoder` to turn a JSON string to a python value, and recursively fixed int attributes to int (it will be float because we stored it in to `Struct`). <br><br> ~~Note: For `FlyteTypes`, we will inherit the mashumaro `SerializableType` to define our own deserialization behavior, which includes download file from remote storage.~~                                                                                                                     |
| **After**  | Python Value to Literal | ~~Uses `msgpack.dumps()` to turn a Python value into a byte string.~~ <br> Uses `mashumaro JSON Encoder` to turn a dataclass value to a JSON string, and uses `msgpack.dumps()` to turn the JSON string into a byte string, and store it to protobuf `Json`. <br><br> ~~Note: For `FlyteTypes`, we will need to customize serialization behavior by msgpack reference [here](https://github.com/msgpack/msgpack-python?tab=readme-ov-file#packingunpacking-of-custom-data-type).~~ |
|            | Literal to Python Value | ~~Uses `msgpack.loads()` to turn a byte string into a Python value.~~ <br> Uses `msgpack.loads()` to turn a byte string into a JSON string, and uses `mashumaro JSON Decoder` to turn the JSON string into a Python value. <br><br> ~~Note: For `FlyteTypes`, we will need to customize deserialization behavior by `msgpack` reference [here](https://github.com/msgpack/msgpack-python?tab=readme-ov-file#packingunpacking-of-custom-data-type).~~                               |

### Pydantic Transformer

| **Stage**  | **Conversion**          | **Description**                                                                                                                                                                                                                                                                         |
| ---------- | ----------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Before** | Python Value to Literal | Convert `BaseModel` to a JSON string, and then convert it to a Protobuf `Struct`.                                                                                                                                                                                                       |
|            | Literal to Python Value | Convert Protobuf `Struct` to a JSON string and then convert it to a `BaseModel`.                                                                                                                                                                                                        |
| **After**  | Python Value to Literal | ~~Convert the Pydantic `BaseModel` to a JSON string, then convert the JSON string to a `dictionary`, and finally, convert it to a `byte string` using msgpack.~~ <br> Convert the Pydantic `BaseModel` to a JSON string, then convert the JSON string to a `byte string` using msgpack. |
|            | Literal to Python Value | ~~Convert `byte string` to a `dictionary` using `msgpack`, then convert dictionary to a JSON string, and finally, convert it to Pydantic `BaseModel`.~~ <br> Convert `byte string` to a JSON string using `msgpack`, then convert it to Pydantic `BaseModel`.                           |

### FlyteCtl

In FlyteCtl, we can construct input for the execution, so we have to make sure the values we passed to FlyteAdmin
can all be constructed to Literal.

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

~~There's no breaking changes if we use `Annotated[dict, Json]`, but we need to be very careful about will there be any breaking changes.~~

None

## 6 Alternatives

~~Use UTF-8 format to encode and decode, this will be more easier for implementation, but maybe will cause performance issue when using Pydantic Transformer.~~
None, it's doable.

## 7 Potential Impact and Dependencies

_Here, we aim to be mindful of our environment and generate empathy towards others who may be impacted by our decisions._

- _What other systems or teams are affected by this proposal?_
- _How could this be exploited by malicious attackers?_

## 8 Unresolved questions

~~I am not sure use UTF-8 format or msgpack to encode and decode is a better option.~~
MsgPack is better because it's more smaller and faster.

## 9 Conclusion

~~Whether use UTF-8 format or msgpack to encode and decode, we will definitely do it.~~
We will use msgpack to do it.
