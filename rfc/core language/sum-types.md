# Sum Types (Unions)

**Authors:**

- @maximsmol

## 1 Executive Summary

Goals:
- Implement support for sum types (also known as union types) in FlyteIDL
    - Example: `a : int | str` can assume values `10` and `"hello world"`
- Implement support for new Python types in Flytekit
    - `typing.Union`
    - `typing.Optional` (which is a special case of `typing.Union`)

Two implementation are considered.
- A tagged literal representation using a new `Literal` message (primary alternative)
    ```proto
    message LiteralSum {
        Literal value = 1;
        SumType type = 2;
        uint64 summand_idx = 3;
    }
    ```
- A type-erased literal representation where existing literals are made castable to acceptable sum types (secondary alternative)
    - Example: `10` is made castable to `int | str`, `bool | int | list str`, etc.

## 2 Motivation

Currently any type can take none values ([see this comment in Propeller's sources](https://github.com/flyteorg/flytepropeller/blob/master/pkg/compiler/validators/typing.go#L32)). This creates a few unwanted outcomes:
- The type system does not enforce required workflow parameters as always having a valid value
- Collections are not in practice homogeneous since they may contain none values
    - Example: `[1, 2, 3, null]` is a valid `list int` value since `int` is implicitly nullable
- In `Python` the use of `typing.Optional` is a no-op and furthermore not supported by default as there is no type transformer for `typing.Optional` since it would be useless
- Examining the types of workflow parameters is not enough to determine whether the parameter is intended to be optional
    - This particular point affects Latch as we generate workflow interfaces based on type information
    - Collections are the most troublesome here as it is unclear whether `list int` is intended to take none-values (and thus whether the interface should allow them)

## 3 Proposed Implementation

- Add the following to [`flyteidl/protos/flyteidl/core/types.proto`](https://github.com/flyteorg/flyteidl/blob/master/protos/flyteidl/core/types.proto):
    ```proto
    message SumType {
        repeated LiteralType summands = 1;
    }
    // ...
    message LiteralType {
        oneof type {
            // ...
            SumType sum = 8;
        }
        // ...
    }
    ```
- Add the following to [`flyteidl/protos/flyteidl/core/literals.proto`](https://github.com/flyteorg/flyteidl/blob/master/protos/flyteidl/core/literals.proto):
    ```proto
    message LiteralSum {
        Literal value = 1;
        SumType type = 2;
        uint64 summand_idx = 3;
    }
    // ...
    message Scalar {
        oneof value {
            // ...
            LiteralSum sum = 8;
        }
    }
    ```
- Implement a new type checker in [`flytepropeller/pkg/compiler/validators/typing.go`](https://github.com/flyteorg/flytepropeller/blob/master/pkg/compiler/validators/typing.go):
    ```go
    func (t sumChecker) CastsFrom(upstreamType *flyte.LiteralType) bool {
        for _, x := range t.literalType.GetSum().GetSummands() {
            if getTypeChecker(x).CastsFrom(upstreamType) {
                return true;
            }
        }
        return false;
    }
    ```
- Do not implicitly accept none values for other types (potentially breaking change):
    - Do not accept other types as Void downstream
        - https://github.com/flyteorg/flytepropeller/blob/master/pkg/compiler/validators/typing.go#L59
    - Do not accept Void as other types downstream
        - In "trivial" types https://github.com/flyteorg/flytepropeller/blob/master/pkg/compiler/validators/typing.go#L33
        - In maps https://github.com/flyteorg/flytepropeller/blob/master/pkg/compiler/validators/typing.go#L66
        - In collections https://github.com/flyteorg/flytepropeller/blob/master/pkg/compiler/validators/typing.go#L80
        - In schemas https://github.com/flyteorg/flytepropeller/blob/master/pkg/compiler/validators/typing.go#L101
- Update the bindings type validation code in [`flytepropeller/pkg/compiler/validators/bindings.go`](https://github.com/flyteorg/flytepropeller/blob/master/pkg/compiler/validators/bindings.go#L14):
    ```go
    func validateBinding(w c.WorkflowBuilder, nodeID c.NodeID, nodeParam string, binding *flyte.BindingData,
	    expectedType *flyte.LiteralType, errs errors.CompileErrors) (
        resolvedType *flyte.LiteralType, upstreamNodes []c.NodeID, ok bool) {

        switch binding.GetValue().(type) {
        case *flyte.BindingData_Scalar:
            // Goes through SumType-aware AreTypesCastable
            break
        case *flyte.BindingData_Promise:
            // Goes through SumType-aware AreTypesCastable
            break
        default:
            if expectedType.GetSum() != nil {
                for _, t := range expectedType.GetSum().GetSummands() {
                    if resolvedType, nodeIds, ok := validateBinding(w, nodeID, nodeParam, binding, t, errors.NewCompileErrors()); ok {
                        // there can be no errors otherwise ok = false
                        return resolvedType, nodeIds, ok
                    }
                }
                errs.Collect(errors.NewMismatchingBindingsErr(nodeID, nodeParam, expectedType.String(), binding.GetCollection().String()))
                return nil, nil, !errs.HasErrors()
            }
        }
        // ...
    }
    ```
    - TODO: It might be necessary to accumulate the errors for each of the summands' failed binding validations to ease debugging. If that is the case, it would be preferable to ignore errors by default and re-run the verification if no candidate was found to avoid slowing down the non-exceptional case
        - The verbosity of the resulting messages would make it very hard to read so only a broad error is collected right now. It is unclear whether the extra complexity in the code and in the output is justified
- Implement a `typing.Union` type transformer in Python FlyteKit:
    - `get_literal_type`:
        ```py
        return LiteralType(sum=_type_models.SumType(summands=[TypeEngine.to_literal_type(x) for x in t.__args__]))
        ```
    - `to_literal`
        - Iterate through the types in `python_type.__args__` and try `TypeEngine.to_literal` for each. The first succeeding type is accepted
        - TODO: this might mean that order of summands matters e.g. `X | Y` is different from `Y | X`
    - `to_python_value`
        - Use the `TypeTransformer` for the `lv.sum.type` to transform `lv.sum.value`
    - `guess_python_type`
        - Return `TypeEngine.guess_python_type(lv.sum.type)`
- All `TypeTransformer`s' `to_literal` must be updated to fail with a specific error class so the `typing.Union` transformer can distinguish between user or programmer error and actual failure to convert type
- Update [`flytekit/core/interface.py`](https://github.com/flyteorg/flytekit/blob/master/flytekit/core/interface.py) to support `None` values as parameter defaults
    - Check whether the default is present by comparing with `inspect.Parameter.empty` in [`transform_inputs_to_parameters`](https://github.com/flyteorg/flytekit/blob/master/flytekit/core/interface.py#L186)
    - Pass `inspect.Parameter.empty` to the interface as is in [`transform_signature_to_interface`](https://github.com/flyteorg/flytekit/blob/master/flytekit/core/interface.py#L283)

## 4 Metrics & Dashboards

None

## 5 Drawbacks

- Projects relying on types being implicitly nullable will be broken by this update since parameter types and return types might need to be changed to optionals
    - A feature flag can be used to ease the transition
    - It would be nice to estimate the number of projects affected by this

## 6 Alternatives

TODO: discuss the type-erased version

## 7 Potential Impact and Dependencies

See drawbacks

## 8 Unresolved questions

TODO: discuss the type-erased version

## 9 Conclusion

TODO

## 10 RFC Process Guide, remove this section when done

**Checklist:**

- [x]  Copy template
- [x]  Draft RFC (think of it as a wireframe)
- [ ]  Share as WIP with folks you trust to gut-check
- [ ]  Send pull request when comfortable
- [ ]  Label accordingly
- [ ]  Assign reviewers
- [ ]  Merge PR