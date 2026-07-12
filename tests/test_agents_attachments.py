import importlib.machinery
import importlib.util
import json
import os
import subprocess
import sys
import tempfile
import unittest
from pathlib import Path


REPO_ROOT = Path(__file__).resolve().parents[1]
SCRIPT = REPO_ROOT / ".scripts" / "agents-attachments"


def load_helper_module():
    loader = importlib.machinery.SourceFileLoader("agents_attachments", str(SCRIPT))
    spec = importlib.util.spec_from_loader(loader.name, loader)
    module = importlib.util.module_from_spec(spec)
    loader.exec_module(module)
    return module


class AgentsAttachmentsImageStageTests(unittest.TestCase):
    def run_helper(self, args, env=None):
        completed = subprocess.run(
            [sys.executable, str(SCRIPT), *args],
            text=True,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            env=env,
            check=False,
        )
        if completed.returncode != 0:
            self.fail(f"agents-attachments failed: {completed.stderr}\nstdout:\n{completed.stdout}")
        return completed

    def write_manifest(self, path, attachments):
        path.write_text(json.dumps({"attachments": attachments}, indent=2) + "\n", encoding="utf-8")

    def test_stage_images_accepts_paths_and_manifest_refs_with_redacted_mapping(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            explicit = root / "explicit.png"
            explicit.write_bytes(b"explicit-image")
            secret_image = root / "sim-8912345678901234567.png"
            secret_image.write_bytes(b"secret-image")
            manifest = root / "manifest.json"
            out_dir = root / "stage"
            self.write_manifest(
                manifest,
                [
                    {
                        "id": "photo-ref",
                        "name": secret_image.name,
                        "mime_type": "image/png",
                        "size_bytes": secret_image.stat().st_size,
                        "local_path": str(secret_image),
                    }
                ],
            )

            completed = self.run_helper(
                [
                    "stage-images",
                    "--manifest",
                    str(manifest),
                    "--out-dir",
                    str(out_dir),
                    str(explicit),
                    "photo-ref",
                ]
            )

            payload = json.loads(completed.stdout)
            mapping_path = Path(payload["mapping_path"])
            self.assertTrue(mapping_path.exists())
            self.assertEqual(len(payload["items"]), 2)
            self.assertEqual(payload["items"][0]["source"]["kind"], "path")
            self.assertEqual(payload["items"][1]["source"]["kind"], "manifest")
            self.assertEqual(explicit.read_bytes(), b"explicit-image")
            self.assertEqual(secret_image.read_bytes(), b"secret-image")

            mapping_text = mapping_path.read_text(encoding="utf-8")
            self.assertNotIn("8912345678901234567", mapping_text)
            self.assertIn("REDACTED_ICCID", mapping_text)
            for item in payload["items"]:
                self.assertTrue(Path(item["staged"]["path"]).exists())
                self.assertEqual(item["action"], "copied")
                self.assertTrue(item["source_read_only"])

    def test_stage_images_all_filters_manifest_to_images(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            image = root / "image.png"
            image.write_bytes(b"image")
            text = root / "notes.txt"
            text.write_text("not an image", encoding="utf-8")
            manifest = root / "manifest.json"
            self.write_manifest(
                manifest,
                [
                    {"id": "image", "name": image.name, "mime_type": "image/png", "local_path": str(image)},
                    {"id": "text", "name": text.name, "mime_type": "text/plain", "local_path": str(text)},
                ],
            )

            completed = self.run_helper(
                ["stage-images", "--manifest", str(manifest), "--out-dir", str(root / "stage"), "--all"]
            )

            payload = json.loads(completed.stdout)
            self.assertEqual(len(payload["items"]), 1)
            self.assertEqual(payload["items"][0]["source"]["manifest"]["id"], "image")

    def test_heic_converter_prefers_sips_on_macos(self):
        helper = load_helper_module()

        def fake_which(name):
            return { "sips": "/usr/bin/sips", "magick": "/opt/homebrew/bin/magick" }.get(name)

        self.assertEqual(helper.choose_heic_converter(which=fake_which, platform="darwin"), ("sips", "/usr/bin/sips"))

    def test_heic_converter_uses_imagemagick_fallback(self):
        helper = load_helper_module()

        def fake_which(name):
            return {"magick": "/usr/local/bin/magick"}.get(name)

        self.assertEqual(
            helper.choose_heic_converter(which=fake_which, platform="linux"),
            ("imagemagick", "/usr/local/bin/magick"),
        )

    def test_stage_images_heic_uses_detected_portable_fallback(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            bin_dir = root / "bin"
            bin_dir.mkdir()
            magick = bin_dir / "magick"
            magick.write_text("#!/bin/sh\ncp \"$1\" \"$2\"\n", encoding="utf-8")
            magick.chmod(0o755)
            heic = root / "sample.heic"
            heic.write_bytes(b"fake-heic")
            env = os.environ.copy()
            env["PATH"] = f"{bin_dir}{os.pathsep}{env.get('PATH', '')}"
            env["AGENTS_ATTACHMENTS_PLATFORM"] = "linux"

            completed = self.run_helper(["stage-images", "--out-dir", str(root / "stage"), str(heic)], env=env)

            payload = json.loads(completed.stdout)
            self.assertEqual(payload["items"][0]["action"], "normalized")
            self.assertEqual(payload["items"][0]["normalization"]["converter"], "imagemagick")
            self.assertEqual(Path(payload["items"][0]["staged"]["path"]).read_bytes(), b"fake-heic")
            self.assertEqual(heic.read_bytes(), b"fake-heic")


if __name__ == "__main__":
    unittest.main()
