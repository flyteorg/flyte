import tensorflow as tf

from flytekit import task, workflow


@task
def generate_model() -> tf.keras.Model:
    inputs = tf.keras.Input(shape=(32,))
    outputs = tf.keras.layers.Dense(1)(inputs)
    model = tf.keras.Model(inputs, outputs)
    model.compile(
        optimizer=tf.keras.optimizers.Adam(learning_rate=1e-3),
        loss=tf.keras.losses.BinaryCrossentropy(),
        metrics=[
            tf.keras.metrics.BinaryAccuracy(),
        ],
    )
    return model


@task
def generate_sequential_model() -> tf.keras.Sequential:
    model = tf.keras.Sequential(
        [
            tf.keras.layers.Input(shape=(32,)),
            tf.keras.layers.Dense(1),
        ]
    )
    model.compile(
        optimizer=tf.keras.optimizers.Adam(learning_rate=1e-3),
        loss=tf.keras.losses.BinaryCrossentropy(),
        metrics=[
            tf.keras.metrics.BinaryAccuracy(),
        ],
    )
    return model


@task
def model_forward_pass(model: tf.keras.Model) -> tf.Tensor:
    x: tf.Tensor = tf.ones((1, 32))
    return model(x)


@workflow
def wf():
    model1 = generate_model()
    model2 = generate_sequential_model()
    model_forward_pass(model=model1)
    model_forward_pass(model=model2)


def test_wf():
    wf()
