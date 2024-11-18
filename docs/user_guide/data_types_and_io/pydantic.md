(pydantic)=

# Pydantic BaseModel

```{eval-rst}
.. tags:: Basic
```

When you've multiple values that you want to send across Flyte entities, and you want them to have, you can use a `pydantic.BaseModel`.
Note:
You can put Dataclass and FlyteTypes (FlyteFile, FlyteDirectory, FlyteSchema, and StructuredDataset) in a pydantic BaseModel.

:::{important}
Pydantic BaseModle V2 only works when you are using flytekit version >= v1.14.0.

If you're using Flytekit version >= v1.14.0 and you want to produce protobuf struct literal for pydantic basemodels, 
you can set environment variable  `FLYTE_USE_OLD_DC_FORMAT` to `true`.
:::
