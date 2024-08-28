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

- A new `LiteralType` and `Literal` will be introduced to represent tuple types in FlyteIDL. Each element in the tuple will be considered independently, preserving the order of elements. For NamedTuple, the name of each element will be stored along with its type, while for regular tuples, the names will be left empty.

    - `LiteralType`
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
    - `Literal`
        ```proto
        message LiteralField {
            string name = 1;
            Literal value = 2;
        }

        message LiteralFieldCollection {
            string tuple_name = 1;
            repeated LiteralField fields = 1;
        }

        message Literal {
            oneof value {
                // ...
                // A tuple of literals.
                LiteralFieldCollection tuple = 4;
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

- The above code should be added to [flyteidl/protos/flyteidl/core/types.proto](https://github.com/flyteorg/flyte/blob/master/flyteidl/protos/flyteidl/core/types.proto).

#### Literal of NamedTuple & Tuple

- <s>Since the type, including the names of the fields and the tuple, is stored in `LiteralType`, we only need to store the value of each field in the tuple. We can use the existing `LiteralCollection` message from [flyteidl/protos/flyteidl/core/literals.proto](https://github.com/flyteorg/flyte/blob/master/flyteidl/protos/flyteidl/core/literals.proto) to store the value of each field.</s>

- The approach above is incorrect since we couldn't get the literal type from the literal in [flytepropeller/pkg/compiler/validators/utils.go](https://github.com/flyteorg/flyte/blob/master/flytepropeller/pkg/compiler/validators/utils.go). Also, we need to keep the names of each field and the tuple in the literal so that given the literal, we can restore the literal type. (#TO_DISCUSS: Is there any better way to do this?)

    ```proto
    message LiteralField {
        string name = 1;
        Literal value = 2;
    }

    message LiteralFieldCollection {
        string tuple_name = 1;
        repeated LiteralField fields = 1;
    }

    message Literal {
        oneof value {
            // ...
            // A tuple of literals.
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
        string tuple_name = 1;
        repeated BindingDataField fields = 1;
    }
    
    message BindingData {
        oneof value {
            // ...
            // A tuple collection of bindings. This allows nesting of binding data to any number of levels.
            BindingDataFieldCollection tuple = 5;
        }

        UnionInfo union = 5;
    }
    ```

- The above code should be added to [flyteidl/protos/flyteidl/core/literals.proto](https://github.com/flyteorg/flyte/blob/master/flyteidl/protos/flyteidl/core/literals.proto).

#### Code that may affect Flytectl

- Update `MakeDefaultLiteralForType()` function in [flyteidl/clients/go/coreutils/literals.go](https://github.com/flyteorg/flyte/blob/master/flyteidl/clients/go/coreutils/literals.go)
    
    ```go
    func MakeDefaultLiteralForType(typ *core.LiteralType) (*core.Literal, error) {
        switch t := typ.GetType().(type) {
        // ...
        case *core.LiteralType_TupleType:
            fields := make([]*core.LiteralField, 0, len(t.TupleType.Fields))
            for _, f := range t.TupleType.Fields {
                val, err := MakeDefaultLiteralForType(f.Type)
                if err != nil {
                    return nil, errors.Errorf("Failed to create default literal for the field [%v] with type [%v]", f.Name, f.Type)
                }
                fields = append(fields, &core.LiteralField{
                    Name: f.Name,
                    Value: val,
                })
            }
            return &core.Literal{
                Value: &core.Literal_Tuple{
                    Tuple: &core.LiteralFieldCollection{
                        TupleName: t.TupleType.TupleName,
                        Fields: fields,
                    },
                },
            }, nil
        }
        // ...
    }
    ```

-  Update `MakeLiteralForType()`function in [flyteidl/clients/go/coreutils/literals.go](https://github.com/flyteorg/flyte/blob/master/flyteidl/clients/go/coreutils/literals.go). We need to discuss the tuple interface in `Inputs` field of flytectl to write the correct code. 

    ```go
    func MakeLiteralForType(t *core.LiteralType, v interface{}) (*core.Literal, error) {
        l := &core.Literal{}
	    switch newT := t.Type.(type) {
        // ...
        case *core.LiteralType_TupleType:
            // TO_DISCUSS: How to handle the tuple interface in Inputs field of flytectl?
            // 
            // [Example usage]
            // inputs:
            //     tuple_data:
            //         tuple_name: "my-tuple"
            //         fields:
            //           - name: "key1"
            //             value: "foo"
            //           - name: "key2"
            //             value: 123

            vMap, ok := v.(map[string]interface{})
            if !ok {
                return nil, errors.Errorf("Expected a map[string]interface{} for tuple type, got [%v]", v)
            }
            tupleName, ok := vMap["tuple_name"].(string)
            if !ok {
                return nil, errors.Errorf("Expected a string for tuple_name, got [%v]", vMap["tuple_name"])
            }
            fields, ok := vMap["fields"].([]interface{})
            if !ok {
                return nil, errors.Errorf("Expected a []interface{} for fields, got [%v]", vMap["fields"])
            }
            if len(fields) != len(newT.TupleType.Fields) {
                return nil, errors.Errorf("Expected [%v] fields, got [%v]", len(newT.TupleType.Fields), len(fields))
            }

            literalFields := make([]*core.LiteralField, 0, len(fields))
            for i, f := range fields {
                fieldMap, ok := f.(map[string]interface{})
                if !ok {
                    return nil, errors.Errorf("Expected a map[string]interface{} for field, got [%v]", f)
                }
                name, ok := fieldMap["name"].(string)
                if !ok {
                    return nil, errors.Errorf("Expected a string for name, got [%v]", fieldMap["name"])
                }
                value, ok := fieldMap["value"].(interface{})
                if !ok {
                    return nil, errors.Errorf("Expected an interface{} for value, got [%v]", fieldMap["value"])
                }
                val, err := MakeLiteralForType(newT.TupleType.Fields[i].Type, value)
                if err != nil {
                    return nil, errors.Wrapf(err, "Failed to create literal for field [%v]", name)
                }
                literalFields = append(literalFields, &core.LiteralField{
                    Name:  name,
                    Value: val,
                })
            }
            return &core.Literal{
                Value: &core.Literal_Tuple{
                    Tuple: &core.LiteralFieldCollection{
                        TupleName: tupleName,
                        Fields:    literalFields,
                    },
                },
            }, nil
        // ...
        }
        // ...
    }
    ```

### FlyteAdmin & FlytePropeller

1. Update `LiteralTypeForLiteral()` functions in [flytepropeller/pkg/compiler/validators/utils.go](https://github.com/flyteorg/flyte/blob/master/flytepropeller/pkg/compiler/validators/utils.go) to correctly handle the type.

    ```go
    func TupleFieldTypesForLiteralFields(fields []*core.LiteralField) []*core.TupleFieldType {
        tupleFields := make([]*core.TupleFieldType, 0, len(fields))
        for _, f := range fields {
            tupleFields = append(tupleFields, &core.TupleFieldType{
                Name: f.Name,
                Type: LiteralTypeForLiteral(f.Value),
            })
        }
        return tupleFields
    }

    func LiteralTypeForLiteral(l *core.Literal) *core.LiteralType {
        switch l.GetValue().(type) {
        // ...
        case *core.Literal_Tuple:
            fields := TupleFieldTypesForLiteralFields(l.GetTuple().Fields)
            return &core.LiteralType{
                Type: &core.LiteralType_TupleType{
                    TupleType: &core.TupleType{
                        TupleName: l.GetTuple().TupleName,
                        Fields:    fields,
                    },
                },
            }
        }
    }
    ```


2. Implement a new type checker in [flytepropeller/pkg/compiler/validators/typing.go](https://github.com/flyteorg/flyte/blob/master/flytepropeller/pkg/compiler/validators/typing.go). We need to discuss whether to check the name of each field and the tuple itself, as this will affect the casting between NamedTuple and Tuple. (#TO_DISCUSS)

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

3. Update the collection bindings type validation code in [flytepropeller/pkg/compiler/validators/bindings.go](https://github.com/flyteorg/flyte/blob/master/flytepropeller/pkg/compiler/validators/bindings.go). Similarly, we need to discuss whether we should keep the names of each field and the tuple. (#TO_DISCUSS)

    ```go
    func validateBinding(w c.WorkflowBuilder, nodeID c.NodeID, nodeParam string, binding *flyte.BindingData,
        expectedType *flyte.LiteralType, errs errors.CompileErrors, validateParamTypes bool) (
        resolvedType *flyte.LiteralType, upstreamNodes []c.NodeID, ok bool) {

        // ...

        switch val := binding.GetValue().(type) {
        // ...
        case *flyte.BindingData_Tuple:
            if val.Tuple == nil {
                errs.Collect(errors.NewParameterNotBoundErr(nodeID, nodeParam))
                return nil, nil, !errs.HasErrors()
            }

            // TODO: Add other checking
            // 1. Check the names of the fields and tuple
            // 2. Check the length of the field from val and expectedType

            if expectedType.GetTupleType() != nil {
                allNodeIds := make([]c.NodeID, 0, len(val.Tuple.GetBindings()))
                var fieldTypes []*flyte.TupleFieldType
                for i, f := range val.Tuple.GetFields() {
                    if resolvedType, nodeIds, ok := validateBinding(w, nodeID, nodeParam, f.getBinding(), expectedType.GetTupleType().GetFields()[i].GetType(), errs.NewScope(), validateParamTypes); ok {
                        allNodeIds = append(allNodeIds, nodeIds...)
                        fieldTypes = append(fieldTypes, &flyte.TupleFieldType{
                            Name: expectedType.GetTupleType().GetFields()[i].GetName(),
                            Type: resolvedType,
                        })
                    }
                }

                return &flyte.LiteralType{
                    Type: &flyte.LiteralType_TupleType{
                        TupleType: &flyte.TupleType{
                            TupleName: expectedType.GetTupleType().GetTupleName(), 
                            Fields: fieldTypes,
                    },
                    }, allNodeIds, !errs.HasErrors()
                }
            }

            errs.Collect(errors.NewMismatchingBindingsErr(nodeID, nodeParam, expectedType.String(), val.Tuple.String()))
        // ...
        }
        // ...
    }
    ```

4. Update the type metadata stripping code in [flytepropeller/pkg/compiler/transformers/k8s/utils.go](https://github.com/flyteorg/flyte/blob/master/flytepropeller/pkg/compiler/transformers/k8s/utils.go).

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

### FlyteKit

- Supporting the following types in Flytekit:
    - `NamedTuple("NAME", ("K1": T1), ("K2": T2))`
    - `Tuple[T1, T2]`
    - [#TO_DISCUSS] `Tuple[T1, ...]`: There are two potential approaches to support this type:
        - Add a field `is_univariate` in `TupleType` and `TupleFieldType`.
            ```proto
            // types.proto
            message TupleType {
                string tuple_name = 1;
                bool is_univariate = 2;
                repeated TupleFieldType fields = 3;
            }

            // literals.proto
            message LiteralFieldCollection {
                string tuple_name = 1;
                bool is_univariate = 2;
                repeated LiteralField fields = 1;
            }
            ```
        - Treat it as a list, given that the performance is quite similar to `List[T1]`. However, this approach may lead to confusion with `List[T1]`.

## 4 Metrics & Dashboards

None

## 5 Drawbacks

None

## 6 Alternatives

None

## 7 Potential Impact and Dependencies

- This feature will affect FlyteIDL, FlytePropeller, FlyteKit, Flytectl, and FlyteConsole.
- It will enable the support of NamedTuple and Tuple in Flytekit, which are common types in Python.

## 8 Unresolved questions

1. Should we check the names of each field and the tuple when casting between `NamedTuple` and `Tuple`?
2. Should we use `is_univariate` to support `Tuple[int, ...]`, given its similarity to `List[int]`?
3. [Literal of NamedTuple & Tuple](#literal-of-namedtuple--tuple): Is there any better way to store the literal of NamedTuple and Tuple? Is it necessary to keep names of the fields and tuple in both `Literal` and `LiteralType`?
4. The specification of the new IDL for the Tuple Literal and LiteralType is open for further discussion.

## 9 Conclusion

We need a new IDL for the Tuple Literal and LiteralType to support the usage of `NamedTuple` and `Tuple` in Flytekit.
