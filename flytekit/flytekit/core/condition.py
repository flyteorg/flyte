from __future__ import annotations

import datetime
import typing
from typing import Optional, Tuple, Union, cast

from flytekit.core.context_manager import FlyteContextManager
from flytekit.core.node import Node
from flytekit.core.promise import (
    ComparisonExpression,
    ComparisonOps,
    ConjunctionExpression,
    ConjunctionOps,
    NodeOutput,
    Promise,
    VoidPromise,
    create_task_output,
)
from flytekit.models.core import condition as _core_cond
from flytekit.models.core import workflow as _core_wf
from flytekit.models.literals import Binding, BindingData, Literal, RetryStrategy
from flytekit.models.types import Error


class BranchNode(object):
    def __init__(self, name: str, ifelse_block: _core_wf.IfElseBlock):
        self._name = name
        self._ifelse_block = ifelse_block

    @property
    def name(self):
        return self._name


class ConditionalSection:
    """
    ConditionalSection is used to denote a condition within a Workflow. This default conditional section only works
    for Compilation mode. It is advised to derive the class and re-implement the `start_branch` and `end_branch` methods
    to override the compilation behavior

    .. note::

        Conditions can only be used within a workflow context.

    Usage:

    .. code-block:: python

        v =  conditional("fractions").if_((my_input > 0.1) & (my_input < 1.0)).then(...)...

    """

    def __init__(self, name: str):
        self._name = name
        self._cases: typing.List[Case] = []
        self._last_case = False
        self._condition = Condition(self)
        ctx = FlyteContextManager.current_context()
        # A new conditional section has been started, so lets push the context
        FlyteContextManager.push_context(ctx.enter_conditional_section().build())

    @property
    def name(self) -> str:
        return self._name

    @property
    def cases(self) -> typing.List[Case]:
        return self._cases

    def start_branch(self, c: Case, last_case: bool = False) -> Case:
        """
        At the start of an execution of every branch this method should be called.
        :param c: -> the case that represents this branch
        :param last_case: -> a boolean that indicates if this is the last branch in the ifelseblock
        """
        self._last_case = last_case
        self._cases.append(c)
        return self._cases[-1]

    def end_branch(self) -> Optional[Union[Condition, Promise, Tuple[Promise], VoidPromise]]:
        """
        This should be invoked after every branch has been visited.
        In case this is not local workflow execution then, we should check if this is the last case.
        If so then return the promise, else return the condition
        """
        if self._last_case:
            # We have completed the conditional section, lets pop off the branch context
            FlyteContextManager.pop_context()
            ctx = FlyteContextManager.current_context()
            # Question: This is commented out because we don't need it? Nodes created in the conditional
            #   compilation state are captured in the to_case_block? Always?
            #   Is this still true of nested conditionals? Is that why propeller compiler is complaining?
            # branch_nodes = ctx.compilation_state.nodes
            node, promises = to_branch_node(self._name, self)
            # Verify branch_nodes == nodes in bn
            bindings: typing.List[Binding] = []
            upstream_nodes = set()
            for p in promises:
                if not p.is_ready:
                    bindings.append(Binding(var=p.var, binding=BindingData(promise=p.ref)))
                    upstream_nodes.add(p.ref.node)

            n = Node(
                id=f"{ctx.compilation_state.prefix}n{len(ctx.compilation_state.nodes)}",  # type: ignore
                metadata=_core_wf.NodeMetadata(self._name, timeout=datetime.timedelta(), retries=RetryStrategy(0)),
                bindings=sorted(bindings, key=lambda b: b.var),
                upstream_nodes=list(upstream_nodes),  # type: ignore
                flyte_entity=node,
            )
            FlyteContextManager.current_context().compilation_state.add_node(n)  # type: ignore
            return self._compute_outputs(n)
        return self._condition

    def if_(self, expr: Union[ComparisonExpression, ConjunctionExpression]) -> Case:
        return self._condition._if(expr)

    def compute_output_vars(self) -> typing.Optional[typing.List[str]]:
        """
        Computes and returns the minimum set of outputs for this conditional block, based on all the cases that have
        been registered
        """
        output_vars: typing.List[str] = []
        output_vars_set = set()
        for c in self._cases:
            if c.output_promise is None and c.err is None:
                # One node returns a void output and no error, we will default to None return
                return None
            if isinstance(c.output_promise, VoidPromise):
                return None
            if c.output_promise is not None:
                var = []
                if isinstance(c.output_promise, tuple):
                    var = [i.var for i in c.output_promise]
                else:
                    var = [c.output_promise.var]
                curr_set = set(var)
                if not output_vars:
                    output_vars = var
                    output_vars_set = curr_set
                else:
                    output_vars_set = output_vars_set.intersection(curr_set)
                    new_output_var = []
                    for v in output_vars:
                        if v in output_vars_set:
                            new_output_var.append(v)
                    output_vars = new_output_var

        return output_vars

    def _compute_outputs(self, n: Node) -> Optional[Union[Promise, Tuple[Promise], VoidPromise]]:
        curr = self.compute_output_vars()
        if curr is None or len(curr) == 0:
            return VoidPromise(n.id, NodeOutput(node=n, var="placeholder"))
        promises = [Promise(var=x, val=NodeOutput(node=n, var=x)) for x in curr]
        # TODO: Is there a way to add the Python interface here? Currently, it's an optional arg.
        return create_task_output(promises)

    def __repr__(self):
        return self._condition.__repr__()

    def __str__(self):
        return self.__repr__()


class LocalExecutedConditionalSection(ConditionalSection):
    def __init__(self, name: str):
        self._selected_case: Optional[Case] = None
        super().__init__(name=name)

    def start_branch(self, c: Case, last_case: bool = False) -> Case:
        """
        At the start of an execution of every branch this method should be called.
        :param c: -> the case that represents this branch
        :param last_case: -> a boolean that indicates if this is the last branch in the ifelseblock
        """
        added_case = super().start_branch(c, last_case)
        ctx = FlyteContextManager.current_context()
        # This is a short-circuit for the case when the branch was taken
        # We already have a candidate case selected
        if self._selected_case is None:
            if c.expr is None or c.expr.eval() or last_case:
                ctx.execution_state.take_branch()  # type: ignore
                self._selected_case = added_case
        return added_case

    def end_branch(self) -> Union[Condition, Promise]:
        """
        This should be invoked after every branch has been visited
        In case of Local workflow execution, we should first mark the branch as complete, then
        Then we first check for if this is the last case,
        In case this is the last case, we return the output from the selected case - A case should always
        be selected (see start_branch)
        If this is not the last case, we should return the condition so that further chaining can be done
        """
        ctx = FlyteContextManager.current_context()
        # Let us mark the execution state as complete
        ctx.execution_state.branch_complete()  # type: ignore
        if self._last_case and self._selected_case:
            # We have completed the conditional section, lets pop off the branch context
            FlyteContextManager.pop_context()
            if self._selected_case.output_promise is None and self._selected_case.err is None:
                raise AssertionError("Bad conditional statements, did not resolve in a promise")
            elif self._selected_case.output_promise is not None:
                return typing.cast(Promise, self._compute_outputs(self._selected_case.output_promise))
            raise ValueError(self._selected_case.err)
        return self._condition

    def _compute_outputs(self, selected_output_promise) -> Optional[Union[Tuple[Promise], Promise, VoidPromise]]:
        """
        For the local execution case only returns the least common set of outputs
        """
        curr = self.compute_output_vars()
        if curr is None:
            return VoidPromise(self.name)
        if not isinstance(selected_output_promise, tuple):
            selected_output_promise = (selected_output_promise,)
        promises = [Promise(var=x, val=v.val) for x, v in zip(curr, selected_output_promise)]
        return create_task_output(promises)


class SkippedConditionalSection(ConditionalSection):
    """
    This ConditionalSection is used for nested conditionals, when the branch has been evaluated to false.
    This ensures that the branch is not evaluated and thus the local tasks are not executed.
    """

    def __init__(self, name: str):
        super().__init__(name=name)

    def end_branch(self) -> Optional[Union[Condition, Tuple[Promise], Promise, VoidPromise]]:
        """
        This should be invoked after every branch has been visited
        """
        if self._last_case:
            FlyteContextManager.pop_context()
            curr = self.compute_output_vars()
            if curr is None:
                return VoidPromise(self.name)
            promises = [Promise(var=x, val=None) for x in curr]
            return create_task_output(promises)
        return self._condition


class Case(object):
    def __init__(
        self,
        cs: ConditionalSection,
        expr: Optional[Union[ComparisonExpression, ConjunctionExpression]],
        stmt: str = "elif",
    ):
        self._cs = cs
        if expr is not None:
            if isinstance(expr, bool):
                raise AssertionError(
                    "Logical (and/or/is/not) operations are not supported. "
                    "Expressions Comparison (<,<=,>,>=,==,!=) or Conjunction (&/|) are supported."
                    f"Received an evaluated expression with val {expr} in {cs.name}.{stmt}"
                )
            if isinstance(expr, Promise):
                raise AssertionError(
                    "Flytekit does not support unary expressions of the form `if_(x) - where x is an"
                    " input value or output of a previous node."
                    f" Received var {expr} in condition {cs.name}.{stmt}"
                )
            if not (isinstance(expr, ConjunctionExpression) or isinstance(expr, ComparisonExpression)):
                raise AssertionError(
                    "Flytekit only supports Comparison (<,<=,>,>=,==,!=) or Conjunction (&/|) "
                    f"expressions, Received var {expr} in condition {cs.name}.{stmt}"
                )
        self._expr = expr
        self._output_promise: Optional[Union[Tuple[Promise], Promise]] = None
        self._err: Optional[str] = None
        self._stmt = stmt
        self._output_node = None

    @property
    def output_node(self) -> Optional[Node]:
        # This is supposed to hold a pointer to the node that created this case.
        # It is set in the then() call. but the value will not be set if it's a VoidPromise or None was returned.
        return self._output_node

    @property
    def expr(self) -> Optional[Union[ComparisonExpression, ConjunctionExpression]]:
        return self._expr

    @property
    def output_promise(self) -> Optional[Union[Tuple[Promise], Promise]]:
        return self._output_promise

    @property
    def err(self) -> Optional[str]:
        return self._err

    # TODO this is complicated. We do not want this to run
    def then(
        self, p: Union[Promise, Tuple[Promise]]
    ) -> Optional[Union[Condition, Promise, Tuple[Promise], VoidPromise]]:
        self._output_promise = p
        if isinstance(p, Promise):
            if not p.is_ready:
                self._output_node = p.ref.node  # type: ignore
        elif isinstance(p, VoidPromise):
            if p.ref is not None:
                self._output_node = p.ref.node
        elif hasattr(p, "_fields"):
            # This condition detects the NamedTuple case and iterates through the fields to find one that has a node
            # which should be the first one.
            for f in p._fields:  # type: ignore
                prom = getattr(p, f)
                if not prom.is_ready:
                    self._output_node = prom.ref.node
                    break

        # We can always mark branch as completed
        return self._cs.end_branch()

    def fail(self, err: str) -> Promise:
        self._err = err
        return cast(Promise, self._cs.end_branch())

    def __repr__(self):
        return f"{self._stmt}({self.expr.__repr__()})"

    def __str__(self):
        return self.__repr__()


class Condition(object):
    def __init__(self, cs: ConditionalSection):
        self._cs = cs

    def _if(self, expr: Union[ComparisonExpression, ConjunctionExpression]) -> Case:
        if expr is None:
            raise AssertionError(f"Required an expression received None for condition:{self._cs.name}.if_(...)")
        return self._cs.start_branch(Case(cs=self._cs, expr=expr, stmt="if_"))

    def elif_(self, expr: Union[ComparisonExpression, ConjunctionExpression]) -> Case:
        if expr is None:
            raise AssertionError(f"Required an expression received None for condition:{self._cs.name}.elif(...)")
        return self._cs.start_branch(Case(cs=self._cs, expr=expr, stmt="elif_"))

    def else_(self) -> Case:
        return self._cs.start_branch(Case(cs=self._cs, expr=None, stmt="else_"), last_case=True)

    def __repr__(self):
        return f"condition({self._cs.name}) " + "".join([x.__repr__() for x in self._cs.cases])

    def __str__(self):
        return self.__repr__()


_logical_ops = {
    ConjunctionOps.AND: _core_cond.ConjunctionExpression.LogicalOperator.AND,
    ConjunctionOps.OR: _core_cond.ConjunctionExpression.LogicalOperator.OR,
}
_comparators = {
    ComparisonOps.EQ: _core_cond.ComparisonExpression.Operator.EQ,
    ComparisonOps.NE: _core_cond.ComparisonExpression.Operator.NEQ,
    ComparisonOps.GT: _core_cond.ComparisonExpression.Operator.GT,
    ComparisonOps.GE: _core_cond.ComparisonExpression.Operator.GTE,
    ComparisonOps.LT: _core_cond.ComparisonExpression.Operator.LT,
    ComparisonOps.LE: _core_cond.ComparisonExpression.Operator.LTE,
}


def create_branch_node_promise_var(node_id: str, var: str) -> str:
    """
    Generates a globally (wf-level) unique id for a variable.

    When building bindings for the branch node, the inputs to the conditions (e.g. (x==5)) need to have variable names
    (e.g. x). Because it's currently infeasible to get the name (e.g. x), we resolve to using the referenced node's
    output name (e.g. o0, my_out,... etc.). In order to avoid naming collisions (in cases when, for example, the
    conditions reference two outputs of two different nodes named the same), we build a variable name composed of the
    referenced node name + '.' + the referenced output name. Ideally we use something like
    (https://github.com/pwwang/python-varname) to retrieve the assigned variable name (e.g. x). However, because of
    https://github.com/pwwang/python-varname/issues/28, this is not currently supported for all AST nodes types.

    :param str node_id: the original node_id that produced the variable.
    :param str var: the output variable name from the original node.
    :return: The generated unique id of the variable.
    """
    return f"{node_id}.{var}"


def merge_promises(*args: Optional[Promise]) -> typing.List[Promise]:
    node_vars: typing.Set[typing.Tuple[str, str]] = set()
    merged_promises: typing.List[Promise] = []
    for p in args:
        if p is not None and p.ref:
            node_var = (p.ref.node_id, p.ref.var)
            if node_var not in node_vars:
                new_p = p.with_var(create_branch_node_promise_var(p.ref.node_id, p.ref.var))
                merged_promises.append(new_p)
                node_vars.add(node_var)
    return merged_promises


def transform_to_conj_expr(
    expr: ConjunctionExpression,
) -> Tuple[_core_cond.ConjunctionExpression, typing.List[Promise]]:
    left, left_promises = transform_to_boolexpr(expr.lhs)
    right, right_promises = transform_to_boolexpr(expr.rhs)
    return (
        _core_cond.ConjunctionExpression(
            left_expression=left,
            right_expression=right,
            operator=_logical_ops[expr.op],
        ),
        merge_promises(*left_promises, *right_promises),
    )


def transform_to_operand(v: Union[Promise, Literal]) -> Tuple[_core_cond.Operand, Optional[Promise]]:
    if isinstance(v, Promise):
        return _core_cond.Operand(var=create_branch_node_promise_var(v.ref.node_id, v.var)), v
    if v.scalar.none_type:
        return _core_cond.Operand(scalar=v.scalar), None
    return _core_cond.Operand(primitive=v.scalar.primitive), None


def transform_to_comp_expr(expr: ComparisonExpression) -> Tuple[_core_cond.ComparisonExpression, typing.List[Promise]]:
    o_lhs, b_lhs = transform_to_operand(expr.lhs)
    o_rhs, b_rhs = transform_to_operand(expr.rhs)
    return (
        _core_cond.ComparisonExpression(left_value=o_lhs, right_value=o_rhs, operator=_comparators[expr.op]),
        merge_promises(b_lhs, b_rhs),
    )


def transform_to_boolexpr(
    expr: Union[ComparisonExpression, ConjunctionExpression]
) -> Tuple[_core_cond.BooleanExpression, typing.List[Promise]]:
    if isinstance(expr, ConjunctionExpression):
        cexpr, promises = transform_to_conj_expr(expr)
        return _core_cond.BooleanExpression(conjunction=cexpr), promises
    cexpr, promises = transform_to_comp_expr(expr)
    return _core_cond.BooleanExpression(comparison=cexpr), promises


def to_case_block(c: Case) -> Tuple[Union[_core_wf.IfBlock], typing.List[Promise]]:
    expr, promises = transform_to_boolexpr(cast(Union[ComparisonExpression, ConjunctionExpression], c.expr))
    if c.output_promise is not None:
        n = c.output_node
    return _core_wf.IfBlock(condition=expr, then_node=n), promises


def to_ifelse_block(node_id: str, cs: ConditionalSection) -> Tuple[_core_wf.IfElseBlock, typing.List[Binding]]:
    if len(cs.cases) == 0:
        raise AssertionError("Illegal Condition block, with no if-else cases")
    if len(cs.cases) < 2:
        raise AssertionError("At least an if/else is required. Dangling If is not allowed")
    all_promises: typing.List[Promise] = []
    first_case, promises = to_case_block(cs.cases[0])
    all_promises.extend(promises)
    other_cases: Optional[typing.List[_core_wf.IfBlock]] = None
    if len(cs.cases) > 2:
        other_cases = []
        for c in cs.cases[1:-1]:
            case, promises = to_case_block(c)
            all_promises.extend(promises)
            other_cases.append(case)
    last_case = cs.cases[-1]
    node = None
    err = None
    if last_case.output_promise is not None:
        node = last_case.output_node
    else:
        err = Error(failed_node_id=node_id, message=last_case.err if last_case.err else "Condition failed")
    return (
        _core_wf.IfElseBlock(case=first_case, other=other_cases, else_node=node, error=err),
        merge_promises(*all_promises),
    )


def to_branch_node(name: str, cs: ConditionalSection) -> Tuple[BranchNode, typing.List[Promise]]:
    blocks, promises = to_ifelse_block(name, cs)
    return BranchNode(name=name, ifelse_block=blocks), promises


def conditional(name: str) -> ConditionalSection:
    """
    Use a conditional section to control the flow of a workflow. Conditional sections can only be used inside a workflow
    context. Outside of a workflow they will raise an Assertion.

    The ``conditional`` method returns a new conditional section, that allows to create a - ternary operator like
    if-else clauses. The reason why it is called ternary-like is because, it returns the output of the branch result.
    So in-effect it is a functional style condition.

    Example of a condition usage. Note the nesting and the assignment to a LHS variable

    .. code-block:: python

         v = (
            conditional("fractions")
            .if_((my_input > 0.1) & (my_input < 1.0))
            .then(
                conditional("inner_fractions")
                .if_(my_input < 0.5)
                .then(double(n=my_input))
                .elif_((my_input > 0.5) & (my_input < 0.7))
                .then(square(n=my_input))
                .else_()
                .fail("Only <0.7 allowed")
            )
            .elif_((my_input > 1.0) & (my_input < 10.0))
            .then(square(n=my_input))
            .else_()
            .then(double(n=my_input))
        )
    """
    ctx = FlyteContextManager.current_context()

    if ctx.compilation_state:
        return ConditionalSection(name)
    elif ctx.execution_state:
        if ctx.execution_state.is_local_execution():
            # In case of Local workflow execution, we will actually evaluate the expression and based on the result
            # make the branch to be active using `take_branch` method
            from flytekit.core.context_manager import BranchEvalMode

            if ctx.execution_state.branch_eval_mode == BranchEvalMode.BRANCH_SKIPPED:
                return SkippedConditionalSection(name)
            return LocalExecutedConditionalSection(name)
    raise AssertionError("Branches can only be invoked within a workflow context!")
