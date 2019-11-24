Simplest Example of a Flyte Workflow
======================================

Introduction:
-------------
In this example, we take one image as an input and run canny edge detection algorithm on that image. the output is an edge highlighted image.

Note:
-----
We could have used a Types.Blob to represent the input, but we also wanted to show that this complete type-safety is optional
Users can choose to use string to represent remote objects.
For examples where we use more complex types and see how they interplay with other python libraries look at other examples, like the **DiabetesXGBoostModelTrainer**