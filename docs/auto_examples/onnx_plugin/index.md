(onnx)=

# ONNX

```{eval-rst}
.. tags:: Integration, MachineLearning, Intermediate
```

Open Neural Network Exchange ([ONNX](https://github.com/onnx/onnx)) is an open standard format for representing machine learning
and deep learning models. It enables interoperability between different frameworks and streamlines the path from research to production.

The flytekit onnx type plugin comes in three flavors:

```{eval-rst}
.. tabbed:: ScikitLearn

  .. code-block::

      pip install flytekitplugins-onnxpytorch

  This plugin enables the conversion from scikitlearn models to ONNX models.
```

```{eval-rst}
.. tabbed:: TensorFlow

  .. code-block::

      pip install flytekitplugins-onnxtensorflow

  This plugin enables the conversion from tensorflow models to ONNX models.
```

```{eval-rst}
.. tabbed:: PyTorch

  .. code-block::

      pip install flytekitplugins-onnxpytorch

  This plugin enables the conversion from pytorch models to ONNX models.
```

:::{note}
If you'd like to add support for a new framework, please create an issue and submit a pull request to the flytekit repo.
You can find the ONNX plugin source code [here](https://github.com/flyteorg/flytekit/tree/master/plugins).
:::

```{auto-examples-toc}
pytorch_onnx
scikitlearn_onnx
tensorflow_onnx
```
