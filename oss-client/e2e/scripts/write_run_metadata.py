import os
import json
from pathlib import Path
from typing import Any, Dict

def write_run_metadata(entry: Dict[str, Any], filename: str = "e2e/scripts/fixtures.json"):
    """
    Append or update a run entry in a shared JSON file.
    """
    path = Path(filename).resolve()
    path.parent.mkdir(parents=True, exist_ok=True)

    # Load existing file if present
    if path.exists():
        try:
            with open(path, "r", encoding="utf-8") as f:
                data = json.load(f)
                if not isinstance(data, dict):
                    data = {}
        except Exception as e:
            data = {}
    else:
        print(f"[DEBUG] File does not exist yet at {path}")
        data = {}

    # Write or update
    data[entry["name"]] = entry

    with open(path, "w", encoding="utf-8") as f:
        json.dump(data, f, indent=2)
        f.write("\n")
