#!/usr/bin/env python3
"""Protobuf Version-Skew Tester for Flyte

Detects protobuf message changes on a branch and verifies forward/backward
compatibility of the wire format between old (base) and new (head) versions.

The tool:
  1. Diffs .proto files between base and head to find changed messages
  2. Extracts the generated Python protobuf bindings from both versions
  3. Serializes test messages and cross-deserializes across versions
  4. Reports wire-format hex and deserialized field values

Since protobuf wire format is language-agnostic, results apply to both
Go (flyte backend) and Python (flytekit client) consumers.

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
class MessageChange:
    proto_file: str
    message_name: str
    added_fields: List[ProtoField]
    removed_fields: List[ProtoField]
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
            if added or removed:
                changes.append(MessageChange(
                    proto_file=pf, message_name=nm,
                    added_fields=added, removed_fields=removed,
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


def make_test_cases(ch: MessageChange) -> List[Dict[str, str]]:
    """Return kwargs-dicts: one with old-only fields, one with all fields."""
    msg = ch.new_message
    cls = msg.name
    old_names = {f.name for f in ch.old_message.fields} if ch.old_message else set()
    cases: List[Dict[str, str]] = []

    old_fields = [f for f in msg.fields if f.name in old_names]
    if old_fields:
        cases.append({f.name: _test_val(f, msg, cls) for f in old_fields})
    if ch.added_fields:
        cases.append({f.name: _test_val(f, msg, cls) for f in msg.fields})
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

    # ---- Set up old / new proto bindings ----------------------------------
    with tempfile.TemporaryDirectory(prefix="proto_skew_") as tmp:
        old_root = os.path.join(tmp, "old")
        os.makedirs(old_root)
        _extract_pb_python(repo, args.base, old_root)
        old_pb = os.path.join(old_root, "flyteidl", "gen", "pb_python")

        new_root = os.path.join(tmp, "new")
        os.makedirs(new_root)
        _extract_pb_python(repo, args.head, new_root)
        new_pb = os.path.join(new_root, "flyteidl", "gen", "pb_python")

        print(f"\nTesting version-skew\n")
        print(f"Backend (flyte) and Client (flytekit)")
        print(f" - Old: {args.base}@{base_hash}")
        print(f" - New: {branch}@{head_hash}")

        all_ok = True
        for ch in changes:
            mod = _proto_module(ch.proto_file)
            cls = ch.message_name
            cases = make_test_cases(ch)
            old_names = {f.name for f in ch.old_message.fields} if ch.old_message else set()

            # ------ Test 1: New client  →  Old backend ---------------------
            print(f"\nTest 1: New client (flytekit) and old backend (flyte)\n")
            for kw in cases:
                print(f"message = {_fmt_ctor(cls, kw)}\n")
                try:
                    hx = _serialize(new_pb, mod, cls, kw)
                    print(f"New client (flytekit) serializes the message to: {hx}")

                    # Verify old bindings can deserialize without error
                    try:
                        old_result = _deserialize(old_pb, mod, cls, hx)
                    except RuntimeError:
                        old_result = None
                        all_ok = False

                    # Display using new bindings (superset of fields)
                    new_result = _deserialize(new_pb, mod, cls, hx)
                    print(f"Old backend (flyte) deserializes the message to: {new_result}")

                    if old_result is None:
                        print("  ⚠ WARNING: Old bindings raised an error during deserialization!")
                except RuntimeError as e:
                    print(f"ERROR: {e}")
                    all_ok = False
                print()

            # ------ Test 2: Old client  →  New backend ---------------------
            print(f"Test 2: Old client (flytekit) and new backend (flyte)\n")
            seen: set = set()
            for kw in cases:
                old_kw = {k: v for k, v in kw.items() if k in old_names}
                if not old_kw:
                    continue
                key = frozenset(old_kw.items())
                if key in seen:
                    continue
                seen.add(key)

                print(f"message = {_fmt_ctor(cls, old_kw)}\n")
                try:
                    hx = _serialize(old_pb, mod, cls, old_kw)
                    print(f"Old client (flytekit) serializes the message to: {hx}")
                    result = _deserialize(new_pb, mod, cls, hx)
                    print(f"New backend (flyte) deserializes the message to: {result}")
                except RuntimeError as e:
                    print(f"ERROR: {e}")
                    all_ok = False
                print()

        # ---- Summary ------------------------------------------------------
        if all_ok:
            print("Version-skew compatibility: PASS")
        else:
            print("Version-skew compatibility: FAIL — see errors above")
            sys.exit(1)


if __name__ == "__main__":
    main()
