# Tuple IDL

**Authors:**

- [@Chun-Mao (Michael) Lai](https://github.com/Mecoli1219)

## 1 Executive Summary

### Goals

We want to add support for tuple (`typing.Tuple`, `tuple`) and namedtuple (`typing.NamedTuple`) types in Flytekit.

- NamedTuple:
  - A named tuple consists of multiple named values, each with its own type.
  - Python example: `typing.NamedTuple("NAME", ("K1", T1), ("K2", T2))`
  - `collections.namedtuple` is not supported in Flytekit because it lacks type annotations for the tuple’s elements.
- Tuple:
  - A tuple contains multiple values, each with its own type.
  - Python example: `typing.Tuple[T1, T2, T3]`, `tuple[T1, T2, T3]`
  - `typing.Tuple[T1, ...]` is excluded from this RFC, as its behavior differs significantly from that of tuple and NamedTuple.

### Literal Value

We introduce a new Literal called `LiteralTupleMap` to store the values of `NamedTuple` and `Tuple`. This message includes three fields:

1. `tuple_name`: The name of the tuple.
2. `order`: A list of strings to store the order of the fields in the tuple.
3. `literals`: A map of field names to their corresponding literals.

### Literal Type

We introduce a new LiteralType called `TupleType` to store the type of `NamedTuple` and `Tuple`. This message includes three fields:

1. `tuple_name`: The name of the tuple.
2. `orders`: A list of strings to store the order of the fields in the tuple.
3. `fields`: A map of field names to their corresponding types.

### Notes:

1. The `tuple_name` and `orders` fields in `LiteralTupleMap` and `TupleType` should be the same.
2. For tuple, the `tuple_name` will be left as an empty string, and Flytekit will automatically generate names for each field.

## 2 Motivation

Before this RFC, Flytekit restricts the support of tuple and NamedTuple types, which are commonly used in Python. This RFC aims to add support for tuple types in Flyte system.

Note: Several issues regarding `NamedTuple` and `tuple` support in Flytekit have been identified:

- [#1337](https://github.com/flyteorg/flyte/issues/1337)
- [#3158](https://github.com/flyteorg/flyte/issues/3158)
- [#4358](https://github.com/flyteorg/flyte/issues/4358)

This feature aims to resolve all these issues.

## 3 Proposed Implementation

### FlyteKit Examples

#### `NamedTuple("NAME", ("K1": T1), ("K2": T2))`

```python
from typing import NamedTuple

class MyNamedTuple(NamedTuple):
    key1: int
    key2: str

@task
def my_task(a: MyNamedTuple):
    # protobuf tuple value for input a:
    # {
    #     "tuple_name": "MyNamedTuple"
    #     "orders": ["key1", "key2"]
    #     "literals": {
    #         "key1": 1,
    #         "key2": "foo"
    #     }
    # }
    ...
```

#### `Tuple[T1, T2]`

```python
from typing import Tuple

@task
def my_task(a: Tuple[int, str]):
    # protobuf tuple value for input a:
    # {
    #     "tuple_name": ""
    #     "orders": ["t0", "t1"]
    #     "literals": {
    #         "t0": 1,
    #         "t1": "foo"
    #     }
    # }
    ...
```

### FlyteIDL

To support `NamedTuple`, we need to store both the name of each field and the tuple itself, while ensuring the field order is preserved. We propose a new IDL structure for the Tuple `Literal` and `LiteralType`.

#### Literal Value

```proto
message LiteralTupleMap {
    // The name of the NamedTuple. If it is original tuple, it would be empty string.
    string tuple_name = 1;

    // The order of each fields stored in the tuple.
    repeated string order = 2;

    // A map of literals.
    map<string, Literal> literals = 3;
}

message Literal {
    oneof value {
        // ...
        LiteralTupleMap tuple = 9;
    }
    // ...
}
```

#### Literal Type

```proto
message TupleType {
    // The name of the NamedTuple. If it is original tuple, it would be empty string.
    string tuple_name = 1;

    // The order of each fields stored in the tuple.
    repeated string order = 2;

    // A map of types.
    map<string, LiteralType> fields = 3;
}

message LiteralType {
    oneof type {
        // ...
        TupleType tuple_type = 12;
    }
    // ...
}
```

Both `tuple_name` and `order` in `LiteralTupleMap` and `TupleType` should be identical for the same tuple. For regular `Tuple`, the `tuple_name` will be left empty, and Flytekit will automatically generate names for each field (e.g., `t0`, `t1`, `t2`, etc.).

#### Other Considerations

An alternative approach to storing the literal values for `NamedTuple` and `Tuple` would be combining the order and the map into a list of messages, as shown below:

```proto
// !!! This is not the proposed implementation !!!
message FakeLiteralTupleField {
    string name = 1;
    Literal value = 2;
}

message FakeLiteralTupleMap {
    string tuple_name = 1;
    repeated FakeLiteralTupleField fields = 2;
}
```

While this structure may seem more straightforward, it has a drawback: looking up a field’s value by name requires iterating through the list, which results in O(n) complexity. In contrast, using a map allows for more efficient O(1) lookup times when retrieving the value of a field by its name. Although in practice the tuple size is usually small, the map-based approach is more scalable and efficient.

### FlytePropeller

#### General Support for Literal and LiteralType

1. Create a Literal Type for Literal when doing type validation.

   ```go
   func TupleFieldTypesForLiterals(fields map[string]*core.Literal) map[string]*core.LiteralType {
       res := make(map[string]*core.LiteralType, len(fields))
       for k, v := range fields {
           res[k] = LiteralTypeForLiteral(v)
       }

       return res
   }

   func LiteralTypeForLiteral(l *core.Literal) *core.LiteralType {
       switch l.GetValue().(type) {
       // ...
       case *core.Literal_Tuple:
           fields := TupleFieldTypesForLiterals(l.GetTuple().Literals)
           return &core.LiteralType{
               Type: &core.LiteralType_TupleType{
                   TupleType: &core.TupleType{
                       TupleName: l.GetTuple().GetTupleName(),
                       Order:     l.GetTuple().GetOrder(),
                       Fields:    fields,
                   },
               },
           }
       }
   }
   ```

2. Support input and default input.

   ```go
   // create a literal for TupleType
   func MakeLiteralForType(t *core.LiteralType, v interface{}) (*core.Literal, error) {
       l := &core.Literal{}
       switch newT := t.Type.(type) {
       // ...
       case *core.LiteralType_TupleType:
           // In this part, the LiteralType is already given, so the name of the tuple and the order of each field should be already defined. The only thing that is missing is the value of each field.
           // Therefore, we only need to provide the value of each tuple field (via the key-value pair of each field) in flytectl for the registered task or workflow for further execution needs.
           //
           // [Example usage]
           // inputs:
           //     tuple_input:
           //         key1: "foo"
           //         key2: 123

           vMap, ok := v.(map[string]interface{})
           if !ok {
               return nil, errors.Errorf("Expected a map[string]interface{} for tuple type, got [%v]", v)
               }
           // check whether all the key provided by vMap is valid.
           for key := range vMap {
               if _, ok := t.GetTupleType().GetFields()[key]; !ok {
                   return nil, fmt.Errorf("key %s not found in tuple type", key)
               }
           }

           literals := make(map[string]*core.Literal, len(vMap))
           // iterate over the fields in the tuple type
           for key, fieldType := range t.GetTupleType().GetFields() {
               l, err := MakeLiteralForType(fieldType, vMap[key])
               if err != nil {
                   return nil, err
               }
               literals[key] = l
           }
           l = &core.Literal{
               Value: &core.Literal_Tuple{
                   Tuple: &core.LiteralTupleMap{
                       TupleName: t.GetTupleType().GetTupleName(),
                       Order:     t.GetTupleType().GetOrder(),
                       Literals:  literals,
                   },
               },
           }
           return l, nil
       // ...
       }
       // ...
   }

   // Default Literal for TupleType
   func MakeDefaultLiteralForType(typ *core.LiteralType) (*core.Literal, error) {
       switch t := typ.GetType().(type) {
           // ...
           case *core.LiteralType_TupleType:
               return MakeLiteralForType(typ, nil)
       }
       // ...
   }
   ```

3. Implement a new type checker for compiler. We need to discuss whether to check the name of each field and the tuple itself, as this will affect the casting between NamedTuple and Tuple.

   ```go
   type tupleTypeChecker struct {
       literalType *flyte.LiteralType
   }

   func (t tupleTypeChecker) CastsFrom(upstreamType *flyte.LiteralType) bool {
       tupleType := t.literalType.GetTupleType()
       upstreamTupleType := upstreamType.GetTupleType()
       if upstreamTupleType != nil {
           // check order
           if len(upstreamTupleType.GetOrder()) != len(tupleType.GetOrder()) {
               return false
           }
           for i, upstreamField := range upstreamTupleType.GetOrder() {
               if upstreamField != tupleType.GetOrder()[i] {
                   return false
               }
           }

           if len(upstreamTupleType.GetFields()) == len(tupleType.GetFields()) && upstreamTupleType.GetTupleName() == tupleType.GetTupleName() {
               for k, downstreamType := range tupleType.GetFields() {
                   if upstreamFieldType, ok := upstreamTupleType.GetFields()[k]; !ok || !getTypeChecker(downstreamType).CastsFrom(upstreamFieldType) {
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

4. Strip metadata from the tuple type for workflow building in compiler.

   ```go
   func stripTypeMetadata(t *flyte.LiteralType) *flyte.LiteralType {
       // ...
       switch underlyingType := c.Type.(type) {
       case *core.LiteralType_TupleType:
           fields := make(map[string]*core.LiteralType, len(c.GetTupleType().Fields))
           for k, field := range c.GetTupleType().Fields {
               fields[k] = StripTypeMetadata(field)
           }

           underlyingType.TupleType.TupleName = c.GetTupleType().TupleName
           underlyingType.TupleType.Order = c.GetTupleType().Order
           underlyingType.TupleType.Fields = fields
       }
       // ...
   }
   ```

5. Visualize the tuple type in the Graphviz.

   ```go
   func flatten(binding *core.BindingData, flatMap map[common.NodeID]sets.String) {
       switch binding.GetValue().(type) {
       // ...
       case *core.BindingData_Tuple:
           for _, v := range binding.GetTuple().GetBindings() {
               flatten(v, flatMap)
           }
       }
       // ...
   }
   ```

#### Tuple Binding and Access with Attribute Paths

In `tuple` and `NamedTuple`, we can access the value of each field by its name or index. For instance:

- `NamedTuple`:
  We can access the value of each field by its name and index.

  ```python
  from typing import NamedTuple

  class MyNamedTuple(NamedTuple):
      key1: int
      key2: str

  @task
  def my_task() -> MyNamedTuple:
      return MyNamedTuple(1, "foo")

  @workflow
  def my_wf() -> tuple[int, str]:
      a = my_task()

      # We should be able to access the value of each field by its name or index.
      # Access the value of each field by its name
      int_result = a.key1
      str_result = a.key2

      # Access the value of each field by its index
      int_result = a[0]
      str_result = a[1]

      return int_result, str_result
  ```

- `Tuple`:
  We can only access the value of each field by its index.

  ```python
  from typing import Tuple

  @task
  def my_task() -> Tuple[int, str]:
      return 1, "foo"

  @workflow
  def my_wf() -> tuple[int, str]:
      a = my_task()

      # We should be able to access the value of each field by its index.
      int_result = a[0]
      str_result = a[1]

      return int_result, str_result
  ```

Therefore, we need to support both integer and string attribute paths for `tuple` and `NamedTuple` in FlytePropeller. In order to do this, we need to:

1. Validate the binding data for the tuple type and the Promise binding with attribute paths of the tuple type in compiler.

   ```go
   func validateBinding(w c.WorkflowBuilder, nodeID c.NodeID, nodeParam string, binding *flyte.BindingData,
       expectedType *flyte.LiteralType, errs errors.CompileErrors, validateParamTypes bool) (
       resolvedType *flyte.LiteralType, upstreamNodes []c.NodeID, ok bool) {
       // ...
       switch val := binding.GetValue().(type) {
       // ...
       case *flyte.BindingData_Promise:
           // ...
           for _, attr := range val.Promise.AttrPath {
               var tmpType *flyte.LiteralType
               var exist bool
               if sourceType.GetCollectionType() != nil {
               // ...
               } else if sourceType.GetTupleType() != nil {
                   var key string

                   if attr.GetStringValue() != "" {
                       // If the attribute path is a string, we can use it as the key to access the value of the field.
                       key = attr.GetStringValue()
                   } else {
                       // If the attribute path is an integer, we have to get the key by the order of the tuple first, then use the key to access the value of the field.
                       key = sourceType.GetTupleType().Order[attr.GetIntValue()]
                   }
                   sourceType, exist = sourceType.GetTupleType().GetFields()[key]
                   if !exist {
                       errs.Collect(errors.NewFieldNotFoundErr(nodeID, val.Promise.Var, sourceType.String(), key))
                       return nil, nil, !errs.HasErrors()
                   }
               } // ...
           }


       // ...
       case *flyte.BindingData_Tuple:
           if val.Tuple == nil {
               errs.Collect(errors.NewParameterNotBoundErr(nodeID, nodeParam))
               return nil, nil, !errs.HasErrors()
           }

           // validate the tuple name
           if val.Tuple.TupleName != expectedType.GetTupleType().TupleName {
               errs.Collect(errors.NewMismatchingBindingsErr(nodeID, nodeParam, expectedType.String(), val.Tuple.String()))
               return nil, nil, !errs.HasErrors()
           }

           // validate the order of the tuple
           if len(val.Tuple.Order) != len(expectedType.GetTupleType().Order) {
               errs.Collect(errors.NewMismatchingBindingsErr(nodeID, nodeParam, expectedType.String(), val.Tuple.String()))
               return nil, nil, !errs.HasErrors()
           }
           for i, v := range val.Tuple.Order {
               if v != expectedType.GetTupleType().Order[i] {
                   errs.Collect(errors.NewMismatchingBindingsErr(nodeID, nodeParam, expectedType.String(), val.Tuple.String()))
                   return nil, nil, !errs.HasErrors()
               }
           }

           // validate the fields of the tuple
           if expectedType.GetTupleType() != nil {
               allNodeIds := make([]c.NodeID, 0, len(val.Tuple.GetBindings()))
               fields := make(map[string]*flyte.LiteralType, len(val.Tuple.GetBindings()))
               if len(val.Tuple.GetBindings()) != len(expectedType.GetTupleType().GetFields()) {
                   errs.Collect(errors.NewMismatchingBindingsErr(nodeID, nodeParam, expectedType.String(), val.Tuple.String()))
                   return nil, nil, !errs.HasErrors()
               }
               for k, v := range val.Tuple.GetBindings() {
                   if expectedField, ok := expectedType.GetTupleType().GetFields()[k]; !ok {
                       errs.Collect(errors.NewFieldNotFoundErr(nodeID, nodeParam, expectedType.String(), k))
                   } else {
                       if fieldType, nodeIds, ok := validateBinding(w, node, nodeParam, v, expectedField, errs.NewScope(), validateParamTypes); ok {
                           allNodeIds = append(allNodeIds, nodeIds...)
                           fields[k] = fieldType
                       }
                   }
               }

               return &flyte.LiteralType{
                   Type: &flyte.LiteralType_TupleType{
                       TupleType: &flyte.TupleType{
                           TupleName: expectedType.GetTupleType().GetTupleName(),
                           Order:     expectedType.GetTupleType().GetOrder(),
                           Fields:    fields,
                       },
                   },
               }, allNodeIds, !errs.HasErrors()
           }
       // ...
       }
       // ...
   }
   ```

2. Resolve the attribute paths of the tuple type in promise for the controller.

   ```go
   func resolveAttrPathInPromise(nodeID string, literal *core.Literal, bindAttrPath []*core.PromiseAttribute) (*core.Literal, error) {
       // ...
       for _, attr := range bindAttrPath {
   	    switch currVal.GetValue().(type) {
           // ...
           case *core.Literal_Tuple:
               var key string
               if attr.GetStringValue() != "" {
                   // If the attribute path is a string, we can use it as the key to access the value of the field.
                   key = attr.GetStringValue()
               } else {
                   // If the attribute path is an integer, we have to get the key by the order of the tuple first, then use the key to access the value of the field.
                   if int(attr.GetIntValue()) >= len(currVal.GetTuple().GetOrder()) {
                       return nil, errors.Errorf(errors.PromiseAttributeResolveError, nodeID, "index [%v] is out of range of %v", attr.GetIntValue(), currVal.GetTuple().GetOrder())
                   }
                   key = currVal.GetTuple().GetOrder()[attr.GetIntValue()]
               }
               tmpVal, exist = currVal.GetTuple().GetLiterals()[key]
               if !exist {
                   return nil, errors.Errorf(errors.PromiseAttributeResolveError, nodeID, "key [%v] does not exist in literal %v", key, currVal.GetTuple().GetLiterals())
               }
               currVal = tmpVal
               count++
           // ...
           }
       }
       // ...
   }
   ```

3. Resolve the binding data for tuple into Literal.

   ```go
   func ResolveBindingData(ctx context.Context, outputResolver OutputResolver, nl executors.NodeLookup, bindingData *core.BindingData) (*core.Literal, error) {
       // ...
       switch bindingData.GetValue().(type) {
       // ...
       case *core.BindingData_Tuple:
           logger.Debugf(ctx, "bindingData.GetValue() [%v] is of type Tuple", bindingData.GetValue())
           literalMap := make(map[string]*core.Literal, len(bindingData.GetTuple().GetBindings()))
           for k, v := range bindingData.GetTuple().GetBindings() {
               l, err := ResolveBindingData(ctx, outputResolver, nl, v)
               if err != nil {
                   logger.Debugf(ctx, "Failed to resolve binding data. Error: [%v]", err)
                   return nil, err
               }
               literalMap[k] = l
           }

           literal.Value = &core.Literal_Tuple{
               Tuple: &core.LiteralTupleMap{
                   TupleName: bindingData.GetTuple().GetTupleName(),
                   Order:     bindingData.GetTuple().GetOrder(),
                   Literals:  literalMap,
               },
           }
       // ...
       }
       // ...
   }
   ```

4. Update binding node ID for dynamic workflow.

   ```go
   func updateBindingNodeIDsWithLineage(parentNodeID, retryAttempt string, binding *core.BindingData) (err error) {
       switch b := binding.Value.(type) {
       // ...
       case *core.BindingData_Tuple:
           for _, item := range b.Tuple.Bindings {
               err = updateBindingNodeIDsWithLineage(parentNodeID, retryAttempt, item)
               if err != nil {
                   return err
               }
           }
       }
       // ...
   }
   ```

### FlyteKit

#### TupleTransformer

We have to create a new TypeTransformer `TupleTransformer` to support attribute access for `NamedTuple` and `Tuple`.

This transformer support three types of tuple, `typing.Tuple`, `tuple`, and `typing.NamedTuple`. Currently, since `collections.namedtuple` is not type annotated, so we shouldn't allow it in Flytekit.

##### Literal Value

- `to_literal`:
  - We can extract each field in the Literal from the NamedTuple and Tuple as follows:
    - `tuple_name`:
      - `NamedTuple`: The name of the NamedTuple.
      - `Tuple`: An empty string.
    - `order` & `literals`:
      - `Tuple`: We first automatically generate names for each field (e.g., `t0`, `t1`, `t2`, etc.), then extract the order and literals map.
      - `NamedTuple`: We use the field names directly to extract the order and literals map.
  - Examples:
    - `Tuple`: `Tuple[int, str]` -> `LiteralTupleMap(tuple_name="", order=["t0", "t1"], literals={"t0": 1, "t1": "foo"})`
    - `NamedTuple`: `NamedTuple("NAME", ("K1", int), ("K2", str))` -> `LiteralTupleMap(tuple_name="NAME", order=["K1", "K2"], literals={"K1": 1, "K2": "foo"})`
- `to_python_val`:
  - Since we have the order of each field and the map of literals, we can easily convert the `LiteralTupleMap` to a `NamedTuple` or `Tuple`.
  - The `tuple_name` and the field name in the `LiteralTupleMap` for `Tuple` is not necessary to recover the original tuple, so we can ignore them.

##### Literal Type

- `get_literal_type`:
  - Similar to `to_literal`, we will first generate the `tuple_name` and the field names for `Tuple`. Then, we extract the order and the map of types for each field.
  - Examples:
    - `Tuple`: `Tuple[int, str]` -> `TupleType(tuple_name="", order=["t0", "t1"], fields={"t0": int, "t1": str})`
    - `NamedTuple`: `NamedTuple("NAME", ("K1", int), ("K2", str))` -> `TupleType(tuple_name="NAME", order=["K1", "K2"], fields={"K1": int, "K2": str})`
- `guess_python_type`:
  - We use the `tuple_name` as the identifier for the NamedTuple and Tuple. If the `tuple_name` is empty, we will guess the type as `Tuple`. Otherwise, we will guess the type as `NamedTuple`.
  - We can easily get the each field of the NamedTuple and Tuple via the `order` and the map of types.

#### Supporting Attribute Paths for Tuple

It is commonly used to access the value of each field in `NamedTuple` and `Tuple` by its name or index. We need to support both integer and string attribute paths for `NamedTuple` and `Tuple` in FlyteKit.

### Flytectl

From Flytectl, we can construct the input of tuple for the execution.

When we received `LiteralType_TupleType`, since the `tuple_name` and the `order` are already defined in type definition, we only need to provide the value of each tuple field (via the key-value pair of each field) in Flytectl for the registered task or workflow for further execution needs.

For instance, if we want to create a tuple input `tuple_input` from flytectl, and the Literal_TupleType is as follows:

```proto
type: {
    tuple_type: {
        tuple_name: ""
        order: ["t0", "t1"]
        fields: {
            key: "t0"
            type: {
                simple: STRING
            }
        }
        fields: {
            key: "t1"
            type: {
                simple: INTEGER
            }
        }
    }
}
```

Here is an example YAML file for the input of a task with a tuple type:

```yaml
iamRoleARN: ""
inputs:
  tuple_input:
    key1: "foo"
    key2: 123
envs: {}
```

This will create a Literal for the tuple input as follows:

```proto
value: {
    tuple: {
        tuple_name: ""
        order: ["t0", "t1"]
        literals: {
            key: "t0"
            value: {
                scalar: {
                    primitive: {
                        string_value: "foo"
                    }
                }
            }
        }
        literals: {
            key: "t1"
            value: {
                scalar: {
                    primitive: {
                        integer: 123
                    }
                }
            }
        }
    }
}
```

### FlyteCopilot

We need to support handling tuple literals for the FlyteCopilot.

### FlyteConsole

#### How users can input the tuple value in the console?

We need to provide a way to let users input the tuple value in the console. This could follow the same pattern as Flytectl described above.

When we received `LiteralType_TupleType`, since the `tuple_name` and the `order` are already defined in type definition, we only need prompt the user to provide the value of each tuple field (via the key-value pair of each field) in the console for the registered task or workflow for further execution needs.

#### Display Input/Output of Tuple in the Console

Distinguishing between `NamedTuple` and `Tuple` in the console is not necessary, as the `tuple_name` and the field names are not necessary to display to the user. We can simply reconstruct the tuple value from the `order` and the map of literals in Python format like `(1, "foo")`.

#### Copying Inputs/Outputs in the Console

Allow users to copy the input/output of the tuple to the clipboard in the console.

## 4 Metrics & Dashboards

None

## 5 Drawbacks

None

## 6 Alternatives

1. There are another possible structure for the `LiteralTupleMap` and `TupleType` in FlyteIDL as discussed above in [Other Considerations](#other-considerations).
2. In Flytekit, we can separate the `NamedTuple` and `Tuple` into two different transformers, and this will make the code more readable and maintainable. However, the main problem is that the NamedTuple is not a type in Python, it is a function that returns a type. Separating them will make the registries of the transformers more complex and harder to maintain.

## 7 Potential Impact and Dependencies

- This feature will affect FlyteIDL, FlytePropeller, FlyteKit, Flytectl, and FlyteConsole.
- It will enable the support of NamedTuple and Tuple in Flytekit, which are common types in Python.

## 8 Unresolved Questions

### Univariate Tuple

This approach still couldn't support univariate tuple like `Tuple[int, ...]`. We can ask users to use `List[int]` instead. If we want to support it, we could add a field `is_univariate` in `TupleType` and `TupleFieldType`, and treat it as a list, but this will make the implementation extremely complex.

## 9 Conclusion

This RFC proposes adding support for NamedTuple and Tuple in Flytekit. We introduce a new `LiteralTupleMap` and `TupleType` in FlyteIDL to store the literal and type of NamedTuple and Tuple. We also introduce a new `TupleTransformer` in FlyteKit to support NamedTuple and Tuple. This feature will enable the support of NamedTuple and Tuple in Flyte system.
