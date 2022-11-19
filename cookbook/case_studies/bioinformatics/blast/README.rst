.. _blast:

Nucleotide Sequence Querying with BLASTX
----------------------------------------

.. tags:: Advanced

This tutorial shows how computational biology intermixes with Flyte. The problem statement we will be looking at is
querying a nucleotide sequence against a local protein database, to identify potential homologues.
This guide will show you how to:

- Load the data
- Instantiate a :ref:`ShellTask <sphx_glr_auto_core_flyte_basics_shell_task.py>` to generate and run the BLASTX search command
- Load BLASTX results and plot a graph (``e_value`` vs. ``pc_identity``)

`Using BLAST+ Programmatically with Biopython <https://widdowquinn.github.io/2018-03-06-ibioic/02-sequence_databases/03-programming_for_blast.html>`__ has been used as a reference to construct the tutorial.

About BLAST
===========

The Basic Local Alignment Search Tool (BLAST) finds regions of local similarity between sequences.
The program compares nucleotide or protein sequences to sequence databases and calculates the statistical significance of matches.
BLAST can be used to infer functional and evolutionary relationships between sequences as well as help identify members of gene families.

You can read more about BLAST in the `BLAST Homepage <https://blast.ncbi.nlm.nih.gov/Blast.cgi>`__.

BLASTX
^^^^^^

BLASTx is a powerful tool to search for genes and predict their functions or relationships
with other gene sequences, and is typically used for identifying the protein‐coding genes in genomic DNA/cDNA.
It is also used to detect whether a novel nucleotide sequence is a protein‐coding gene or identify proteins encoded by transcripts or transcript variants.

In this tutorial, we will run a BLASTX search.

Data
====

The database comprises predicted gene products from five Kitasatospora genomes.
The query is a single nucleotide sequence of a predicted penicillin-binding protein from Kitasatospora sp. CB01950.

To run the example, download the database from `Flytesnacks datasets <https://github.com/flyteorg/flytesnacks/tree/datasets/blast/kitasatospora>`__.

.. note::
    To run the example locally, download BLAST first.
    You can find OS-specific installation instructions in the `user manual <https://www.ncbi.nlm.nih.gov/books/NBK569861/>`__.
    This example uses BLAST 2.12.0 version.

Dockerfile
==========

.. literalinclude:: ../../../../../case_studies/bioinformatics/blast/Dockerfile
    :language: docker
    :emphasize-lines: 40-47
