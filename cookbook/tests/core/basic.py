import os

from core.basic import basic_workflow, files, folders, mocking


def test_basic_wf():
    res = basic_workflow.my_wf(a=3, b="hello")
    assert res == (5, "helloworld")


def test_folders():
    output_dir = folders.download_and_rotate()
    files = os.listdir(output_dir)
    assert len(files) == 2


def test_files():
    output_file = files.rotate_one_workflow(in_image=files.default_images[0])
    assert os.path.exists(output_file)


def test_mocks():
    # Assertions already in main
    mocking.main_1()
    mocking.main_2()
