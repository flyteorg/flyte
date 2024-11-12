import requests

from docutils import nodes
from docutils.parsers.rst import directives
from six import text_type
from typing import Dict, Any, List, Tuple

from sphinx.locale import __
from sphinx.util import parselinenos
from sphinx.util.docutils import SphinxDirective
from sphinx.config import Config
from sphinx.util.nodes import set_source_info
from sphinx.directives.code import logger
from sphinx.directives.code import container_wrapper


class NotebookRemoteLiteralIncludeReader(object):

    def __init__(self, url: str, options: Dict[str, Any], config: Config) -> None:
        self.url = url
        self.options = options

    def read_file(self, url: str) -> List[str]:
        response = requests.get(url)
        json_data = response.json()

        if json_data:
            if not response.status_code == requests.codes.ok:
                raise ValueError(
                    "HTTP request returned error code %s" % response.status_code
                )
            cell_idx = int(self.options["cell"])
            if "cell" in self.options:
                if len(json_data["cells"]) > cell_idx:
                    if json_data["cells"][cell_idx]["cell_type"] == "code":
                        lines = json_data["cells"][cell_idx]["source"]
                        text = "".join(lines)
                    else:
                        raise ValueError("Cell is not a code cell")
                else:
                    raise ValueError("Cell exceeds the number of cells in the notebook")
            else:
                raise ValueError("Cell not specified in options")

            if "tab-width" in self.options:
                text = text.expandtabs(self.options["tab-width"])

            return text.splitlines(True)
        else:
            raise IOError(__("Include file %r not found or reading it failed") % url)

    def read(self) -> Tuple[str, int]:
        lines = self.read_file(self.url)
        lines = self.lines_filter(lines)

        return "".join(lines), len(lines)

    def lines_filter(self, lines: List[str]) -> List[str]:
        linespec = self.options.get("lines")
        if linespec:
            linelist = parselinenos(linespec, len(lines))
            if any(i >= len(lines) for i in linelist):
                raise ValueError(
                    "Line number spec is out of range (1 - %s)" % len(lines)
                )

            lines = [lines[n] for n in linelist if n < len(lines)]
            if lines == []:
                raise ValueError(
                    __("Line spec %r: no lines pulled from include file %r")
                    % (linespec, self.url)
                )

        return lines


class NotebookRemoteLiteralInclude(SphinxDirective):
    """
    Like ``.. include:: :literal:``, but only warns if the include file is
    not found, and does not raise errors.  Also has several options for
    selecting what to include.
    """

    has_content = False
    required_arguments = 1
    optional_arguments = 0
    final_argument_whitespace = True
    option_spec = {
        "tab-width": int,
        "language": directives.unchanged_required,
        "lines": directives.unchanged_required,
        "cell": directives.unchanged_required,
        "emphasize-lines": directives.unchanged_required,
        "caption": directives.unchanged,
        "class": directives.class_option,
        "name": directives.unchanged,
    }

    def run(self) -> List[nodes.Node]:
        document = self.state.document
        if not document.settings.file_insertion_enabled:
            return [
                document.reporter.warning("File insertion disabled", line=self.lineno)
            ]
        try:
            url = self.arguments[0]

            reader = NotebookRemoteLiteralIncludeReader(url, self.options, self.config)
            text, lines = reader.read()

            retnode = nodes.literal_block(text, text, source=url)
            set_source_info(self, retnode)
            if "language" in self.options:
                retnode["language"] = self.options["language"]
            retnode["classes"] += self.options.get("class", [])
            extra_args = retnode["highlight_args"] = {}
            if "emphasize-lines" in self.options:
                hl_lines = parselinenos(self.options["emphasize-lines"], lines)
                if any(i >= lines for i in hl_lines):
                    logger.warning(
                        __("line number spec is out of range(1-%d): %r")
                        % (lines, self.options["emphasize-lines"])
                    )
                extra_args["hl_lines"] = [x + 1 for x in hl_lines if x < lines]

            if "caption" in self.options:
                caption = self.options["caption"] or self.arguments[0]
                retnode = container_wrapper(self, retnode, caption)

            # retnode will be note_implicit_target that is linked from caption and numref.
            # when options['name'] is provided, it should be primary ID.
            self.add_name(retnode)

            return [retnode]
        except Exception as exc:
            return [document.reporter.warning(text_type(exc), line=self.lineno)]


def setup(app):
    directives.register_directive("nb-rli", NotebookRemoteLiteralInclude)

    return {
        "parallel_read_safe": True,
        "parallel_write_safe": False,
    }
