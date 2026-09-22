"""Run inside the worker to exercise actual bounded FFprobe, not analysis."""
import subprocess
import tempfile
import zipfile
from pathlib import Path
from zip_safety import prepare_upload, validate_media

with tempfile.TemporaryDirectory() as directory:
    root = Path(directory)
    video = root / 'valid.mp4'
    subprocess.run(['ffmpeg', '-v', 'error', '-f', 'lavfi', '-i', 'color=size=32x32:rate=1',
                    '-t', '1', '-c:v', 'mpeg4', str(video)], check=True, timeout=10)
    validate_media(video)
    archive = root / 'bad.zip'
    with zipfile.ZipFile(archive, 'w') as zf:
        zf.writestr('fake.mp4', b'not a video')
    try:
        prepare_upload(archive, root / 'workspace', print)
    except ValueError:
        pass
    else:
        raise AssertionError('fake media accepted')
    assert not archive.exists() and not (root / 'workspace').exists()
print('PASS: real video probe accepted; fake extension rejected; ZIP and extraction workspace cleaned')
