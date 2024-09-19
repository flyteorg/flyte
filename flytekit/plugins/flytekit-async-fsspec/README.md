# Flytekit Async fsspec Plugin

The Flyte async fsspec plugin is a powerful addition to the Flyte ecosystem designed to optimize the performance of object transmission. This plugin focuses on overriding key methods of the file systems in fsspec to introduce efficiency improvements, resulting in accelerated data transfers between Flyte workflows and object storage.

Currently, the async fsspec plugin improves the following file systems:
1. s3fs

To install the plugin, run the following command:

```bash
pip install flytekitplugins-async-fsspec
```

Once installed, the plugin will automatically override the original file system and register optimized ones, seamlessly integrating with your Flyte workflows.
