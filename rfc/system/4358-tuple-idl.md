# Tuple IDL

**Authors:**

- [@Chun-Mao (Michael) Lai](https://github.com/Mecoli1219)

## 1 Executive Summary

Goals:

- Implement support for tuple types (lists of different types) in FlyteIDL.
	- Example: a : Tuple[int, str, float] must be assigned an integer for the first element, a string for the second element, and a floating value for the third element.
- Implement support for new Python types in Flytekit:
	- `typing.NamedTuple`
	    - `typing.NamedTuple("NAME", ("K1", T1), ("K2", T2))`
	- `typing.Tuple`
	    - `typing.Tuple[T1, T2, T3]` (Simplified case for NamedTuple)
	    - `typing.Tuple[T1, ...]` (It is quite similar to `typing.List[T1]`, we can address it in this RFC or use `LiteralType_CollectionType` to support it for now.)

Implementation:

-  A new `LiteralType` will be introduced to represent tuple types in FlyteIDL. Each element in the tuple will be considered independently, preserving the order of elements. For NamedTuple, the name of each element will be stored along with its type, while for regular tuples, the names will be left empty.
    ```proto
    message TupleFieldType {
        string name = 1;
        LiteralType type = 2;
    }

    message TupleType {
        string tuple_name = 1;
        repeated TupleFieldType fields = 2;
    }

    message LiteralType {
        oneof type {
            // ...
            TupleType tuple_type = 11;
        }
        // ...
    }
    ```

## 2 Motivation

Flytekit currently restricts the support of tuple types, which are commonly used in Python. This RFC aims to add support for tuple types in FlyteIDL and Flytekit.

Note: Several issues regarding `NamedTuple` and `tuple` support in Flytekit have been identified:

- [#1337](https://github.com/flyteorg/flyte/issues/1337)
- [#3158](https://github.com/flyteorg/flyte/issues/3158)
- [#4358](https://github.com/flyteorg/flyte/issues/4358)

This feature aims to resolve all these issues.

## 3 Proposed Implementation

### FlyteIDL

#### Type of NamedTuple and Tuple

- To support NamedTuple, we need to store the name of each field and the name of the tuple itself. We will create a new message type `TupleFieldType` to store the name and type of each field. For `Tuple`, we can use the same `TupleType`, but the `name` and `tuple_name` will be empty strings.

    ```proto
    message TupleFieldType {
        string name = 1;
        LiteralType type = 2;
    }

    message TupleType {
        string tuple_name = 1;
        repeated TupleFieldType fields = 2;
    }

    message LiteralType {
        oneof type {
            // ...
            TupleType tuple_type = 11;
        }
        // ...
    }
    ```

- We could also consider supporting `typing.Tuple[T1, ...]` in the future by adding an additional field `is_univariate` in `TupleType`. However, the performance of `typing.Tuple[T1, ...]` is closely related to `typing.List[T1]`, so this use case could be covered by `LiteralType_CollectionType` for now. (#TO_DISCUSS)

    ```proto
    message TupleType {
        string tuple_name = 1;
        bool is_univariate = 2;
        repeated TupleFieldType fields = 3;
    }
    ```

- The above code should be added to [flyteidl/protos/flyteidl/core/types.proto](https://github.com/flyteorg/flyte/blob/master/flyteidl/protos/flyteidl/core/types.proto).

#### Literal of NamedTuple & Tuple

- Since the type, including the names of the fields and the tuple, is stored in `LiteralType`, we only need to store the value of each field in the tuple. We can use the existing `LiteralCollection` message from [flyteidl/protos/flyteidl/core/types.proto](https://github.com/flyteorg/flyte/blob/master/flyteidl/protos/flyteidl/core/literals.proto) to store the value of each field.

    ```proto
    message LiteralCollection {
        repeated Literal literals = 1;
    }

    message Literal {
        oneof value {
            // ...
            // A collection of literals to allow nesting.
            LiteralCollection collection = 2;
            // ...
        }
        // ...
    }
    ```

#### Other Changes

- Update `MakeDefaultLiteralForType()` and `MakeLiteralForType()` functions in [flyteidl/clients/go/coreutils/literals.go](https://github.com/flyteorg/flyte/blob/master/flyteidl/clients/go/coreutils/literals.go)

### FlytePropeller

1. Implement a new type checker in [flytepropeller/pkg/compiler/validators/typing.go](https://github.com/flyteorg/flyte/blob/master/flytepropeller/pkg/compiler/validators/typing.go). We need to discuss whether to check the name of each field and the tuple itself, as this will affect the casting between NamedTuple and Tuple. (#TO_DISCUSS)

   ```go
    type tupleTypeChecker struct {
        literalType *flyte.LiteralType
    }

    func (t tupleTypeChecker) CastsFrom(upstreamType *flyte.LiteralType) bool {
        tupleType := t.literalType.GetTupleType()
        upstreamTupleType := upstreamType.GetTupleType()

        if upstreamTupleType != nil && tupleType != nil {
            // TO_DISCUSS: Should we check the name of the tuple?
            if tupleType.GetTupleName() != upstreamTupleType.GetTupleName() {
                return false
            }

            // For each field in the tuple, check if the upstream type can be casted to the downstream type
            tupleFields := tupleType.GetFields()
            upstreamTupleFields := upstreamTupleType.GetFields()
            if len(tupleFields) == len(upstreamTupleFields) {
                for i, downstreamType := range tupleFields {
                    // TO_DISCUSS: Should we check the name of each field?
                    if downstreamType.GetName() != upstreamTupleFields[i].GetName() {
                        return false
                    }

                    // Check if the upstream type can be casted to the downstream type
                    if !getTypeChecker(downstreamType).CastsFrom(upstreamTupleFields[i]) {
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

2. Update the collection bindings type validation code in [flytepropeller/pkg/compiler/validators/bindings.go](https://github.com/flyteorg/flyte/blob/master/flytepropeller/pkg/compiler/validators/bindings.go). Similarly, we need to discuss whether we should keep the names of each field and the tuple. (#TO_DISCUSS)

    ```go
    func validateBinding(w c.WorkflowBuilder, nodeID c.NodeID, nodeParam string, binding *flyte.BindingData,
        expectedType *flyte.LiteralType, errs errors.CompileErrors, validateParamTypes bool) (
        resolvedType *flyte.LiteralType, upstreamNodes []c.NodeID, ok bool) {

        // ...

        switch val := binding.GetValue().(type) {
        case *flyte.BindingData_Collection:
            if val.Collection == nil {
                errs.Collect(errors.NewParameterNotBoundErr(nodeID, nodeParam))
                return nil, nil, !errs.HasErrors()
            }

            if expectedType.GetCollectionType() != nil {
                // ...
            }

            if expectedType.GetTupleType() != nil {
                allNodeIds := make([]c.NodeID, 0, len(val.Collection.GetBindings()))
                var fieldTypes []*flyte.TupleFieldType
                for i, v := range val.Collection.GetBindings() {
                    if resolvedType, nodeIds, ok := validateBinding(w, nodeID, nodeParam, v, expectedType.GetTupleType().GetFields()[i].GetType(), errs.NewScope(), validateParamTypes); ok {
                        allNodeIds = append(allNodeIds, nodeIds...)
                        fieldTypes = append(fieldTypes, &flyte.TupleFieldType{
                            Name: expectedType.GetTupleType().GetFields()[i].GetName(),   // TO_DISCUSS: Should we keep the name of each field?
                            Type: resolvedType,
                        })
                    }
                }

                return &flyte.LiteralType{
                    Type: &flyte.LiteralType_TupleType{
                        TupleType: &flyte.TupleType{
                            TupleName: expectedType.GetTupleType().GetTupleName(),  // TO_DISCUSS: Should we keep the name of the tuple?
                            Fields: fieldTypes,
                    },
                    }, allNodeIds, !errs.HasErrors()
                }
            }

            errs.Collect(errors.NewMismatchingBindingsErr(nodeID, nodeParam, expectedType.String(), val.Collection.String()))
        // ...
        }
        // ...
    }
    ```

3. Update the type metadata stripping code in [flytepropeller/pkg/compiler/transformers/k8s/utils.go](https://github.com/flyteorg/flyte/blob/master/flytepropeller/pkg/compiler/transformers/k8s/utils.go).

    ```go
    func stripTypeMetadata(t *flyte.LiteralType) *flyte.LiteralType {
        // ...
        switch underlyingType := c.Type.(type) {
        case *flyte.LiteralType_TupleType:
            fields := make([]*flyte.TupleFieldType, 0, len(t.GetTupleType().GetFields()))
            for _, f := range t.GetTupleType().GetFields() {
                fields = append(fields, &flyte.TupleFieldType{
                    Name: f.GetName(),
                    Type: StripTypeMetadata(f.GetType()),
                })
            }
            return &flyte.LiteralType{
                Type: &flyte.LiteralType_TupleType{
                    TupleType: &flyte.TupleType{
                        TupleName: t.GetTupleType().GetTupleName(),
                        Fields: fields,
                    },
                },
            }
        }
        // ...
    }
    ```

4. Update `LiteralTypeForLiteral()` and `buildMultipleTypeUnion()` functions in [flytepropeller/pkg/compiler/validators/utils.go](https://github.com/flyteorg/flyte/blob/master/flytepropeller/pkg/compiler/validators/utils.go) to correctly handle the type.

### FlyteKit

- Supporting the following types in Flytekit:
    - `NamedTuple("NAME", ("K1": T1), ("K2": T2))`
    - `Tuple[T1, T2]`
    - [#TO_DISCUSS] `Tuple[T1, ...]`: There are two potential approaches to support this type:
        - Add a field `is_univariate` in `TupleType`.
        - Treat it as a list, given that the performance is quite similar to `List[T1]`. However, this approach may lead to confusion with `List[T1]`.

## 4 Metrics & Dashboards

None

## 5 Drawbacks

None

## 6 Alternatives

None

## 7 Potential Impact and Dependencies

- This feature will affect FlyteIDL, FlytePropeller, and FlyteKit.
- It will enable the support of NamedTuple and Tuple in Flytekit, which are common types in Python.

## 8 Unresolved questions

1. Should we check the names of each field and the tuple when casting between `NamedTuple` and `Tuple`?
2. Should we use `is_univariate` to support `Tuple[int, ...]`, given its similarity to `List[int]`?
3. The specification of the new IDL for the Tuple type is open for further discussion.

## 9 Conclusion

We need a new IDL for the Tuple type to support the usage of `NamedTuple` and `Tuple` in Flytekit.
