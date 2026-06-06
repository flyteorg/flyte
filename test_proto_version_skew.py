#!/usr/bin/env python3
"""Protobuf Version-Skew Tester for Flyte

Detects protobuf message changes on a branch and verifies forward/backward
compatibility of the wire format between old (base) and new (head) versions.

The tool:
  1. Diffs .proto files between base and head to find changed messages
  2. Extracts Python (flytekit) and Go (flyte) protobuf bindings from both versions
  3. Serializes with Python bindings (client), deserializes with Go bindings (backend)
  4. Reports wire-format hex and deserialized field values

Usage:
    .venv/bin/python test_proto_version_skew.py
    .venv/bin/python test_proto_version_skew.py --base master --head HEAD
    .venv/bin/python test_proto_version_skew.py --repo /path/to/flyte

Requires the `protobuf` package (used by generated bindings in subprocesses).
"""

import argparse
import os
import re
import subprocess
import sys
import tempfile
from dataclasses import dataclass, field as dc_field
from typing import Dict, List, Optional, Tuple


# ---------------------------------------------------------------------------
# Data structures
# ---------------------------------------------------------------------------

@dataclass
class ProtoField:
    name: str
    type: str
    number: int
    repeated: bool = False


@dataclass
class ProtoEnum:
    name: str
    values: List[Tuple[str, int]]


@dataclass
class ProtoMessage:
    name: str
    fields: List[ProtoField] = dc_field(default_factory=list)
    enums: List[ProtoEnum] = dc_field(default_factory=list)


@dataclass
class FieldTypeChange:
    name: str
    old_type: str
    new_type: str


@dataclass
class MessageChange:
    proto_file: str
    message_name: str
    added_fields: List[ProtoField]
    removed_fields: List[ProtoField]
    type_changed_fields: List[FieldTypeChange]
    old_message: Optional[ProtoMessage]
    new_message: ProtoMessage


# ---------------------------------------------------------------------------
# Proto parser — extracts top-level messages with fields and nested enums
# ---------------------------------------------------------------------------

def parse_proto(content: str) -> List[ProtoMessage]:
    lines = content.splitlines()
    messages: List[ProtoMessage] = []
    i = 0
    while i < len(lines):
        line = _strip_comment(lines[i])
        if re.match(r"message\s+\w+\s*\{", line):
            msg, i = _parse_message(lines, i)
            messages.append(msg)
        else:
            i += 1
    return messages


def _strip_comment(line: str) -> str:
    pos = line.find("//")
    return (line[:pos] if pos >= 0 else line).strip()


def _parse_message(lines: List[str], start: int) -> Tuple[ProtoMessage, int]:
    name = re.match(r"message\s+(\w+)", _strip_comment(lines[start])).group(1)
    msg = ProtoMessage(name=name)

    depth, i = 0, start
    while i < len(lines):
        raw = _strip_comment(lines[i])
        depth += raw.count("{") - raw.count("}")
        if i > start and depth <= 0:
            break
        i += 1
    block_end = i

    j = start + 1
    while j < block_end:
        raw = _strip_comment(lines[j])

        if re.match(r"(message|oneof)\s+\w+\s*\{", raw):
            d = 1
            j += 1
            while j < block_end and d > 0:
                d += _strip_comment(lines[j]).count("{") - _strip_comment(lines[j]).count("}")
                j += 1
            continue

        em = re.match(r"enum\s+(\w+)\s*\{", raw)
        if em:
            vals: List[Tuple[str, int]] = []
            j += 1
            while j < block_end:
                er = _strip_comment(lines[j])
                if "}" in er:
                    j += 1
                    break
                vm = re.match(r"(\w+)\s*=\s*(-?\d+)", er)
                if vm:
                    vals.append((vm.group(1), int(vm.group(2))))
                j += 1
            msg.enums.append(ProtoEnum(name=em.group(1), values=vals))
            continue

        fm = re.match(r"(repeated\s+)?(\w[\w.]*)\s+(\w+)\s*=\s*(\d+)", raw)
        if fm and fm.group(2) not in ("option", "reserved", "extensions"):
            msg.fields.append(ProtoField(
                name=fm.group(3), type=fm.group(2),
                number=int(fm.group(4)), repeated=bool(fm.group(1)),
            ))
        j += 1

    return msg, block_end + 1


# ---------------------------------------------------------------------------
# Git helpers
# ---------------------------------------------------------------------------

def _git(repo: str, *args: str) -> str:
    r = subprocess.run(
        ["git", "-C", repo] + list(args),
        capture_output=True, text=True, check=True,
    )
    return r.stdout.strip()


def _git_file(repo: str, ref: str, path: str) -> Optional[str]:
    try:
        return _git(repo, "show", f"{ref}:{path}")
    except subprocess.CalledProcessError:
        return None


def _extract_pb_python(repo: str, ref: str, dest: str) -> None:
    arc = subprocess.run(
        ["git", "-C", repo, "archive", ref, "--", "flyteidl/gen/pb_python/"],
        capture_output=True, check=True,
    )
    subprocess.run(["tar", "-xf", "-", "-C", dest], input=arc.stdout, check=True)


def _extract_go_module(repo: str, ref: str, dest: str) -> str:
    """Extract flyteidl Go module from a git ref. Returns path to flyteidl dir."""
    arc = subprocess.run(
        ["git", "-C", repo, "archive", ref, "--",
         "flyteidl/go.mod", "flyteidl/go.sum", "flyteidl/gen/pb-go/"],
        capture_output=True, check=True,
    )
    subprocess.run(["tar", "-xf", "-", "-C", dest], input=arc.stdout, check=True)
    flyteidl_dir = os.path.join(dest, "flyteidl")

    # The flyteidl go.mod has a replace directive for flytestdlib pointing at
    # a relative path (../flytestdlib).  Create a minimal stub module so that
    # `go build` can resolve it without needing the real source tree.
    stdlib_dir = os.path.join(dest, "flytestdlib")
    os.makedirs(stdlib_dir, exist_ok=True)
    with open(os.path.join(stdlib_dir, "go.mod"), "w") as f:
        f.write("module github.com/flyteorg/flyte/flytestdlib\n\ngo 1.21\n")

    return flyteidl_dir


# ---------------------------------------------------------------------------
# Change detection
# ---------------------------------------------------------------------------

def detect_changes(repo: str, base: str, head: str) -> List[MessageChange]:
    out = _git(repo, "diff", "--name-only", f"{base}...{head}", "--", "*.proto")
    files = [l for l in out.splitlines() if l.endswith(".proto")] if out else []
    changes: List[MessageChange] = []

    for pf in files:
        old_src = _git_file(repo, base, pf)
        new_src = _git_file(repo, head, pf)
        if not new_src:
            continue
        old_msgs = {m.name: m for m in parse_proto(old_src)} if old_src else {}
        new_msgs = {m.name: m for m in parse_proto(new_src)}

        for nm, new_m in new_msgs.items():
            old_m = old_msgs.get(nm)
            old_names = {f.name for f in old_m.fields} if old_m else set()
            new_names = {f.name for f in new_m.fields}
            added = [f for f in new_m.fields if f.name not in old_names]
            removed = [f for f in (old_m.fields if old_m else []) if f.name not in new_names]

            type_changed: List[FieldTypeChange] = []
            if old_m:
                old_by_name = {f.name: f for f in old_m.fields}
                for f in new_m.fields:
                    old_f = old_by_name.get(f.name)
                    if old_f and old_f.type != f.type:
                        type_changed.append(FieldTypeChange(
                            name=f.name, old_type=old_f.type, new_type=f.type,
                        ))

            if added or removed or type_changed:
                changes.append(MessageChange(
                    proto_file=pf, message_name=nm,
                    added_fields=added, removed_fields=removed,
                    type_changed_fields=type_changed,
                    old_message=old_m, new_message=new_m,
                ))
    return changes


# ---------------------------------------------------------------------------
# Test-value generation
# ---------------------------------------------------------------------------

_SCALAR_VALUES: Dict[str, str] = {
    "string": '"test"', "bytes": 'b"test"', "bool": "True",
    "int32": "1", "int64": "1", "uint32": "1", "uint64": "1",
    "sint32": "1", "sint64": "1", "fixed32": "1", "fixed64": "1",
    "sfixed32": "1", "sfixed64": "1", "float": "1.0", "double": "1.0",
}


def _test_val(field: ProtoField, msg: ProtoMessage, cls: str) -> str:
    if field.type in _SCALAR_VALUES:
        return _SCALAR_VALUES[field.type]
    for en in msg.enums:
        if field.type == en.name:
            # Use the zero-value (first) enum member as the default test value;
            # the caller can override with _test_val_nondefault for cases that
            # need to exercise non-default enum values.
            return f"{cls}.{en.values[0][0]}" if en.values else "0"
    if field.repeated:
        return "[]"
    return "0"


@dataclass
class TestCase:
    new_kw: Dict[str, str]
    old_kw: Dict[str, str]


def make_test_cases(ch: MessageChange) -> List[TestCase]:
    """Return test cases with kwargs typed for each schema version.

    Each TestCase carries two kwargs dicts:
      - new_kw: values matching the *new* proto schema (for new-client serialization)
      - old_kw: values matching the *old* proto schema (for old-client serialization)

    When a field's type changed between versions the two dicts will have
    different values for that field so each side gets a valid literal.
    """
    new_msg = ch.new_message
    old_msg = ch.old_message
    cls = new_msg.name

    old_fields_by_name = {f.name: f for f in old_msg.fields} if old_msg else {}
    new_fields_by_name = {f.name: f for f in new_msg.fields}

    cases: List[TestCase] = []

    # Case 1: fields that exist in *both* schemas
    shared_names = [f.name for f in new_msg.fields if f.name in old_fields_by_name]
    if shared_names:
        new_kw = {n: _test_val(new_fields_by_name[n], new_msg, cls) for n in shared_names}
        old_kw = {n: _test_val(old_fields_by_name[n], old_msg, cls) for n in shared_names}
        cases.append(TestCase(new_kw=new_kw, old_kw=old_kw))

    # Case 2: all new fields (added or type-changed warrant a second case)
    if ch.added_fields or ch.type_changed_fields:
        new_kw = {f.name: _test_val(f, new_msg, cls) for f in new_msg.fields}
        old_kw = {n: _test_val(old_fields_by_name[n], old_msg, cls)
                  for n in shared_names}
        cases.append(TestCase(new_kw=new_kw, old_kw=old_kw))

    return cases


# ---------------------------------------------------------------------------
# Subprocess protobuf operations
# ---------------------------------------------------------------------------

def _proto_module(proto_file: str) -> str:
    """flyteidl/protos/flyteidl/core/types.proto -> flyteidl.core.types_pb2"""
    p = proto_file
    prefix = "flyteidl/protos/"
    if p.startswith(prefix):
        p = p[len(prefix):]
    return p.replace("/", ".").replace(".proto", "_pb2")


def _run_py(pb_dir: str, code: str) -> str:
    full = f"import sys; sys.path.insert(0, {pb_dir!r})\n{code}"
    r = subprocess.run([sys.executable, "-c", full], capture_output=True, text=True)
    if r.returncode != 0:
        raise RuntimeError(r.stderr.strip())
    return r.stdout.strip()


def _serialize(pb_dir: str, mod: str, cls: str, kw: Dict[str, str]) -> str:
    args = ", ".join(f"{k}={v}" for k, v in kw.items())
    return _run_py(pb_dir, f"from {mod} import {cls}\nprint({cls}({args}).SerializeToString().hex())")


def _deserialize(pb_dir: str, mod: str, cls: str, hx: str) -> str:
    code = "\n".join([
        f"from {mod} import {cls}",
        f"msg = {cls}()",
        f'msg.ParseFromString(bytes.fromhex("{hx}"))',
        "parts = []",
        "for f in msg.DESCRIPTOR.fields:",
        "    v = getattr(msg, f.name)",
        "    if f.enum_type:",
        f'        parts.append("{cls}." + f.enum_type.values_by_number[v].name)',
        "    elif isinstance(v, str):",
        '        parts.append(\'"\' + v + \'"\')',
        "    else:",
        "        parts.append(repr(v))",
        f'print("{cls}(" + ", ".join(parts) + ")")',
    ])
    return _run_py(pb_dir, code)


def _fmt_ctor(cls: str, kw: Dict[str, str]) -> str:
    inner = ",\n  ".join(f"{k}={v}" for k, v in kw.items())
    return f"{cls}(\n  {inner},\n)"


# ---------------------------------------------------------------------------
# Go protobuf operations (backend deserialization)
# ---------------------------------------------------------------------------

def _go_import_for_proto(proto_file: str) -> str:
    """Proto file path -> Go import path for the generated package.

    flyteidl/protos/flyteidl/core/types.proto
      -> github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core
    """
    p = proto_file
    prefix = "flyteidl/protos/"
    if p.startswith(prefix):
        p = p[len(prefix):]
    pkg_dir = os.path.dirname(p)
    return f"github.com/flyteorg/flyte/flyteidl/gen/pb-go/{pkg_dir}"


def _proto_full_name(proto_file: str, message_name: str) -> str:
    """Proto file path + message name -> proto full name.

    flyteidl/protos/flyteidl/core/types.proto, BlobType -> flyteidl.core.BlobType
    """
    p = proto_file
    prefix = "flyteidl/protos/"
    if p.startswith(prefix):
        p = p[len(prefix):]
    pkg = os.path.dirname(p).replace("/", ".")
    return f"{pkg}.{message_name}"


def _go_helper_source(go_imports: List[str]) -> str:
    """Generate the Go source for the deserialization helper binary."""
    import_lines = "\n".join(f'\t_ "{pkg}"' for pkg in sorted(set(go_imports)))
    # Avoid f-string because Go source is full of braces.
    return _GO_HELPER_TEMPLATE.replace("GO_IMPORTS_PLACEHOLDER", import_lines)


_GO_HELPER_TEMPLATE = r"""package main

import (
	"encoding/hex"
	"fmt"
	"os"
	"strings"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"

GO_IMPORTS_PLACEHOLDER
)

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintf(os.Stderr, "usage: %s <proto_full_name> <hex>\n", os.Args[0])
		os.Exit(1)
	}
	fullName := protoreflect.FullName(os.Args[1])
	hexData := os.Args[2]

	mt, err := protoregistry.GlobalTypes.FindMessageByName(fullName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "unknown message %q: %v\n", fullName, err)
		os.Exit(1)
	}

	data, err := hex.DecodeString(hexData)
	if err != nil {
		fmt.Fprintf(os.Stderr, "hex decode: %v\n", err)
		os.Exit(1)
	}

	msg := mt.New().Interface()
	if err := proto.Unmarshal(data, msg); err != nil {
		fmt.Fprintf(os.Stderr, "unmarshal: %v\n", err)
		os.Exit(1)
	}

	refl := msg.ProtoReflect()
	desc := refl.Descriptor()
	name := string(desc.Name())
	parts := make([]string, 0, desc.Fields().Len())

	for i := 0; i < desc.Fields().Len(); i++ {
		fd := desc.Fields().Get(i)
		v := refl.Get(fd)
		switch fd.Kind() {
		case protoreflect.StringKind:
			parts = append(parts, fmt.Sprintf(`"%s"`, v.String()))
		case protoreflect.BoolKind:
			if v.Bool() {
				parts = append(parts, "True")
			} else {
				parts = append(parts, "False")
			}
		case protoreflect.EnumKind:
			ev := fd.Enum().Values().ByNumber(v.Enum())
			if ev != nil {
				parts = append(parts, fmt.Sprintf("%s.%s", name, ev.Name()))
			} else {
				parts = append(parts, fmt.Sprintf("%d", v.Enum()))
			}
		case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind,
			protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
			parts = append(parts, fmt.Sprintf("%d", v.Int()))
		case protoreflect.Uint32Kind, protoreflect.Fixed32Kind,
			protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
			parts = append(parts, fmt.Sprintf("%d", v.Uint()))
		case protoreflect.FloatKind, protoreflect.DoubleKind:
			parts = append(parts, fmt.Sprintf("%g", v.Float()))
		case protoreflect.BytesKind:
			parts = append(parts, fmt.Sprintf(`b"%s"`, hex.EncodeToString(v.Bytes())))
		default:
			parts = append(parts, fmt.Sprintf("%v", v.Interface()))
		}
	}

	fmt.Printf("%s(%s)\n", name, strings.Join(parts, ", "))
}
"""


def _build_go_helper(flyteidl_dir: str, go_imports: List[str]) -> str:
    """Write and compile the Go deserialization helper. Returns binary path."""
    helper_dir = os.path.join(flyteidl_dir, "_proto_skew_helper")
    os.makedirs(helper_dir, exist_ok=True)

    with open(os.path.join(helper_dir, "main.go"), "w") as f:
        f.write(_go_helper_source(go_imports))

    binary = os.path.join(helper_dir, "helper")
    r = subprocess.run(
        ["go", "build", "-o", binary, "."],
        cwd=helper_dir,
        capture_output=True, text=True,
    )
    if r.returncode != 0:
        raise RuntimeError(f"Go build failed:\n{r.stderr}")
    return binary


def _deserialize_go(binary: str, full_name: str, hx: str) -> str:
    """Run the Go helper binary to deserialize hex-encoded protobuf bytes."""
    r = subprocess.run([binary, full_name, hx], capture_output=True, text=True)
    if r.returncode != 0:
        raise RuntimeError(r.stderr.strip())
    return r.stdout.strip()


# ---------------------------------------------------------------------------
# Main
# ---------------------------------------------------------------------------

def main() -> None:
    ap = argparse.ArgumentParser(
        description="Test protobuf version-skew between flyte backend and flytekit client",
    )
    ap.add_argument("--base", default="master", help="Base git ref (default: master)")
    ap.add_argument("--head", default="HEAD", help="Head git ref (default: HEAD)")
    ap.add_argument("--repo", default=".", help="Repository root (default: .)")
    args = ap.parse_args()

    try:
        import google.protobuf  # noqa: F401
    except ImportError:
        print(
            "Error: protobuf package not found.\n"
            "Run with a Python that has protobuf installed, e.g.:\n"
            "  .venv/bin/python test_proto_version_skew.py",
            file=sys.stderr,
        )
        sys.exit(1)

    repo = os.path.abspath(args.repo)
    base_hash = _git(repo, "rev-parse", args.base)
    head_hash = _git(repo, "rev-parse", args.head)
    branch = _git(repo, "rev-parse", "--abbrev-ref", args.head)

    remote_label = ""
    try:
        upstream = _git(repo, "rev-parse", "--abbrev-ref", f"{args.head}@{{upstream}}")
        remote_name = upstream.split("/")[0]
        url = _git(repo, "remote", "get-url", remote_name)
        m = re.search(r"[:/]([^/]+/[^/.]+?)(?:\.git)?$", url)
        if m:
            remote_label = m.group(1)
    except (subprocess.CalledProcessError, RuntimeError):
        pass

    # ---- Detect changes ---------------------------------------------------
    changes = detect_changes(repo, args.base, args.head)
    if not changes:
        print(f"No protobuf message changes detected between {args.base} and {args.head}.")
        return

    tag = f"{remote_label}:{branch}" if remote_label else branch
    print(f"\nAnalyzed branch {tag}\n")
    print(f"Detected changes to: {', '.join(c.message_name for c in changes)}\n")

    for ch in changes:
        rel = ch.proto_file
        prefix = "flyteidl/protos/"
        if rel.startswith(prefix):
            rel = rel[len(prefix):]
        print(f"{ch.message_name} ({rel})")
        for f in ch.added_fields:
            print(f"  + {f.type} {f.name} ")
        for f in ch.removed_fields:
            print(f"  - {f.type} {f.name} ")
        for f in ch.type_changed_fields:
            print(f"  ~ {f.name}: {f.old_type} -> {f.new_type}")

    # ---- Set up old / new proto bindings ----------------------------------
    with tempfile.TemporaryDirectory(prefix="proto_skew_") as tmp:
        # Python bindings (for flytekit client serialization)
        old_py_root = os.path.join(tmp, "old_py")
        os.makedirs(old_py_root)
        _extract_pb_python(repo, args.base, old_py_root)
        old_pb = os.path.join(old_py_root, "flyteidl", "gen", "pb_python")

        new_py_root = os.path.join(tmp, "new_py")
        os.makedirs(new_py_root)
        _extract_pb_python(repo, args.head, new_py_root)
        new_pb = os.path.join(new_py_root, "flyteidl", "gen", "pb_python")

        # Go modules (for flyte backend deserialization)
        go_imports = sorted({_go_import_for_proto(ch.proto_file) for ch in changes})

        old_go_root = os.path.join(tmp, "old_go")
        os.makedirs(old_go_root)
        old_go_idl = _extract_go_module(repo, args.base, old_go_root)

        new_go_root = os.path.join(tmp, "new_go")
        os.makedirs(new_go_root)
        new_go_idl = _extract_go_module(repo, args.head, new_go_root)

        print(f"\nBuilding Go deserialization helpers...")
        try:
            old_go_bin = _build_go_helper(old_go_idl, go_imports)
        except RuntimeError as e:
            print(f"Failed to build old Go helper:\n{e}", file=sys.stderr)
            sys.exit(1)
        try:
            new_go_bin = _build_go_helper(new_go_idl, go_imports)
        except RuntimeError as e:
            print(f"Failed to build new Go helper:\n{e}", file=sys.stderr)
            sys.exit(1)

        print(f"\nTesting version-skew\n")
        print(f"Backend (flyte) and Client (flytekit)")
        print(f" - Old: {args.base}@{base_hash}")
        print(f" - New: {branch}@{head_hash}")

        all_ok = True
        for ch in changes:
            py_mod = _proto_module(ch.proto_file)
            cls = ch.message_name
            full_name = _proto_full_name(ch.proto_file, cls)
            cases = make_test_cases(ch)

            # ------ Test 1: New client (Python)  →  Old backend (Go) ------
            print(f"\nTest 1: New client (flytekit) and old backend (flyte)\n")
            seen_new: set = set()
            for tc in cases:
                key = frozenset(tc.new_kw.items())
                if key in seen_new:
                    continue
                seen_new.add(key)
                print(f"message = {_fmt_ctor(cls, tc.new_kw)}\n")
                try:
                    hx = _serialize(new_pb, py_mod, cls, tc.new_kw)
                    print(f"New client (flytekit) serializes the message to: {hx}")
                    result = _deserialize_go(old_go_bin, full_name, hx)
                    print(f"Old backend (flyte) deserializes the message to: {result}")
                except RuntimeError as e:
                    print(f"ERROR: {e}")
                    all_ok = False
                print()

            # ------ Test 2: Old client (Python)  →  New backend (Go) ------
            print(f"Test 2: Old client (flytekit) and new backend (flyte)\n")
            seen_old: set = set()
            for tc in cases:
                if not tc.old_kw:
                    continue
                key = frozenset(tc.old_kw.items())
                if key in seen_old:
                    continue
                seen_old.add(key)

                print(f"message = {_fmt_ctor(cls, tc.old_kw)}\n")
                try:
                    hx = _serialize(old_pb, py_mod, cls, tc.old_kw)
                    print(f"Old client (flytekit) serializes the message to: {hx}")
                    result = _deserialize_go(new_go_bin, full_name, hx)
                    print(f"New backend (flyte) deserializes the message to: {result}")
                except RuntimeError as e:
                    print(f"ERROR: {e}")
                    all_ok = False
                print()

        # ---- Summary ------------------------------------------------------
        if all_ok:
            print("Version-skew compatibility: PASS (but please manually verify the results)")
        else:
            print("Version-skew compatibility: FAIL — see errors above")
            sys.exit(1)


if __name__ == "__main__":
    main()
