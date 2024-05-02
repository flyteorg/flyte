# Accelerated datasets

:::{admonition} *Accelerated datasets* and *Accelerators* are entirely different things
Accelerated datasets is a Union feature that enables quick access to large datasets from within a task.
An [accelerator](../core-concepts/tasks/task-hardware-environment/accelerators), on the other hand, is a specialized hardware device that is used to accelerate the execution of a task.
These concepts are entirely different and should not be confused.
:::

Many of the workflows that you may want to run in Union will involve tasks that use large static assets such as reference genomes, training datasets, or pre-trained models.
These assets are often stored in an object store and need to be downloaded to the task pod each time before the task can run.
This can be a significant bottleneck, especially if the data must be loaded into memory to be randomly accessed and therefore cannot be streamed.

To remedy this, Union provides a way to pre-load large static assets into a shared object store that is mounted to all machine nodes in your cluster by default.
This allows you to upload your data once and then access it from any task without needing to download it each time.

Data items stored in this way are called an accelerated datasets.

:::{admonition} Only on S3
Currently, this feature is only available for AWS S3.
:::

## How it works

* Each customer has a dedicated S3 bucket where they can store their accelerated datasets.
* The naming and set up of this bucket must be coordinated with the Union team, in order that a suitable name is chosen. In general it will usually be something like `s3://union-<org-name>-persistent`.
* You can upload any data you wish to this bucket.
* The bucket will be automatically mounted into every node in your cluster using a mechanism that make it locally accessible through `FlyteFile` (see below). To your task logic, it will appear to be a local directory in the task container.

## Example usage

Assuming that your organization is called `my-company` and the file you want to access is called `my_data.csv`, you would first need to upload the file to the persistent bucket. See [Upload a File to Your Amazon S3 Bucket](https://docs.aws.amazon.com/quickstarts/latest/s3backup/step-2-upload-file.html).

The code to access the data looks like this:

```{code-block} python
from flytekit.types.file import FlyteFile

@task
def my_task()
    my_data = FlyteFile(path='s3://union-my-company-persistent/my_data.csv')
    contents = open(my_data).read()
    // Do something with the contents
```

Note that you do not have to invoke `FlyteFile.download()` because the file will already have been made available locally within the container.

## Considerations

### Caching

While the persistent bucket appears to your task as a locally mounted volume, the data itself will not be resident in the local file system until after the first access. After the first access it will be cached locally. This fact should be taken into account when using this feature.

### Storage consumption

Data cached during the use of accelerated datasets will consume local storage on the nodes in your cluster. This should be taken into account when selecting and sizing your cluster nodes.
