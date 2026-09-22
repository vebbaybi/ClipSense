import json
import shutil
import stat
import struct
import subprocess
import zipfile
from pathlib import Path, PurePosixPath

MAX_FILES_PER_BATCH = 200
MAX_ENTRY_BYTES = 256 * 1024 * 1024
MAX_TOTAL_EXTRACTED_BYTES = 1024 * 1024 * 1024
MAX_COMPRESSION_RATIO = 100
ALLOWED_VIDEO_EXTENSIONS = {'.mp4', '.mov', '.mkv', '.webm', '.avi'}


def safe_zip_member_path(name: str) -> PurePosixPath:
    normalized = name.replace('\\', '/').removesuffix('/')
    parts = normalized.split('/')
    if normalized.startswith('/') or ':' in normalized:
        raise ValueError('archive absolute path')
    if len(name) > 512 or len(parts) > 8:
        raise ValueError('archive unsafe path')
    for part in parts:
        base = part.split('.')[0].upper()
        if (part in ('', '.', '..') or part.rstrip(' .') != part
                or any(ord(c) < 32 or ord(c) > 126 or c in ':*?"<>|' for c in part)
                or base in {'CON', 'PRN', 'AUX', 'NUL'}
                or (len(base) == 4 and base[:3] in {'COM', 'LPT'} and base[3] in '123456789')):
            raise ValueError('archive unsafe path')
    return PurePosixPath(*parts)


def check_directory(zip_path):
    # Bound zipfile's central-directory allocation before constructing ZipInfo objects.
    with zip_path.open('rb') as source:
        source.seek(0, 2)
        size = source.tell()
        source.seek(max(0, size - 65557))
        tail = source.read(65557)
    pos = tail.rfind(b'PK\x05\x06')
    if pos < 0 or len(tail) - pos < 22:
        raise ValueError('invalid archive')
    _, disk, cd_disk, local_count, count, cd_size, offset, comment = struct.unpack('<4s4H2IH', tail[pos:pos+22])
    if (disk or cd_disk or local_count != count or not 0 < count <= MAX_FILES_PER_BATCH
            or cd_size > 128 * 1024 or len(tail) - pos != 22 + comment
            or offset + cd_size != size - 22 - comment):
        raise ValueError('archive empty or too many files or invalid directory')


def extract_zip(zip_path: Path, dest: Path, logger=None):
    check_directory(zip_path)
    # Own only a newly created workspace; never adopt or delete an existing one.
    dest.mkdir(parents=False, exist_ok=False)
    root = dest.resolve()
    try:
        with zipfile.ZipFile(zip_path, 'r') as zf:
            entries = zf.infolist()
            if not 0 < len(entries) <= MAX_FILES_PER_BATCH:
                raise ValueError('archive too many files')
            paths = {}
            total = 0
            videos = 0
            for info in entries:
                rel = safe_zip_member_path(info.filename)
                key = str(rel).lower()
                kind = stat.S_IFMT(info.external_attr >> 16)
                if (key in paths or kind not in (0, stat.S_IFREG, stat.S_IFDIR)
                        or info.flag_bits & 1 or info.compress_type not in (zipfile.ZIP_STORED, zipfile.ZIP_DEFLATED)):
                    raise ValueError('unsafe archive member')
                paths[key] = info.is_dir()
                total += info.file_size
                if (info.file_size > MAX_ENTRY_BYTES or total > MAX_TOTAL_EXTRACTED_BYTES
                        or info.file_size > MAX_COMPRESSION_RATIO * max(1, info.compress_size)):
                    raise ValueError('archive exceeds expanded size or ratio limit')
                if not info.is_dir():
                    if rel.suffix.lower() not in ALLOWED_VIDEO_EXTENSIONS or info.file_size == 0:
                        raise ValueError('unsupported or empty media member')
                    videos += 1
            for key in paths:
                for parent in PurePosixPath(key).parents:
                    if str(parent) in paths and not paths[str(parent)]:
                        raise ValueError('conflicting archive paths')
            if not videos:
                raise ValueError('archive contains no supported media')
            copied_total = 0
            for info in entries:
                rel = safe_zip_member_path(info.filename)
                target = (root / Path(*rel.parts)).resolve()
                if not target.is_relative_to(root):
                    raise ValueError('archive unsafe path')
                if info.is_dir():
                    target.mkdir(parents=True, exist_ok=True)
                    continue
                target.parent.mkdir(parents=True, exist_ok=True)
                with zf.open(info) as source, target.open('xb') as output:
                    copied = copy_limited(source, output, min(MAX_ENTRY_BYTES, MAX_TOTAL_EXTRACTED_BYTES - copied_total))
                if copied != info.file_size:
                    raise ValueError('invalid archive size')
                copied_total += copied
    except Exception:
        try:
            shutil.rmtree(root)
        except OSError:
            if logger:
                logger('event=archive_cleanup_failed')
            raise
        if logger:
            logger('event=archive_rejected')
        raise


def copy_limited(src, dst, max_bytes: int) -> int:
    copied = 0
    while True:
        chunk = src.read(min(1024 * 1024, max(1, max_bytes - copied + 1)))
        if not chunk:
            return copied
        copied += len(chunk)
        if copied > max_bytes:
            raise ValueError('archive exceeds expanded size limit')
        dst.write(chunk)


def validate_media(video: Path):
    if video.suffix.lower() not in ALLOWED_VIDEO_EXTENSIONS or video.is_symlink():
        raise ValueError('unsupported media')
    try:
        result = subprocess.run([
            'ffprobe', '-v', 'error', '-protocol_whitelist', 'file',
            '-format_whitelist', 'mov,matroska,webm,avi',
            '-probesize', '5000000', '-analyzeduration', '5000000',
            '-select_streams', 'v:0', '-show_entries', 'stream=codec_type:format=format_name',
            '-of', 'json', str(video.resolve())
        ], check=True, stdout=subprocess.PIPE, stderr=subprocess.DEVNULL, timeout=5)
        if len(result.stdout) > 16384:
            raise ValueError('invalid media')
        data = json.loads(result.stdout)
        formats = set(data.get('format', {}).get('format_name', '').split(','))
        expected = {'.mp4': 'mov', '.mov': 'mov', '.mkv': 'matroska', '.webm': 'webm', '.avi': 'avi'}
        if expected[video.suffix.lower()] not in formats or not any(s.get('codec_type') == 'video' for s in data.get('streams', [])):
            raise ValueError('invalid media')
    except (OSError, subprocess.SubprocessError, ValueError, KeyError, TypeError) as exc:
        raise ValueError('unsupported or malformed media') from exc


def media_files(root: Path):
    for p in root.rglob('*'):
        if p.is_file() and not p.is_symlink() and p.suffix.lower() in ALLOWED_VIDEO_EXTENSIONS:
            yield p


def prepare_upload(zip_path: Path, workdir: Path, logger):
    owned = not workdir.exists() and not workdir.is_symlink()
    try:
        extract_zip(zip_path, workdir, logger=logger)
        videos = list(media_files(workdir))
        for video in videos:
            validate_media(video)
        return videos
    except Exception:
        if owned and workdir.exists():
            try:
                shutil.rmtree(workdir)
            except OSError:
                logger('event=archive_cleanup_failed')
        try:
            zip_path.unlink(missing_ok=True)
        except OSError:
            logger('event=upload_cleanup_failed')
        logger('event=upload_validation_failed')
        raise ValueError('upload validation failed') from None
