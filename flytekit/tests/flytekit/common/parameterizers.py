from datetime import timedelta
from itertools import product

from flyteidl.core import tasks_pb2

from flytekit.extras.accelerators import A100, T4
from flytekit.models import interface, literals, security, task, types
from flytekit.models.core import identifier
from flytekit.models.core import types as _core_types

LIST_OF_SCALAR_LITERAL_TYPES = [
    types.LiteralType(simple=types.SimpleType.BINARY),
    types.LiteralType(simple=types.SimpleType.BOOLEAN),
    types.LiteralType(simple=types.SimpleType.DATETIME),
    types.LiteralType(simple=types.SimpleType.DURATION),
    types.LiteralType(simple=types.SimpleType.ERROR),
    types.LiteralType(simple=types.SimpleType.FLOAT),
    types.LiteralType(simple=types.SimpleType.INTEGER),
    types.LiteralType(simple=types.SimpleType.NONE),
    types.LiteralType(simple=types.SimpleType.STRING),
    types.LiteralType(
        schema=types.SchemaType(
            [
                types.SchemaType.SchemaColumn("a", types.SchemaType.SchemaColumn.SchemaColumnType.INTEGER),
                types.SchemaType.SchemaColumn("b", types.SchemaType.SchemaColumn.SchemaColumnType.BOOLEAN),
                types.SchemaType.SchemaColumn("c", types.SchemaType.SchemaColumn.SchemaColumnType.DATETIME),
                types.SchemaType.SchemaColumn("d", types.SchemaType.SchemaColumn.SchemaColumnType.DURATION),
                types.SchemaType.SchemaColumn("e", types.SchemaType.SchemaColumn.SchemaColumnType.FLOAT),
                types.SchemaType.SchemaColumn("f", types.SchemaType.SchemaColumn.SchemaColumnType.STRING),
            ]
        )
    ),
    types.LiteralType(
        blob=_core_types.BlobType(
            format="",
            dimensionality=_core_types.BlobType.BlobDimensionality.SINGLE,
        )
    ),
    types.LiteralType(
        blob=_core_types.BlobType(
            format="csv",
            dimensionality=_core_types.BlobType.BlobDimensionality.SINGLE,
        )
    ),
    types.LiteralType(
        blob=_core_types.BlobType(
            format="",
            dimensionality=_core_types.BlobType.BlobDimensionality.MULTIPART,
        )
    ),
    types.LiteralType(
        blob=_core_types.BlobType(
            format="csv",
            dimensionality=_core_types.BlobType.BlobDimensionality.MULTIPART,
        )
    ),
    types.LiteralType(
        union_type=types.UnionType(
            variants=[
                types.LiteralType(simple=types.SimpleType.STRING, structure=types.TypeStructure(tag="str")),
                types.LiteralType(simple=types.SimpleType.INTEGER, structure=types.TypeStructure(tag="int")),
            ]
        )
    ),
]


LIST_OF_COLLECTION_LITERAL_TYPES = [
    types.LiteralType(collection_type=literal_type) for literal_type in LIST_OF_SCALAR_LITERAL_TYPES
]

LIST_OF_NESTED_COLLECTION_LITERAL_TYPES = [
    types.LiteralType(collection_type=literal_type) for literal_type in LIST_OF_COLLECTION_LITERAL_TYPES
]

LIST_OF_ALL_LITERAL_TYPES = (
    LIST_OF_SCALAR_LITERAL_TYPES + LIST_OF_COLLECTION_LITERAL_TYPES + LIST_OF_NESTED_COLLECTION_LITERAL_TYPES
)

LIST_OF_INTERFACES = [
    interface.TypedInterface(
        {"a": interface.Variable(t, "description 1")},
        {"b": interface.Variable(t, "description 2")},
    )
    for t in LIST_OF_ALL_LITERAL_TYPES
]


LIST_OF_RESOURCE_ENTRIES = [
    task.Resources.ResourceEntry(task.Resources.ResourceName.CPU, "1"),
    task.Resources.ResourceEntry(task.Resources.ResourceName.GPU, "1"),
    task.Resources.ResourceEntry(task.Resources.ResourceName.MEMORY, "1G"),
    task.Resources.ResourceEntry(task.Resources.ResourceName.EPHEMERAL_STORAGE, "1G"),
]


LIST_OF_RESOURCE_ENTRY_LISTS = [LIST_OF_RESOURCE_ENTRIES]


LIST_OF_RESOURCES = [
    task.Resources(request, limit)
    for request, limit in product(LIST_OF_RESOURCE_ENTRY_LISTS, LIST_OF_RESOURCE_ENTRY_LISTS)
]


LIST_OF_RUNTIME_METADATA = [
    task.RuntimeMetadata(task.RuntimeMetadata.RuntimeType.OTHER, "1.0.0", "python"),
    task.RuntimeMetadata(task.RuntimeMetadata.RuntimeType.FLYTE_SDK, "1.0.0b0", "golang"),
]


LIST_OF_RETRY_POLICIES = [literals.RetryStrategy(retries=i) for i in [0, 1, 3, 100]]

LIST_OF_INTERRUPTIBLE = [None, True, False]

LIST_OF_TASK_METADATA = [
    task.TaskMetadata(
        discoverable,
        runtime_metadata,
        timeout,
        retry_strategy,
        interruptible,
        discovery_version,
        deprecated,
        cache_serializable,
        pod_template_name,
    )
    for discoverable, runtime_metadata, timeout, retry_strategy, interruptible, discovery_version, deprecated, cache_serializable, pod_template_name in product(
        [True, False],
        LIST_OF_RUNTIME_METADATA,
        [timedelta(days=i) for i in range(3)],
        LIST_OF_RETRY_POLICIES,
        LIST_OF_INTERRUPTIBLE,
        ["1.0"],
        ["deprecated"],
        [True, False],
        ["A", "B"],
    )
]

LIST_OF_TASK_TEMPLATES = [
    task.TaskTemplate(
        identifier.Identifier(identifier.ResourceType.TASK, "project", "domain", "name", "version"),
        "python",
        task_metadata,
        interfaces,
        {"a": 1, "b": [1, 2, 3], "c": "abc", "d": {"x": 1, "y": 2, "z": 3}},
        container=task.Container(
            "my_image",
            ["this", "is", "a", "cmd"],
            ["this", "is", "an", "arg"],
            resources,
            {"a": "b"},
            {"d": "e"},
        ),
    )
    for task_metadata, interfaces, resources in product(LIST_OF_TASK_METADATA, LIST_OF_INTERFACES, LIST_OF_RESOURCES)
]

LIST_OF_CONTAINERS = [
    task.Container(
        "my_image",
        ["this", "is", "a", "cmd"],
        ["this", "is", "an", "arg"],
        resources,
        {"a": "b"},
        {"d": "e"},
    )
    for resources in LIST_OF_RESOURCES
]

LIST_OF_TASK_CLOSURES = [task.TaskClosure(task.CompiledTask(template)) for template in LIST_OF_TASK_TEMPLATES]

LIST_OF_SCALARS_AND_PYTHON_VALUES = [
    (literals.Scalar(primitive=literals.Primitive(integer=100)), 100),
    (literals.Scalar(primitive=literals.Primitive(float_value=500.0)), 500.0),
    (literals.Scalar(primitive=literals.Primitive(boolean=True)), True),
    (literals.Scalar(primitive=literals.Primitive(string_value="hello")), "hello"),
    (
        literals.Scalar(primitive=literals.Primitive(duration=timedelta(seconds=5))),
        timedelta(seconds=5),
    ),
    (literals.Scalar(none_type=literals.Void()), None),
    (
        literals.Scalar(
            union=literals.Union(
                value=literals.Literal(scalar=literals.Scalar(primitive=literals.Primitive(integer=10))),
                stored_type=types.LiteralType(
                    simple=types.SimpleType.INTEGER, structure=types.TypeStructure(tag="int")
                ),
            )
        ),
        10,
    ),
    (
        literals.Scalar(
            union=literals.Union(
                value=literals.Literal(scalar=literals.Scalar(primitive=literals.Primitive(integer=10))),
                stored_type=types.LiteralType(
                    simple=types.SimpleType.INTEGER, structure=types.TypeStructure(tag="int")
                ),
            )
        ),
        10,
    ),
    (
        literals.Scalar(
            union=literals.Union(
                value=literals.Literal(scalar=literals.Scalar(primitive=literals.Primitive(string_value="test"))),
                stored_type=types.LiteralType(simple=types.SimpleType.STRING, structure=types.TypeStructure(tag="str")),
            )
        ),
        "test",
    ),
    (
        literals.Scalar(
            union=literals.Union(
                value=literals.Literal(scalar=literals.Scalar(primitive=literals.Primitive(string_value="test"))),
                stored_type=types.LiteralType(simple=types.SimpleType.STRING, structure=types.TypeStructure(tag="str")),
            )
        ),
        "test",
    ),
]

LIST_OF_SCALAR_LITERALS_AND_PYTHON_VALUE = [
    (literals.Literal(scalar=s), v) for s, v in LIST_OF_SCALARS_AND_PYTHON_VALUES
]

LIST_OF_LITERAL_COLLECTIONS_AND_PYTHON_VALUE = [
    (literals.LiteralCollection(literals=[l, l, l]), [v, v, v]) for l, v in LIST_OF_SCALAR_LITERALS_AND_PYTHON_VALUE
]

LIST_OF_ALL_LITERALS_AND_VALUES = (
    LIST_OF_SCALAR_LITERALS_AND_PYTHON_VALUE + LIST_OF_LITERAL_COLLECTIONS_AND_PYTHON_VALUE
)

LIST_OF_SECRETS = [
    None,
    security.Secret(group="x", key="g"),
    security.Secret(group="x", key="y", mount_requirement=security.Secret.MountType.ANY),
    security.Secret(group="x", key="y", group_version="1", mount_requirement=security.Secret.MountType.FILE),
]

LIST_RUN_AS = [
    None,
    security.Identity(iam_role="role"),
    security.Identity(k8s_service_account="service_account"),
]

LIST_OF_SECURITY_CONTEXT = [
    security.SecurityContext(run_as=r, secrets=s, tokens=None) for r in LIST_RUN_AS for s in LIST_OF_SECRETS
] + [None]

LIST_OF_ACCELERATORS = [
    None,
    T4,
    A100,
    A100.unpartitioned,
    A100.partition_1g_5gb,
]

LIST_OF_EXTENDED_RESOURCES = [
    None,
    *[
        tasks_pb2.ExtendedResources(gpu_accelerator=None if accelerator is None else accelerator.to_flyte_idl())
        for accelerator in LIST_OF_ACCELERATORS
    ],
]
