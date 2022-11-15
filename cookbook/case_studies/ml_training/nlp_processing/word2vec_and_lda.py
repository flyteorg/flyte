"""
.. _word2vec_and_lda:

Word Embeddings and Topic Modelling with Gensim
-----------------------------------------------

This example creates six Flyte tasks that:

1. Generate the sample dataset.
2. Train the word2vec model.
3. Train the LDA model and display the words per topic.
4. Compute word similarities.
5. Compute word movers distance.
6. Reduce dimensions using t-SNE and generate a plot using FlyteDeck.

"""

# %%
# First, we import the necessary libraries.
import logging
import os
import random
import typing
from dataclasses import dataclass
from typing import Dict, List

import flytekit
import gensim
import nltk
import numpy as np
import plotly.graph_objects as go
import plotly.io as io
from dataclasses_json import dataclass_json
from flytekit import Resources, task, workflow
from flytekit.types.file import FlyteFile
from gensim import utils
from gensim.corpora import Dictionary
from gensim.models import LdaModel, Word2Vec
from gensim.parsing.preprocessing import STOPWORDS, remove_stopwords
from gensim.test.utils import datapath
from nltk.stem import WordNetLemmatizer
from nltk.tokenize import RegexpTokenizer
from sklearn.manifold import TSNE

logger = logging.getLogger(__file__)


# %%
# We define the output file type.
MODELSER_NLP = typing.TypeVar("model")
model_file = typing.NamedTuple("ModelFile", model=FlyteFile[MODELSER_NLP])

# %%
# Next, we define the path to the lee corpus dataset (installed with gensim).
data_dir = os.path.join(gensim.__path__[0], "test", "test_data")
lee_train_file = os.path.join(data_dir, "lee_background.cor")


# %%
# We declare ``NamedTuple``\s which will be used as signatures of the Flyte task outputs.
# The variable names and types correspond to the values of the unpacked tuples returned
# from the corresponding Flyte task.
plotdata = typing.NamedTuple(
    "PlottingData",
    x_values=List[float],
    y_values=List[float],
    labels=List[str],
)


workflow_outputs = typing.NamedTuple(
    "WorkflowOutputs",
    simwords=Dict[str, float],
    distance=float,
    topics=Dict[int, List[str]],
)


# %%
# We sample sentences of similar contexts to compare using the trained model.
SENTENCE_A = "Australian cricket captain has supported fast bowler"
SENTENCE_B = "Fast bowler received support from cricket captain"


# %%
# Data Generation
# ===============
#
# The data pre-processor implements the following steps:
#
# 1. Turns all words to lowercase and remove stopwords.
# 2. Splits the document into tokens using a regular expression tokenizer from NLTK.
# 3. Removes numeric single-character tokens as they do not tend to be useful, and the dataset contains a lot of them.
# 4. Uses the WordNet lemmatizer from NLTK and returns a list of lemmatized tokens.
def pre_processing(line: str) -> List[str]:
    tokenizer = RegexpTokenizer(r"\w+")
    tokens = tokenizer.tokenize(remove_stopwords(line.lower()))
    lemmatizer = WordNetLemmatizer()
    return [lemmatizer.lemmatize(token) for token in tokens]


# %%
# Now, we implement an iterator that calls the ``pre_processing`` function on each input sentence from the corpus
# and yield the processed results.
class MyCorpus:
    """An iterator that yields sentences (lists of str)."""

    def __init__(self, path):
        self.corpus_path = datapath(path)

    def __iter__(self):
        for line in open(self.corpus_path):
            yield pre_processing(line)


# %%
# We define a Flyte task to generate the processed corpus containing a list of tokenized sentence lists.
@task
def generate_processed_corpus() -> List[List[str]]:
    # download the required packages from the nltk library
    nltk.download("wordnet")
    nltk.download("omw-1.4")
    sentences_train = MyCorpus(lee_train_file)
    train_corpus = list(sentences_train)
    return train_corpus


# %%
# Hyperparameters
# ===============
#
# Next, we create a dataclass comprising Word2Vec hyperparameters:
#
# - ``min_count``:  Prunes the dictionary and removes low-frequency words.
# - ``vector_size``: Number of dimensions (N) of the N-dimensional space that gensim Word2Vec maps the words onto.
#   Bigger size values require more training data but can lead to better (more accurate) models.
# - ``workers``: For training parallelization to speed up training.
# - ``compute_loss``: To toggle computation of loss while training the Word2Vec model.
@dataclass_json
@dataclass
class Word2VecModelHyperparams(object):
    """
    Hyperparameters that can be used while training the word2vec model.
    """

    vector_size: int = 200
    min_count: int = 1
    workers: int = 4
    compute_loss: bool = True


# %%
# LDA needs a similar dataclass:
#
# - ``num_topics``: The number of topics to be extracted from the training corpus.
# - ``alpha``: A-priori belief on document-topic distribution. In `auto` mode, the model learns this from the data.
# - ``passes``: Controls how often the model is trained on the entire corpus or number of epochs.
# - ``chunksize``:  Controls how many documents are processed at a time in the training algorithm. Increasing the
#   chunk size speeds up training, at least as long as the chunk of documents easily fits into memory.
# - ``update_every``: Number of documents to be iterated through for each update.
# - ``random_state``: Seed for reproducibility.
@dataclass_json
@dataclass
class LDAModelHyperparams(object):
    """
    Hyperparameters that can be used while training the LDA model.
    """

    num_topics: int = 5
    alpha: str = "auto"
    passes: int = 10
    chunksize: int = 100
    update_every: int = 1
    random_state: int = 100


# %%
# Training
# ========
#
# We initialize and train a Word2Vec model on the preprocessed corpus.
@task
def train_word2vec_model(
    training_data: List[List[str]], hyperparams: Word2VecModelHyperparams
) -> model_file:

    model = Word2Vec(
        training_data,
        min_count=hyperparams.min_count,
        workers=hyperparams.workers,
        vector_size=hyperparams.vector_size,
        compute_loss=hyperparams.compute_loss,
    )
    training_loss = model.get_latest_training_loss()
    logger.info(f"training loss: {training_loss}")
    out_path = os.path.join(
        flytekit.current_context().working_directory, "word2vec.model"
    )
    model.save(out_path)
    return (out_path,)


# %%
# Next, we transform the documents to a vectorized form and compute the frequency of each word to generate a bag of
# words corpus for the LDA model to train on. We also create a mapping from word IDs to words to send it as an input to
# the LDA model for training.
@task
def train_lda_model(
    corpus: List[List[str]], hyperparams: LDAModelHyperparams
) -> Dict[int, List[str]]:
    id2word = Dictionary(corpus)
    bow_corpus = [id2word.doc2bow(doc) for doc in corpus]
    id_words = [[(id2word[id], count) for id, count in line] for line in bow_corpus]
    logger.info(f"Sample of bag of words generated: {id_words[:2]}")
    lda = LdaModel(
        corpus=bow_corpus,
        id2word=id2word,
        num_topics=hyperparams.num_topics,
        alpha=hyperparams.alpha,
        passes=hyperparams.passes,
        chunksize=hyperparams.chunksize,
        update_every=hyperparams.update_every,
        random_state=hyperparams.random_state,
    )
    return dict(lda.show_topics(num_words=5))


# %%
# Word Similarities
# =================
#
# We deserialize the model from disk and compute the top 10 similar
# words to the given word in the corpus (we will use the word `computer` when running
# the workflow to output similar words). Note that since the model is trained
# on a small corpus, some of the relations might not be clear.
@task(cache_version="1.0", cache=True, limits=Resources(mem="600Mi"))
def word_similarities(
    model_ser: FlyteFile[MODELSER_NLP], word: str
) -> Dict[str, float]:
    model = Word2Vec.load(model_ser.download())
    wv = model.wv
    logger.info(f"Word vector for {word}:{wv[word]}")
    return dict(wv.most_similar(word, topn=10))


# %%
# Sentence Similarity
# ===================
#
# We compute Word Mover’s Distance (WMD) using the trained embeddings of words.
# This enables us to assess the distance between two documents in a meaningful way even when they have
# no words in common.
# WMD outputs a large value for two completely unrelated sentences and small value for two closely related
# sentences.
# Since we chose two similar sentences for comparison, the word movers distance
# should be small. You can try altering either ``SENTENCE_A`` or ``SENTENCE_B`` variables to be dissimilar
# to the other sentence, and check if the value computed is larger.
@task(cache_version="1.0", cache=True, limits=Resources(mem="600Mi"))
def word_movers_distance(model_ser: FlyteFile[MODELSER_NLP]) -> float:
    sentences = [SENTENCE_A, SENTENCE_B]
    results = []
    for i in sentences:
        result = [w for w in utils.tokenize(i) if w not in STOPWORDS]
        results.append(result)
    model = Word2Vec.load(model_ser.download())
    logger.info(f"Computing word movers distance for: {SENTENCE_A} and {SENTENCE_B} ")
    return model.wv.wmdistance(*results)


# %%
# Dimensionality Reduction and Plotting
# =====================================
#
# The word embeddings made by the model can be visualized after reducing the dimensionality to two with t-SNE.
# This task can take a few minutes to complete.
@task(cache_version="1.0", cache=True, limits=Resources(mem="1000Mi"))
def dimensionality_reduction(model_ser: FlyteFile[MODELSER_NLP]) -> plotdata:
    model = Word2Vec.load(model_ser.download())
    num_dimensions = 2
    vectors = np.asarray(model.wv.vectors)
    labels = np.asarray(model.wv.index_to_key)
    logger.info("Running dimensionality reduction using t-SNE")
    tsne = TSNE(n_components=num_dimensions, random_state=0)
    vectors = tsne.fit_transform(vectors)
    x_vals = [float(v[0]) for v in vectors]
    y_vals = [float(v[1]) for v in vectors]
    labels = [str(l) for l in labels]
    return x_vals, y_vals, labels


@task(
    cache_version="1.0", cache=True, limits=Resources(mem="600Mi"), disable_deck=False
)
def plot_with_plotly(x: List[float], y: List[float], labels: List[str]):
    layout = go.Layout(height=600, width=800)
    fig = go.Figure(
        data=go.Scattergl(x=x, y=y, mode="markers", marker=dict(color="aqua")),
        layout=layout,
    )
    indices = list(range(len(labels)))
    selected_indices = random.sample(indices, 50)
    for i in selected_indices:
        fig.add_annotation(
            text=labels[i],
            x=x[i],
            y=y[i],
            showarrow=False,
            font=dict(size=15, color="black", family="Sans Serif"),
        )
    logger.info("Generating the Word Embedding Plot using Flyte Deck")
    flytekit.Deck("Word Embeddings", io.to_html(fig, full_html=True))


# %%
# Running the Workflow
# ====================
#
# Let's kick off a workflow! This will return the inference outputs of both gensim models:
# similar words, WMD and LDA topics.
@workflow
def nlp_workflow(target_word: str = "computer") -> workflow_outputs:
    corpus = generate_processed_corpus()
    model_wv = train_word2vec_model(
        training_data=corpus, hyperparams=Word2VecModelHyperparams()
    )
    lda_topics = train_lda_model(corpus=corpus, hyperparams=LDAModelHyperparams())
    similar_words = word_similarities(model_ser=model_wv.model, word=target_word)
    distance = word_movers_distance(model_ser=model_wv.model)
    axis_labels = dimensionality_reduction(model_ser=model_wv.model)
    plot_with_plotly(
        x=axis_labels.x_values, y=axis_labels.y_values, labels=axis_labels.labels
    )
    return similar_words, distance, lda_topics


if __name__ == "__main__":
    print(f"Running {__file__} main...")
    print(nlp_workflow())
