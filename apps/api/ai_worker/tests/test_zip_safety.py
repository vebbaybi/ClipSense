import tempfile
import unittest
import zipfile
import stat
import subprocess
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

    def test_non_video_entries_are_rejected(self):
        zip_path = self.make_zip([("notes.txt", b"not video")])

        with self.assertRaisesRegex(ValueError, 'unsupported'):
            self.extract_to_temp(zip_path)

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

    def test_unsafe_and_conflicting_paths(self):
        for name in ['..\\evil.mp4', 'C:evil.mp4', 'a/./clip.mp4', 'a//clip.mp4',
                     'CON.mp4', 'a/' * 9 + 'clip.mp4', 'name. /clip.mp4']:
            with self.subTest(name=name), self.assertRaises(ValueError):
                self.extract_to_temp(self.make_zip([(name, b'video')]))
        for entries in [[('clip.mp4', b'a'), ('CLIP.mp4', b'b')],
                        [('clip.mp4', b'a'), ('clip.mp4/other.mp4', b'b')]]:
            with self.assertRaises(ValueError):
                self.extract_to_temp(self.make_zip(entries))

    def test_symlink_empty_and_per_entry_limit(self):
        info = zipfile.ZipInfo('link.mp4')
        info.create_system = 3
        info.external_attr = (stat.S_IFLNK | 0o777) << 16
        with self.assertRaises(ValueError):
            self.extract_to_temp(self.make_zip([(info, b'target')]))
        with self.assertRaises(ValueError):
            self.extract_to_temp(self.make_zip([]))
        with mock.patch.object(zip_safety, 'MAX_ENTRY_BYTES', 2), self.assertRaises(ValueError):
            self.extract_to_temp(self.make_zip([('clip.mp4', b'video')]))

    def test_ratio_and_crc_rejection_cleanup(self):
        archive = self.make_zip([])
        with zipfile.ZipFile(archive, 'w', compression=zipfile.ZIP_DEFLATED) as zf:
            zf.writestr('clip.mp4', b'0' * 100000)
        with self.assertRaises(ValueError):
            self.extract_to_temp(archive)
        archive = self.make_zip([('first.mp4', b'video'), ('second.mp4', b'payload')])
        archive.write_bytes(archive.read_bytes().replace(b'payload', b'corrupt'))
        dest = archive.parent / 'extracted'
        with self.assertRaises(zipfile.BadZipFile):
            zip_safety.extract_zip(archive, dest)
        self.assertFalse(dest.exists())

    def test_existing_workspace_is_never_removed(self):
        archive = self.make_zip([('clip.mp4', b'video')])
        dest = archive.parent / 'owned-by-other'
        dest.mkdir()
        sentinel = dest / 'keep'
        sentinel.write_bytes(b'keep')
        with self.assertRaises(ValueError):
            zip_safety.prepare_upload(archive, dest, lambda msg: None)
        self.assertEqual(sentinel.read_bytes(), b'keep')

    def test_media_probe_and_rejection_cleanup(self):
        archive = self.make_zip([('clip.mp4', b'not-media')])
        dest = archive.parent / 'probe'
        with mock.patch.object(zip_safety.subprocess, 'run', side_effect=subprocess.TimeoutExpired('ffprobe', 5)):
            with self.assertRaisesRegex(ValueError, 'upload validation failed'):
                zip_safety.prepare_upload(archive, dest, lambda msg: None)
        self.assertFalse(dest.exists())
        self.assertFalse(archive.exists())
        result = subprocess.CompletedProcess([], 0, b'{"streams":[{"codec_type":"video"}],"format":{"format_name":"mov,mp4"}}')
        with mock.patch.object(zip_safety.subprocess, 'run', return_value=result) as run:
            zip_safety.validate_media(Path('clip.mp4'))
            self.assertEqual(run.call_args.kwargs['timeout'], 5)
            self.assertIn('-protocol_whitelist', run.call_args.args[0])
        result.stdout = b'{"streams":[],"format":{"format_name":"mov"}}'
        with mock.patch.object(zip_safety.subprocess, 'run', return_value=result), self.assertRaises(ValueError):
            zip_safety.validate_media(Path('clip.mp4'))


if __name__ == "__main__":
    unittest.main()
