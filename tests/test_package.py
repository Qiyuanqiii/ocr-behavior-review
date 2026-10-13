# SPDX-License-Identifier: Apache-2.0
# Copyright 2026 Qiyuanqiii

import hashlib
import json
from pathlib import Path
import tempfile
import unittest
import zipfile

from scripts.package import build


class PackageTest(unittest.TestCase):
    def test_archive_is_installable_and_catalog_matches(self):
        root = Path(__file__).resolve().parents[1]
        with tempfile.TemporaryDirectory() as temporary:
            archive, catalog = build(root, Path(temporary))
            first = archive.read_bytes()
            entry = json.loads(catalog.read_text())["plugins"][0]
            self.assertEqual(hashlib.sha256(first).hexdigest(), entry["sha256"])
            with zipfile.ZipFile(archive) as bundle:
                self.assertEqual(set(bundle.namelist()), {"ocr-plugin.json", "behavior-review.md", "LICENSE", "NOTICE"})
                manifest = json.loads(bundle.read("ocr-plugin.json"))
                self.assertEqual(entry["id"], manifest["id"])
                self.assertEqual(entry["version"], manifest["version"])
                self.assertEqual(manifest["review"]["engine"], "behavior-v1")
                policy = bundle.read(manifest["review"]["prompt"])
                self.assertTrue(policy.strip())
                self.assertNotIn(b"\r", policy)
            build(root, Path(temporary))
            self.assertEqual(first, archive.read_bytes())


if __name__ == "__main__":
    unittest.main()
