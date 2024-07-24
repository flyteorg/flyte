from dataclasses import dataclass

from dataclasses_json import DataClassJsonMixin


@dataclass
class X(DataClassJsonMixin):
    x: int
