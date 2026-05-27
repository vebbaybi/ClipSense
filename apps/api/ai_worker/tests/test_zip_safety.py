import tempfile
import unittest
import zipfile
from pathlib import Path
from unittest import mock

import zip_safety


class ZipSafetyTests(unittest.TestCase):
    def make_zip(self, entries):
        temp_dir = tempfile.TemporaryDirectory()
        self.addCleanup(temp_dir.cleanup)
        zip_path = Path(temp_dir.name) / "batch.zip"
        with zipfile.ZipFile(zip_path, "w") as zf:
            for name, data in entries:
                zf.writestr(name, data)
        return zip_path

    def extract_to_temp(self, zip_path):
        temp_dir = tempfile.TemporaryDirectory()
        self.addCleanup(temp_dir.cleanup)
        dest = Path(temp_dir.name) / "processing"
        zip_safety.extract_zip(zip_path, dest)
        return dest

    def test_supported_video_extension_is_extracted(self):
        zip_path = self.make_zip([("clip.mp4", b"video")])

        dest = self.extract_to_temp(zip_path)

        self.assertEqual((dest / "clip.mp4").read_bytes(), b"video")

    def test_non_video_entries_are_ignored(self):
        zip_path = self.make_zip([("notes.txt", b"not video")])

        dest = self.extract_to_temp(zip_path)

        self.assertFalse(any(dest.rglob("*")))

    def test_path_traversal_is_rejected_and_does_not_escape(self):
        zip_path = self.make_zip([("../evil.mp4", b"video")])
        with tempfile.TemporaryDirectory() as temp_dir:
            root = Path(temp_dir)
            dest = root / "processing"

            with self.assertRaisesRegex(ValueError, "unsafe path"):
                zip_safety.extract_zip(zip_path, dest)

            self.assertFalse((root / "evil.mp4").exists())

    def test_absolute_posix_path_is_rejected(self):
        zip_path = self.make_zip([("/evil.mp4", b"video")])

        with self.assertRaisesRegex(ValueError, "absolute path"):
            self.extract_to_temp(zip_path)

    def test_absolute_windows_path_is_rejected(self):
        zip_path = self.make_zip([("C:\\evil.mp4", b"video")])

        with self.assertRaisesRegex(ValueError, "absolute path"):
            self.extract_to_temp(zip_path)

    def test_excessive_file_count_is_rejected(self):
        entries = [(f"clip-{i}.mp4", b"v") for i in range(zip_safety.MAX_FILES_PER_BATCH + 1)]
        zip_path = self.make_zip(entries)

        with self.assertRaisesRegex(ValueError, "too many files"):
            self.extract_to_temp(zip_path)

    def test_excessive_decompressed_size_is_rejected(self):
        zip_path = self.make_zip([("clip.mp4", b"123456")])

        with mock.patch.object(zip_safety, "MAX_TOTAL_EXTRACTED_BYTES", 5):
            with self.assertRaisesRegex(ValueError, "exceed"):
                self.extract_to_temp(zip_path)


if __name__ == "__main__":
    unittest.main()
