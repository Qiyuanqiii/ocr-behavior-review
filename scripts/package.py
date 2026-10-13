# SPDX-License-Identifier: Apache-2.0
# Copyright 2026 Qiyuanqiii

"""Build the native plugin archive and its matching marketplace catalog."""

import hashlib
import json
from pathlib import Path
import zipfile


def build(root: Path, output: Path) -> tuple[Path, Path]:
    package = root / "plugin"
    manifest = json.loads((package / "ocr-plugin.json").read_text(encoding="utf-8"))
    version = manifest["version"]
    output.mkdir(parents=True, exist_ok=True)
    archive = output / f"behavior-review-{version}.zip"
    with zipfile.ZipFile(archive, "w", compression=zipfile.ZIP_DEFLATED) as bundle:
        for name in ("ocr-plugin.json", "behavior-review.md", "LICENSE", "NOTICE"):
            source = package / name
            data = source.read_bytes()
            info = zipfile.ZipInfo(name, (2026, 1, 1, 0, 0, 0))
            info.compress_type = zipfile.ZIP_DEFLATED
            info.external_attr = 0o100644 << 16
            bundle.writestr(info, data)
    entry = {key: manifest[key] for key in ("id", "name", "version", "description")}
    entry["source"] = f"https://github.com/Qiyuanqiii/ocr-behavior-review/releases/download/v{version}/{archive.name}"
    entry["sha256"] = hashlib.sha256(archive.read_bytes()).hexdigest()
    catalog = output / "marketplace.json"
    catalog.write_text(json.dumps({"schema_version": 1, "name": "Behavior Review", "plugins": [entry]}, indent=2) + "\n", encoding="utf-8", newline="\n")
    return archive, catalog


if __name__ == "__main__":
    project = Path(__file__).resolve().parents[1]
    for artifact in build(project, project / "dist"):
        print(artifact)
