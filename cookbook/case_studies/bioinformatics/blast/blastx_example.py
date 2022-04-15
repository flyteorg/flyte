"""
BLASTX Example
--------------

This example will use BLASTX to search for a nucleotide sequence against a local protein database.
"""

# %%
# First, we need to import some libraries.
from typing import NamedTuple

import matplotlib.pyplot as plt
import pandas as pd

from flytekit import conditional, kwtypes, task, workflow
from flytekit.extras.tasks.shell import OutputLocation, ShellTask
from flytekit.types.file import FlyteFile, PNGImageFile


# %%
# A ``ShellTask`` is useful to run commands on the shell.
# In this example, we use ``ShellTask`` to generate and run the BLASTX command.
#
# First, we define the location of the BLAST output file.
# Then we define variables that contain paths to: the input query sequence file, the database we are searching against, and the file containing the BLAST output.
# Finally, we generate and run the BLASTX command. Both ``stdout`` and ``stderr`` are captured and saved to the ``stdout`` variable.
#
# ``{inputs}`` and ``{outputs}`` are placeholders for the input and output values, respectively.
#
# .. note::
#    The new input/output placeholder syntax of ``ShellTask`` is available starting Flytekit 0.30.0b8+.
blastx_on_shell = ShellTask(
    name="blastx",
    debug=True,
    script="""
    mkdir -p {inputs.outdir}

    query={inputs.datadir}/{inputs.query}
    db={inputs.datadir}/{inputs.db}
    blastout={inputs.outdir}/{inputs.blast_output}

    blastx -out $blastout -outfmt 6 -query $query -db $db >> {outputs.stdout} 2>&1
    """,
    inputs=kwtypes(datadir=str, query=str, outdir=str, blast_output=str, db=str),
    output_locs=[
        OutputLocation(var="stdout", var_type=FlyteFile, location="stdout.txt"),
        OutputLocation(
            var="blastout",
            var_type=FlyteFile,
            location="{inputs.outdir}/{inputs.blast_output}",
        ),
    ],
)

# %%
# .. note::
#   ``outfmt=6`` asks BLASTX to write a tab-separated tabular plain text file.
#   This differs from the usual human-readable output, but is particularly convenient for automated processing.
#
# If the command works, then there should be no standard output and error, i.e., stdout and stderr have to be empty.

# %%
# Next, we define a task to load the BLASTX output. The task returns a pandas DataFrame and a plot.
# ``blastout`` pertains to the BLAST output file.

BLASTXOutput = NamedTuple("blastx_output", result=pd.DataFrame, plot=PNGImageFile)


@task
def blastx_output(blastout: FlyteFile) -> BLASTXOutput:
    # Read BLASTX output
    result = pd.read_csv(blastout, sep="\t", header=None)

    # Define column headers
    headers = [
        "query",
        "subject",
        "pc_identity",
        "aln_length",
        "mismatches",
        "gaps_opened",
        "query_start",
        "query_end",
        "subject_start",
        "subject_end",
        "e_value",
        "bitscore",
    ]

    # Assign headers
    result.columns = headers

    # Create a scatterplot
    result.plot.scatter("pc_identity", "e_value")
    plt.title("E value vs %identity")
    plot = "plot.png"
    plt.savefig(plot)

    return BLASTXOutput(result=result.head(), plot=plot)


# %%
# We write a task to ascertain if the BLASTX standard output and error are empty.
# If empty, then the BLASTX run was successful, else, the run failed.
@task
def is_batchx_success(stdout: FlyteFile) -> bool:
    if open(stdout).read():
        return False
    else:
        return True


# %%
# Next, we define a workflow to call the aforementioned tasks.
# We use :ref:`conditional <sphx_glr_auto_core_control_flow_conditions.py>` to check if the BLASTX command succeeded.
@workflow
def blast_wf(
    datadir: str = "data/kitasatospora",
    outdir: str = "output",
    query: str = "k_sp_CB01950_penicillin.fasta",
    db: str = "kitasatospora_proteins.faa",
    blast_output: str = "AMK19_00175_blastx_kitasatospora.tab",
) -> BLASTXOutput:
    stdout, blastout = blastx_on_shell(
        datadir=datadir, outdir=outdir, query=query, db=db, blast_output=blast_output
    )
    result = is_batchx_success(stdout=stdout)
    final_result, plot = (
        conditional("blastx_output")
        .if_(result.is_true())
        .then(blastx_output(blastout=blastout))
        .else_()
        .fail("BLASTX failed")
    )
    return BLASTXOutput(result=final_result, plot=plot)


# %%
# Finally, we can run the workflow locally.
if __name__ == "__main__":
    print("Running BLASTX...")
    print(f"BLASTX result: {blast_wf()}")
