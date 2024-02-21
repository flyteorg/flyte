from typing import Callable, Dict, Optional

from docstring_parser import parse


class Docstring(object):
    def __init__(self, docstring: Optional[str] = None, callable_: Optional[Callable] = None):
        if docstring is not None:
            self._parsed_docstring = parse(docstring)
        else:
            self._parsed_docstring = parse(callable_.__doc__)

    @property
    def input_descriptions(self) -> Dict[str, str]:
        return {p.arg_name: p.description for p in self._parsed_docstring.params}

    @property
    def output_descriptions(self) -> Dict[str, str]:
        return {p.return_name: p.description for p in self._parsed_docstring.many_returns}

    @property
    def short_description(self) -> Optional[str]:
        return self._parsed_docstring.short_description

    @property
    def long_description(self) -> Optional[str]:
        return self._parsed_docstring.long_description
