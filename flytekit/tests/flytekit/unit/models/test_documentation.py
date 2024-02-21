from flytekit.models.documentation import Description, Documentation, SourceCode


def test_long_description():
    value = "long"
    icon_link = "http://icon"
    obj = Description(value=value, icon_link=icon_link)
    assert Description.from_flyte_idl(obj.to_flyte_idl()) == obj
    assert obj.value == value
    assert obj.icon_link == icon_link
    assert obj.format == Description.DescriptionFormat.RST


def test_source_code():
    link = "https://github.com/flyteorg/flytekit"
    obj = SourceCode(link=link)
    assert SourceCode.from_flyte_idl(obj.to_flyte_idl()) == obj
    assert obj.link == link


def test_documentation():
    short_description = "short"
    long_description = Description(value="long", icon_link="http://icon")
    source_code = SourceCode(link="https://github.com/flyteorg/flytekit")
    obj = Documentation(short_description=short_description, long_description=long_description, source_code=source_code)
    assert Documentation.from_flyte_idl(obj.to_flyte_idl()) == obj
    assert obj.short_description == short_description
    assert obj.long_description == long_description
    assert obj.source_code == source_code
