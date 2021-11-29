# [RR] Typing Metadata

**Authors:**

- @maximsmol
- @kennyworkman

## 1 Executive Summary

A small patch to introduce flytekit support for python-native type metadata.

## 2 Motivation

Users of flyte wish to have access to arbitrary parameter metadata specified in
typing annotations.

We propose adding a new message called `TypeAnnotation` to hold arbitrary annotations
associated with parameter types.

We wish to add python-native support for this field by allowing flytekit to
recognize and parse the use of
[`typing.Annotation`](https://docs.python.org/3/library/typing.html#typing.Annotated).

This could be to change the presentation of a form input ingesting typed values,
surface descriptions to users or store indexes to control the order of
parameters with custom frontends.

This feature will not send metadata to the container at runtime when a task is
invoked.  If the container has access to the original code, it will be able to
access/load metadata from that. This is particularly important to note for
TaskTemplate-based flytekit plugins.

An example of how we would use this annotation to tag a parameter with
presentation information:


```
amplicon_min_alignment_score: Optional[
        List[
            Annotated[
                int,
                FlyteMetadata(
                    {
                        "appearance": {
                            "multiselect": {
                                "options": [50, 60, 70, 80, 90, 100],
                                "allow_custom": True,
                            }
                        }
                    }
                ),
            ]
        ]
    ],
```

## 3 Proposed Implementation

The addition of an additional field to `LiteralType` in flyteidl and an additional
message to represent a `TypeAnnotation` object.

```
message TypeAnnotation {
  google.protobuf.Struct annotations = 1;
  // string type_hint = 2;
  // google.protobuf.Struct json_schema = 2;
 ...
}

message LiteralType 
{
...
   TypeAnnotation annotation = 7;
}
```

Expansion of the `to_literal_type` method in the core [`TypeEngine`
class](https://github.com/flyteorg/flytekit/blob/master/flytekit/core/type_engine.py#L341).

```
    @classmethod
    def to_literal_type(cls, python_type: Type) -> LiteralType:
        """
        Converts a python type into a flyte specific ``LiteralType``
        """
        transformer = cls.get_transformer(python_type)
        res = transformer.get_literal_type(python_type)
        meta = None
        if hasattr(python_type, "__metadata__"):
            for x in python_type.__metadata__:
                if not isinstance(x, FlyteTypeAnnotation):
                    continue
                if x.data.get("__consumed", False):
                    continue
                annotation = x.data
                x.data["__consumed"] = True

        if annotation is not None:
            return LiteralType(
                simple=res.simple,
                schema=res.schema,
                collection_type=res.collection_type,
                map_value_type=res.map_value_type,
                blob=res.blob,
                enum_type=res.enum_type,
                annotation=annotation,
            )
        return res
```

coupled with the introduction of a custom object to represent parsed metadata:

```
class FlyteTypeAnnotation:
    def __init__(self, data: Dict[str, Any]):
        self._data = data

    @property
    def data(self):
        return self._data
```

## 4 Metrics & Dashboards

Not convinced if relevant here.

## 5 Drawbacks

Do not see any.

## 6 Alternatives

Do not see any good ones.

## 7 Potential Impact and Dependencies

Do not see any.

## 8 Unresolved questions

Not sure.

## 9 Conclusion

The people need typing metadata to build rich applications on top of flyte ;)
