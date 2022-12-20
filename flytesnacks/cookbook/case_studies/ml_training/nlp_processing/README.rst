NLP Processing
--------------

.. tags:: MachineLearning, UI, Intermediate

This tutorial will demonstrate how to process text data and generate word embeddings and visualizations
as part of a Flyte workflow. It's an adaptation of the official Gensim `Word2Vec tutorial <https://radimrehurek.com/gensim/auto_examples/tutorials/run_word2vec.html>`__.


About Gensim
============

Gensim is a popular open-source natural language processing (NLP) library used to process
large corpora (can be larger than RAM).
It has efficient multicore implementations of a number of algorithms such as `Latent Semantic Analysis <http://lsa.colorado.edu/papers/dp1.LSAintro.pdf>`__, `Latent Dirichlet Allocation (LDA) <https://www.jmlr.org/papers/volume3/blei03a/blei03a.pdf>`__,
`Word2Vec deep learning <https://arxiv.org/pdf/1301.3781.pdf>`__ to perform complex tasks including understanding
document relationships, topic modeling, learning word embeddings, and more.

You can read more about Gensim `here <https://radimrehurek.com/gensim/>`__.


Data
====

The dataset used for this tutorial is the open-source `Lee Background Corpus <https://github.com/RaRe-Technologies/gensim/blob/develop/gensim/test/test_data/lee_background.cor>`__
that comes with the Gensim library.


Step-by-Step Process
====================

The following points outline the modelling process:

- Returns a preprocessed (tokenized, stop words excluded, lemmatized) corpus from the custom iterator.
- Trains the Word2vec model on the preprocessed corpus.
- Generates a bag of words from the corpus and trains the LDA model.
- Saves the LDA and Word2Vec models to disk.
- Deserializes the Word2Vec model, runs word similarity and computes word movers distance.
- Reduces the dimensionality (using tsne) and plots the word embeddings.

Let's dive into the code!
