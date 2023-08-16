(blast)=

# Nucleotide Sequence Querying with BLASTX

```{tags} Advanced
```

This tutorial demonstrates the integration of computational biology and Flyte.
The focus will be on searching a nucleotide sequence against a local protein database to identify possible homologues.
The steps include:

- Data loading
- Creation of a {ref}`ShellTask <shell_task>` to execute the BLASTX search command
- Loading of BLASTX results and plotting a graph of "e-value" vs "pc identity"

This tutorial is based on the reference guide ["Using BLAST+ Programmatically with Biopython"](https://widdowquinn.github.io/2018-03-06-ibioic/02-sequence_databases/03-programming_for_blast.html).

## BLAST

The Basic Local Alignment Search Tool (BLAST) is a program that identifies similar regions between sequences.
It compares nucleotide or protein sequences with sequence databases and evaluates the statistical significance of the matches.
BLAST can be used to deduce functional and evolutionary relationships between sequences and identify members of gene families.

For additional information, visit the [BLAST Homepage](https://blast.ncbi.nlm.nih.gov/Blast.cgi).

### BLASTX

BLASTx is a useful tool for searching genes and predicting their functions or relationships with other gene sequences.
It is commonly employed to find protein-coding genes in genomic DNA or cDNA, as well as to determine whether a new nucleotide sequence encodes a protein or to identify proteins encoded by transcripts or transcript variants.

This tutorial will demonstrate how to perform a BLASTx search.

## Data

The database used in this example consists of predicted gene products from five Kitasatospora genomes.
The query is a single nucleotide sequence of a predicted penicillin-binding protein from Kitasatospora sp. CB01950.

```{note}
To run the example locally, you need to download BLAST.
You can find OS-specific installation instructions in the [user manual](https://www.ncbi.nlm.nih.gov/books/NBK569861/).
```

## Dockerfile

```{literalinclude} ../../../examples/blast/Dockerfile
:emphasize-lines: 42-44,67-70
:language: docker
```

Initiate the workflow on the Flyte backend by executing the following two commands in the "bioinformatics" directory:

```
pyflyte --pkgs blast package --image ghcr.io/flyteorg/flytecookbook:blast-latest
flytectl register files --project flytesnacks --domain development --archive flyte-package.tgz --version v1
```

## Examples

```{auto-examples-toc}
blastx_example
```
