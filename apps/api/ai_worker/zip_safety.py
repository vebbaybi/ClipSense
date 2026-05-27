import zipfile
from pathlib import Path, PurePosixPath

MAX_FILES_PER_BATCH = 200
MAX_TOTAL_EXTRACTED_BYTES = 10 * 1024 * 1024 * 1024
ALLOWED_VIDEO_EXTENSIONS = {'.mp4', '.mov', '.mkv', '.webm', '.avi'}


def extract_zip(zip_path: Path, dest: Path, logger=None):
    dest.mkdir(parents=True, exist_ok=True)
    dest_root = dest.resolve()
    file_count = 0
    extracted_bytes = 0

    with zipfile.ZipFile(zip_path, 'r') as zf:
        for info in zf.infolist():
            if info.is_dir():
                continue

            file_count += 1
            if file_count > MAX_FILES_PER_BATCH:
                raise ValueError(f"zip has too many files; max is {MAX_FILES_PER_BATCH}")

            rel_path = safe_zip_member_path(info.filename)
            if rel_path.suffix.lower() not in ALLOWED_VIDEO_EXTENSIONS:
                if logger is not None:
                    logger(f"ignoring non-video zip member: {rel_path}")
                continue

            target = (dest_root / Path(*rel_path.parts)).resolve()
            if not target.is_relative_to(dest_root):
                raise ValueError(f"zip member escapes processing directory: {info.filename}")

            target.parent.mkdir(parents=True, exist_ok=True)
            remaining = MAX_TOTAL_EXTRACTED_BYTES - extracted_bytes
            try:
                with zf.open(info, 'r') as src, target.open('wb') as out:
                    copied = copy_limited(src, out, remaining)
            except Exception:
                if target.exists():
                    target.unlink()
                raise

            extracted_bytes += copied
            if extracted_bytes > MAX_TOTAL_EXTRACTED_BYTES:
                raise ValueError(f"zip contents exceed {MAX_TOTAL_EXTRACTED_BYTES} bytes")


def safe_zip_member_path(name: str) -> PurePosixPath:
    normalized = name.replace('\\', '/')
    rel_path = PurePosixPath(normalized)
    if rel_path.is_absolute():
        raise ValueError(f"zip member uses absolute path: {name}")
    if rel_path.parts and rel_path.parts[0].endswith(':'):
        raise ValueError(f"zip member uses absolute path: {name}")
    if any(part in ('', '.', '..') for part in rel_path.parts):
        raise ValueError(f"zip member uses unsafe path: {name}")
    return rel_path


def copy_limited(src, dst, max_bytes: int) -> int:
    if max_bytes < 0:
        raise ValueError(f"zip contents exceed {MAX_TOTAL_EXTRACTED_BYTES} bytes")
    copied = 0
    while True:
        chunk = src.read(1024 * 1024)
        if not chunk:
            break
        copied += len(chunk)
        if copied > max_bytes:
            raise ValueError(f"zip contents exceed {MAX_TOTAL_EXTRACTED_BYTES} bytes")
        dst.write(chunk)
    return copied


def media_files(root: Path):
    for p in root.rglob('*'):
        if p.is_file() and p.suffix.lower() in ALLOWED_VIDEO_EXTENSIONS:
            yield p
