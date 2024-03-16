import typing

from flytekit import Workflow, task


@task
def t1(a: str) -> str:
    return a + " world"


@task
def t2():
    print("side effect")


@task
def t3(a: typing.List[str]) -> str:
    return ",".join(a)


wf = Workflow(name="my.imperative.workflow.example")
wf.add_workflow_input("in1", str)
node_t1 = wf.add_entity(t1, a=wf.inputs["in1"])
wf.add_workflow_output("output_from_t1", node_t1.outputs["o0"])
wf.add_entity(t2)

wf_in2 = wf.add_workflow_input("in2", str)
node_t3 = wf.add_entity(t3, a=[wf.inputs["in1"], wf_in2])

wf.add_workflow_output(
    "output_list",
    [node_t1.outputs["o0"], node_t3.outputs["o0"]],
    python_type=typing.List[str],
)


if __name__ == "__main__":
    print(wf)
    print(wf(in1="hello", in2="foo"))
