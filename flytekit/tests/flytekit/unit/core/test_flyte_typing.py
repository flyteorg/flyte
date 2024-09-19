from flytekit.types.file.file import FlyteFile


def test_filepath_equality():
    a = FlyteFile("/tmp")
    b = FlyteFile("/tmp")
    assert str(b) == "/tmp"
    assert a == b

    a = FlyteFile("/tmp")
    b = FlyteFile["pdf"]("/tmp")
    assert a != b

    a = FlyteFile("/tmp")
    b = FlyteFile("/tmp/c")
    assert a != b

    x = "jpg"
    y = ".jpg"
    a = FlyteFile[x]("/tmp")
    b = FlyteFile[y]("/tmp")
    assert a == b


def test_fdff():
    a = FlyteFile["txt"]("/tmp")
    print(a)
    b = FlyteFile["txt"]("/tmp")
    assert a == b
