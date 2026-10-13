# SPDX-License-Identifier: Apache-2.0
# Copyright 2026 Qiyuanqiii

import hashlib
import json
from pathlib import Path
import tempfile
import unittest
import zipfile

from scripts.package import build, binary_name, PLATFORMS


class PackageTest(unittest.TestCase):
    def test_archive_is_installable_and_catalog_matches(self):
        root = Path(__file__).resolve().parents[1]
        with tempfile.TemporaryDirectory() as temporary:
            binaries = Path(temporary) / "binaries"
            for system, arch in PLATFORMS:
                target = binaries / binary_name(system, arch)
                target.parent.mkdir(parents=True, exist_ok=True)
                target.write_bytes(b"test executable")
            archive, catalog = build(root, Path(temporary), binaries)
            first = archive.read_bytes()
            entry = json.loads(catalog.read_text())["plugins"][0]
            self.assertEqual(hashlib.sha256(first).hexdigest(), entry["sha256"])
            with zipfile.ZipFile(archive) as bundle:
                self.assertEqual(set(bundle.namelist()), {"ocr-plugin.json", "behavior-review.md", "LICENSE", "NOTICE"} | {binary_name(*platform) for platform in PLATFORMS})
                manifest = json.loads(bundle.read("ocr-plugin.json"))
                self.assertEqual(entry["id"], manifest["id"])
                self.assertEqual(entry["version"], manifest["version"])
                self.assertEqual(manifest["review"]["protocol"], "stdio-v1")
                self.assertEqual(manifest["review"]["stages"], ["contracts", "counterexamples", "adjudication"])
                for platform in PLATFORMS:
                    self.assertEqual((bundle.getinfo(binary_name(*platform)).external_attr >> 16) & 0o777, 0o755)
                policy = bundle.read("behavior-review.md")
                self.assertTrue(policy.strip())
                self.assertNotIn(b"\r", policy)
            build(root, Path(temporary), binaries)
            self.assertEqual(first, archive.read_bytes())


if __name__ == "__main__":
    unittest.main()
