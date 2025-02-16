from flytekit.tools import module_loader


def test_load_object():
    loader_self = module_loader.load_object_from_module(f"{module_loader.__name__}.load_object_from_module")
    assert loader_self.__module__ == f"{module_loader.__name__}"
