# [RR] Parameter Docstring Metadata

**Authors:**

- @maximsmol
- @kennyworkman

## 1 Executive Summary

We propose to extend docstring parsing functionality from short/long description to parameter metadata.

## 2 Motivation

There already exists property on the `LiteralType` model for metadata.

The `LiteralType` definition in flyteidl already has a `metadata` field in its
message definition
(https://github.com/flyteorg/flyteidl/blob/master/protos/flyteidl/core/types.proto#L89),
suggesting that support for metadata annotations for flyte types was in the
works.

There also exists docstring parsing support within flytekit, but it is limited
to short and long descriptions only.

We wish to add python-native support for arbitrary parameter metadata specified in
common docstring formats (eg.  ReST, Google, Numpydoc-style, etc.)

This could be to change the presentation of a form input ingesting typed values,
surface descriptions to users or store indexes to control the order of
parameters with custom frontends.

eg.

```
default_min_aln_score:
      Sequences must have at least this homology percentage score with the amplicon to be aligned. Default: 60%.
      __metadata__:
        display_name: Minimum Homology % For Alignment to an Amplicon
        _tmp:
          hidden: true
          section_title: Amplicon and Read Matching
        appearance:
          multiselect:
            options:
              - 50
              - 60
              - 70
              - 80
              - 90
              - 100
            allow_custom: true
```

## 3 Proposed Implementation

Replace the `Docstring` object wholesale with more robust parsing utilites:
```
# flytekit/core/docstring.py
import re
from typing import Any, Dict, Union

import yaml
from docstring_parser import parse as _parse
from docstring_parser.common import Docstring


def parse_docstring(x: Union[str, Any]) -> Docstring:
    if isinstance(x, str):
        return _parse(x)

    if not hasattr(x, "__doc__"):
        raise ValueError(f"No docstring found: '{x}'")

    return _parse(x.__doc__)


_yaml_metadata_regex = re.compile(r"(.*)__metadata__:(.*)", re.MULTILINE | re.DOTALL)


def param_metadata_from_docstring(doc: Docstring) -> Dict[str, Dict[str, Any]]:
    res = {}
    for idx, x in enumerate(doc.params):
        meta = _yaml_metadata_regex.match(x.description)

        if meta is not None:
            try:
                res[x.arg_name] = {
                    "idx": idx,
                    "name": x.arg_name,
                    "default": x.default,
                    "description": meta.group(1),
                    "meta": yaml.safe_load(meta.group(2)),
                }
            except:
                pass

        if x.arg_name not in res:
            res[x.arg_name] = {"idx": idx, "name": x.arg_name, "default": x.default, "description": x.description}

        if idx == 0:
            if doc.long_description is not None:
                meta = _yaml_metadata_regex.match(doc.long_description)
                if meta is not None:
                    try:
                        res[x.arg_name]["__workflow_meta__"] = {
                            "short_description": doc.short_description,
                            "long_description": meta.group(1),
                            "meta": yaml.safe_load(meta.group(2)),
                        }
                    except:
                        pass

            if "__workflow_meta__" not in res[x.arg_name]:
                res[x.arg_name]["__workflow_meta__"] = {
                    "short_description": doc.short_description,
                    "long_description": doc.long_description,
                }

    return res


def returns_metadata_from_docstring(doc: Docstring) -> Dict[str, Dict[str, Any]]:
    res = {}
    for idx, x in enumerate(doc.many_returns):
        meta = _yaml_metadata_regex.match(x.description)
        if meta is not None:
            try:
                tmp = {
                    "idx": idx,
                    "name": x.return_name,
                    "description": meta.group(1),
                    "meta": yaml.safe_load(meta.group(2)),
                }
                res[x.return_name] = tmp
                continue
            except Exception as e:
                print(e)
                pass

        res[x.return_name] = {"idx": idx, "name": x.return_name, "description": x.description}

    return res
```

then update the type transfomer functions when constructing the interface:

```
# flytekit/core/interface.py L290
def transform_type(x: type, description: Optional[Any] = None) -> _interface_models.Variable:
    return _interface_models.Variable(type=TypeEngine.to_literal_type(x), description=json.dumps(description))

```

## 4 Metrics & Dashboards

n/a

## 5 Drawbacks

*Are there any reasons why we should not do this? Here we aim to evaluate risk and check ourselves.*

Adding an additional dependency to flytekit.

https://pypi.org/project/docstring-parser/

## 6 Alternatives

Writing a customer parser instead of using the library.

## 7 Potential Impact and Dependencies

https://pypi.org/project/docstring-parser/

## 8 Unresolved questions

Not sure.

## 9 Conclusion

The people need parameter metadata in-docstring to build richer applications on
top of flyte.
