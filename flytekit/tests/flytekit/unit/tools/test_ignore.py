import os
import subprocess
from pathlib import Path
from tarfile import TarInfo
from typing import Dict
from unittest.mock import patch

import pytest
from docker.utils.build import PatternMatcher

from flytekit.tools.ignore import DockerIgnore, GitIgnore, IgnoreGroup, StandardIgnore


def make_tree(root: Path, tree: Dict):
    for name, content in tree.items():
        if isinstance(content, dict):
            directory = root / name
            directory.mkdir()
            make_tree(directory, content)
        if isinstance(content, str):
            file = root / name
            file.write_text(content)


@pytest.fixture
def simple_gitignore(tmp_path):
    tree = {
        "sub": {"some.bar": ""},
        "test.foo": "",
        "keep.foo": "",
        ".gitignore": "\n".join(["*.foo", "!keep.foo", "# A comment", "sub"]),
    }

    make_tree(tmp_path, tree)
    subprocess.run(["git", "init", str(tmp_path)])
    return tmp_path


@pytest.fixture
def nested_gitignore(tmp_path):
    tree = {
        "sub": {"some.foo": "", "another.foo": "", ".gitignore": "!another.foo"},
        "data": {".gitignore": "*", "large.file": ""},
        "test.foo": "",
        "keep.foo": "",
        ".gitignore": "\n".join(
            [
                "*.foo",
                "!keep.foo",
                "# A comment",
            ]
        ),
    }

    make_tree(tmp_path, tree)
    subprocess.run(["git", "init", str(tmp_path)])
    return tmp_path


@pytest.fixture
def simple_dockerignore(tmp_path):
    tree = {
        "sub": {"some.bar": ""},
        "test.foo": "",
        "keep.foo": "",
        ".dockerignore": "\n".join(["*.foo", "!keep.foo", "# A comment", "sub"]),
    }

    make_tree(tmp_path, tree)
    return tmp_path


@pytest.fixture
def no_ignore(tmp_path):
    tree = {
        "sub": {"some.bar": ""},
        "test.foo": "",
        "keep.foo": "",
    }

    make_tree(tmp_path, tree)
    return tmp_path


@pytest.fixture
def all_ignore(tmp_path):
    tree = {
        "sub": {
            "some.bar": "",
            "__pycache__": {"some.pyc"},
        },
        "data": {"reallybigfile.bar": ""},
        ".cache": {"something.cached": ""},
        "test.foo": "",
        "keep.foo": "",
        ".dockerignore": "\n".join(["# A comment", "data", ".git"]),
        ".gitignore": "\n".join(["*.foo", "!keep.foo"]),
    }

    make_tree(tmp_path, tree)
    subprocess.run(["git", "init", str(tmp_path)])
    return tmp_path


def test_simple_gitignore(simple_gitignore):
    gitignore = GitIgnore(simple_gitignore)
    assert gitignore.is_ignored(str(simple_gitignore / "test.foo"))
    assert gitignore.is_ignored(str(simple_gitignore / "sub"))
    assert gitignore.is_ignored(str(simple_gitignore / "sub" / "some.bar"))
    assert not gitignore.is_ignored(str(simple_gitignore / "keep.foo"))
    assert not gitignore.is_ignored(str(simple_gitignore / ".gitignore"))
    assert not gitignore.is_ignored(str(simple_gitignore / ".git"))


def test_not_subpath(simple_gitignore):
    """Test edge case that if path is not on root it cannot not be ignored"""
    gitignore = GitIgnore(simple_gitignore)
    if os.name == "nt":
        # Relative paths can only be compared if files are in the same drive
        assert not gitignore.is_ignored(str(Path(simple_gitignore.drive) / "whatever" / "test.foo"))
    else:
        assert not gitignore.is_ignored("/whatever/test.foo")


def test_nested_gitignore(nested_gitignore):
    """Test override with nested gitignore and star-ignore"""
    gitignore = GitIgnore(nested_gitignore)
    assert gitignore.is_ignored(str(nested_gitignore / "test.foo"))
    assert not gitignore.is_ignored(str(nested_gitignore / "sub"))
    assert gitignore.is_ignored(str(nested_gitignore / "sub" / "some.foo"))
    assert gitignore.is_ignored(str(nested_gitignore / "data" / "large.file"))
    assert not gitignore.is_ignored(str(nested_gitignore / "sub" / "another.foo"))
    assert not gitignore.is_ignored(str(nested_gitignore / ".gitignore"))
    assert not gitignore.is_ignored(str(nested_gitignore / ".git"))


@patch("flytekit.tools.ignore.which")
def test_no_git(mock_which, simple_gitignore):
    """Test that nothing is ignored if no git cli available"""
    mock_which.return_value = None
    gitignore = GitIgnore(simple_gitignore)
    assert not gitignore.has_git
    assert not gitignore.is_ignored(str(simple_gitignore / "test.foo"))
    assert not gitignore.is_ignored(str(simple_gitignore / "sub"))
    assert not gitignore.is_ignored(str(simple_gitignore / "keep.foo"))
    assert not gitignore.is_ignored(str(simple_gitignore / ".gitignore"))
    assert not gitignore.is_ignored(str(simple_gitignore / ".git"))


def test_dockerignore_parse(simple_dockerignore):
    """Test .dockerignore file parsing"""
    dockerignore = DockerIgnore(simple_dockerignore)
    assert [p.cleaned_pattern for p in dockerignore.pm.patterns] == ["*.foo", "keep.foo", "sub", ".dockerignore"]
    assert [p.exclusion for p in dockerignore.pm.patterns] == [False, True, False, True]


def test_patternmatcher():
    """Test that PatternMatcher works as expected"""
    patterns = ["*.foo", "!keep.foo", "sub"]
    pm = PatternMatcher(patterns)
    assert pm.matches("whatever.foo")
    assert not pm.matches("keep.foo")
    assert pm.matches("sub")
    assert pm.matches("sub/stuff.txt")


def test_simple_dockerignore(simple_dockerignore):
    dockerignore = DockerIgnore(simple_dockerignore)
    assert dockerignore.is_ignored(str(simple_dockerignore / "test.foo"))
    assert dockerignore.is_ignored(str(simple_dockerignore / "sub"))
    assert dockerignore.is_ignored(str(simple_dockerignore / "sub" / "some.bar"))
    assert not dockerignore.is_ignored(str(simple_dockerignore / "keep.foo"))


def test_no_ignore(no_ignore):
    """Test that nothing is ignored if no ignore files present"""
    dockerignore = DockerIgnore(no_ignore)
    assert not dockerignore.is_ignored(str(no_ignore / "test.foo"))
    assert not dockerignore.is_ignored(str(no_ignore / "keep.foo"))
    assert not dockerignore.is_ignored(str(no_ignore / "sub"))
    assert not dockerignore.is_ignored(str(no_ignore / "sub" / "some.bar"))

    gitignore = GitIgnore(no_ignore)
    assert not gitignore.is_ignored(str(no_ignore / "test.foo"))
    assert not gitignore.is_ignored(str(no_ignore / "keep.foo"))
    assert not gitignore.is_ignored(str(no_ignore / "sub"))
    assert not gitignore.is_ignored(str(no_ignore / "sub" / "some.bar"))


def test_standard_ignore():
    """Test the standard ignore cases previously hardcoded"""
    patterns = ["*.pyc", ".cache", ".cache/*", "__pycache__", "**/__pycache__", "*.foo"]
    ignore = StandardIgnore(root=".", patterns=patterns)
    assert not ignore.is_ignored("foo.py")
    assert ignore.is_ignored("foo.pyc")
    assert ignore.is_ignored(".cache/foo")
    assert ignore.is_ignored("__pycache__")
    assert ignore.is_ignored("foo/__pycache__")
    assert ignore.is_ignored("spam/ham/some.foo")


def test_all_ignore(all_ignore):
    """Test all ignores grouped together"""
    ignore = IgnoreGroup(all_ignore, [GitIgnore, DockerIgnore, StandardIgnore])
    assert not ignore.is_ignored("sub")
    assert not ignore.is_ignored("sub/some.bar")
    assert ignore.is_ignored("sub/__pycache__")
    assert ignore.is_ignored("sub/__pycache__/some.pyc")
    assert ignore.is_ignored("data")
    assert ignore.is_ignored("data/reallybigfile.bar")
    assert ignore.is_ignored(".cache")
    assert ignore.is_ignored(".cache/something.cached")
    assert ignore.is_ignored("test.foo")
    assert not ignore.is_ignored("keep.foo")
    assert not ignore.is_ignored(".gitignore")
    assert not ignore.is_ignored(".dockerignore")
    assert ignore.is_ignored(".git")


def test_all_ignore_tar_filter(all_ignore):
    """Test tar_filter method of all ignores grouped together"""
    ignore = IgnoreGroup(all_ignore, [GitIgnore, DockerIgnore, StandardIgnore])
    assert ignore.tar_filter(TarInfo(name="sub")).name == "sub"
    assert ignore.tar_filter(TarInfo(name="sub/some.bar")).name == "sub/some.bar"
    assert not ignore.tar_filter(TarInfo(name="sub/__pycache__"))
    assert not ignore.tar_filter(TarInfo(name="sub/__pycache__/some.pyc"))
    assert not ignore.tar_filter(TarInfo(name="data"))
    assert not ignore.tar_filter(TarInfo(name="data/reallybigfile.bar"))
    assert not ignore.tar_filter(TarInfo(name=".cache"))
    assert not ignore.tar_filter(TarInfo(name=".cache/something.cached"))
    assert not ignore.tar_filter(TarInfo(name="test.foo"))
    assert ignore.tar_filter(TarInfo(name="keep.foo")).name == "keep.foo"
    assert ignore.tar_filter(TarInfo(name=".gitignore")).name == ".gitignore"
    assert ignore.tar_filter(TarInfo(name=".dockerignore")).name == ".dockerignore"
    assert not ignore.tar_filter(TarInfo(name=".git"))
