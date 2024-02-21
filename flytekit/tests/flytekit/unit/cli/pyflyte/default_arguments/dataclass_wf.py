from dataclasses import dataclass

from dataclasses_json import DataClassJsonMixin

from flytekit import task, workflow


@dataclass
class DataclassA(DataClassJsonMixin):
    a: str
    b: int


@task
def t(dca: DataclassA):
    print(dca)


@workflow
def wf(dca: DataclassA = DataclassA("hello", 42)):
    t(dca=dca)
