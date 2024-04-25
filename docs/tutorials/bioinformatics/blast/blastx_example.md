---
jupytext:
  cell_metadata_filter: all
  formats: md:myst
  main_language: python
  notebook_metadata_filter: all
  text_representation:
    extension: .md
    format_name: myst
    format_version: 0.13
    jupytext_version: 1.16.1
kernelspec:
  display_name: Python 3
  language: python
  name: python3
---

# BLASTX Example

This demonstration will utilize BLASTX to search for a nucleotide sequence within a local protein database.

+++ {"lines_to_next_cell": 0}

Import the necessary libraries.

```{code-cell}
from pathlib import Path
from typing import NamedTuple

import matplotlib.pyplot as plt
import pandas as pd
import requests
from flytekit import conditional, kwtypes, task, workflow
from flytekit.extras.tasks.shell import OutputLocation, ShellTask
from flytekit.types.file import FlyteFile, PNGImageFile
```

+++ {"lines_to_next_cell": 0}

Download the data from GitHub.

:::{note}
When running code on the demo cluster, make sure data is included in the Docker image.
Uncomment copy data command in the Dockerfile.
:::

```{code-cell}
def download_dataset():
    Path("kitasatospora").mkdir(exist_ok=True)
    r = requests.get("https://api.github.com/repos/flyteorg/flytesnacks/contents/blast/kitasatospora?ref=datasets")
    for each_file in r.json():
        download_url = each_file["download_url"]
        file_name = f"kitasatospora/{Path(download_url).name}"
        if not Path(file_name).exists():
            r_file = requests.get(each_file["download_url"])
            open(file_name, "wb").write(r_file.content)
```

+++ {"lines_to_next_cell": 0}

A `ShellTask` allows you to run commands on the shell.
In this example, we use a `ShellTask` to create and execute the BLASTX command.

Start by specifying the location of the BLAST output file.
Then, define variables that hold the paths to the input query sequence file, the database we are searching against, and the output file for BLAST.
Finally, generate and run the BLASTX command.
Both the standard output (stdout) and standard error (stderr) are captured and saved in the `stdout` variable.

The `{inputs}` and `{outputs}` are placeholders for input and output values, respectively.

```{code-cell}
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
```

:::{note}
The `outfmt=6` option requests BLASTX to generate a tab-separated plain text file, which is convenient for automated processing.
:::

If the command runs successfully, there should be no standard output or error (stdout and stderr should be empty).

+++ {"lines_to_next_cell": 0}

Next, define a task to load the BLASTX output.
The task returns a pandas DataFrame and a plot.
The file containing the BLASTX results is referred to as blastout.

```{code-cell}
BLASTXOutput = NamedTuple("blastx_output", result=pd.DataFrame, plot=PNGImageFile)


@task
def blastx_output(blastout: FlyteFile) -> BLASTXOutput:
    # read BLASTX output
    result = pd.read_csv(blastout, sep="\t", header=None)

    # define column headers
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

    # assign headers
    result.columns = headers

    # create a scatterplot
    result.plot.scatter("pc_identity", "e_value")
    plt.title("E value vs %identity")
    plot = "plot.png"
    plt.savefig(plot)

    return BLASTXOutput(result=result.head(), plot=plot)
```

+++ {"lines_to_next_cell": 0}

Verify that the BLASTX run was successful by checking if the standard output and error are empty.

```{code-cell}
@task
def is_batchx_success(stdout: FlyteFile) -> bool:
    if open(stdout).read():
        return False
    else:
        return True
```

+++ {"lines_to_next_cell": 0}

Create a workflow that calls the previously defined tasks.
A {ref}`conditional <conditional>` statement is used to check the success of the BLASTX command.

```{code-cell}
@workflow
def blast_wf(
    datadir: str = "kitasatospora",
    outdir: str = "output",
    query: str = "k_sp_CB01950_penicillin.fasta",
    db: str = "kitasatospora_proteins.faa",
    blast_output: str = "AMK19_00175_blastx_kitasatospora.tab",
) -> BLASTXOutput:
    stdout, blastout = blastx_on_shell(datadir=datadir, outdir=outdir, query=query, db=db, blast_output=blast_output)
    result = is_batchx_success(stdout=stdout)
    final_result, plot = (
        conditional("blastx_output")
        .if_(result.is_true())
        .then(blastx_output(blastout=blastout))
        .else_()
        .fail("BLASTX failed")
    )
    return BLASTXOutput(result=final_result, plot=plot)
```

+++ {"lines_to_next_cell": 0}

Run the workflow locally.

```{code-cell}
if __name__ == "__main__":
    print("Downloading dataset...")
    download_dataset()
    print("Running BLASTX...")
    print(f"BLASTX result: {blast_wf()}")
```
