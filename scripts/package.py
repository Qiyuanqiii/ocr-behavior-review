# SPDX-License-Identifier: Apache-2.0
# Copyright 2026 Qiyuanqiii

"""Build the native plugin archive and its matching marketplace catalog."""

import hashlib
import json
import os
from pathlib import Path
import subprocess
import zipfile


PLATFORMS = [(system, arch) for system in ("windows", "linux", "darwin") for arch in ("amd64", "arm64")]


def binary_name(system: str, arch: str) -> str:
    suffix = ".exe" if system == "windows" else ""
    return f"bin/{system}-{arch}/behavior-review{suffix}"


def compile_binaries(root: Path, output: Path) -> None:
    for system, arch in PLATFORMS:
        target = output / binary_name(system, arch)
        target.parent.mkdir(parents=True, exist_ok=True)
        env = {**os.environ, "GOOS": system, "GOARCH": arch, "CGO_ENABLED": "0"}
        subprocess.run([os.environ.get("GO", "go"), "build", "-trimpath", "-ldflags=-s -w -buildid=", "-o", str(target), "./cmd/behavior-review"], cwd=root, env=env, check=True)


def build(root: Path, output: Path, binaries: Path) -> tuple[Path, Path]:
    package = root / "plugin"
    manifest = json.loads((package / "ocr-plugin.json").read_text(encoding="utf-8"))
    version = manifest["version"]
    output.mkdir(parents=True, exist_ok=True)
    archive = output / f"behavior-review-{version}.zip"
    with zipfile.ZipFile(archive, "w", compression=zipfile.ZIP_DEFLATED) as bundle:
        names = ["ocr-plugin.json", "behavior-review.md", "LICENSE", "NOTICE"]
        names += [binary_name(system, arch) for system, arch in PLATFORMS]
        for name in names:
            executable = name.startswith("bin/")
            source = (binaries if executable else package) / name
            data = source.read_bytes()
            info = zipfile.ZipInfo(name, (2026, 1, 1, 0, 0, 0))
            info.compress_type = zipfile.ZIP_DEFLATED
            info.create_system = 3
            info.external_attr = (0o100755 if executable else 0o100644) << 16
            bundle.writestr(info, data)
    entry = {key: manifest[key] for key in ("id", "name", "version", "description")}
    entry["source"] = f"https://github.com/Qiyuanqiii/ocr-behavior-review/releases/download/v{version}/{archive.name}"
    entry["sha256"] = hashlib.sha256(archive.read_bytes()).hexdigest()
    catalog = output / "marketplace.json"
    catalog.write_text(json.dumps({"schema_version": 1, "name": "Behavior Review", "plugins": [entry]}, indent=2) + "\n", encoding="utf-8", newline="\n")
    return archive, catalog


if __name__ == "__main__":
    project = Path(__file__).resolve().parents[1]
    binaries = project / "dist" / "build"
    compile_binaries(project, binaries)
    for artifact in build(project, project / "dist", binaries):
        print(artifact)
