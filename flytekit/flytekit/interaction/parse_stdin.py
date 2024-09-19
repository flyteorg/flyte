from __future__ import annotations

import typing

import click

from flytekit.core.context_manager import FlyteContext
from flytekit.core.type_engine import TypeEngine
from flytekit.models.literals import Literal
from flytekit.models.types import LiteralType


def parse_stdin_to_literal(
    ctx: FlyteContext, t: typing.Type, message: typing.Optional[str], lt: typing.Optional[LiteralType] = None
) -> Literal:
    """
    Parses the user input from stdin and converts it to a literal of the given type.
    """
    from flytekit.interaction.click_types import FlyteLiteralConverter

    if not lt:
        lt = TypeEngine.to_literal_type(t)
    literal_converter = FlyteLiteralConverter(
        ctx,
        literal_type=lt,
        python_type=t,
        is_remote=False,
    )
    user_input = click.prompt(message, type=literal_converter.click_type)
    try:
        option = click.Option(["--input"], type=literal_converter.click_type)
        v = literal_converter.click_type.convert(user_input, option, click.Context(command=click.Command("")))
        return TypeEngine.to_literal(FlyteContext.current_context(), v, t, lt)
    except Exception as e:
        raise click.ClickException(f"Failed to parse input: {e}")
