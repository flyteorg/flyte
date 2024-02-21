from flytekit.models import named_entity


def test_identifier():
    obj = named_entity.NamedEntityIdentifier("proj", "development", "MyWorkflow")
    obj2 = named_entity.NamedEntityIdentifier.from_flyte_idl(obj.to_flyte_idl())
    assert obj == obj2


def test_metadata():
    obj = named_entity.NamedEntityMetadata("i am a description", named_entity.NamedEntityState.ACTIVE)
    obj2 = named_entity.NamedEntityMetadata.from_flyte_idl(obj.to_flyte_idl())
    assert obj == obj2
