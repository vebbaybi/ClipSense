import os
import time
import sqlite3
import subprocess
import uuid
import json
import shutil
from datetime import datetime, timezone
from pathlib import Path

import redis
import psycopg
from sentence_transformers import SentenceTransformer
from sklearn.cluster import KMeans
import whisper
from qdrant_client import QdrantClient
from qdrant_client.http import models as qmodels

from zip_safety import extract_zip, media_files

DB_DRIVER = os.getenv('DB_DRIVER', 'postgres')
DB_PATH = Path(os.getenv('DB_PATH', 'data/clipsense.db'))
DATABASE_URL = os.getenv('DATABASE_URL', 'postgres://clipsense:clipsense@postgres:5432/clipsense')
UPLOAD_DIR = Path(os.getenv('UPLOAD_DIR', 'data/uploads'))
PROCESS_DIR = Path(os.getenv('PROCESS_DIR', 'data/processing'))
MODEL_NAME = os.getenv('WHISPER_MODEL', 'small')
REDIS_URL = os.getenv('REDIS_URL', 'redis://redis:6379')
QDRANT_HOST = os.getenv('QDRANT_HOST', 'qdrant')
QDRANT_PORT = int(os.getenv('QDRANT_PORT', '6333'))
QDRANT_COLLECTION = os.getenv('QDRANT_COLLECTION', 'clipsense_clips')

PROCESS_DIR.mkdir(parents=True, exist_ok=True)

model = whisper.load_model(MODEL_NAME)
embedder = SentenceTransformer('all-MiniLM-L6-v2')
rdb = redis.from_url(REDIS_URL)
qclient = QdrantClient(host=QDRANT_HOST, port=QDRANT_PORT)


def log(msg: str) -> None:
    print(f"[worker] {msg}")


def connect_db():
    if DB_DRIVER == 'sqlite':
        return sqlite3.connect(DB_PATH)
    return psycopg.connect(DATABASE_URL)


def ph(index: int) -> str:
    return '?' if DB_DRIVER == 'sqlite' else '%s'


def update_status(conn, batch_id: str, status: str):
    cur = conn.cursor()
    sql = f"UPDATE batches SET status = {ph(1)}, updated_at = {current_timestamp()} WHERE id = {ph(2)}"
    cur.execute(sql, (status, batch_id))
    conn.commit()


def current_timestamp():
    return 'datetime('"'"'now'"'"')' if DB_DRIVER == 'sqlite' else 'CURRENT_TIMESTAMP'


def extract_audio(video_path: Path, out_wav: Path):
    cmd = [
        'ffmpeg', '-y', '-i', str(video_path),
        '-vn', '-acodec', 'pcm_s16le', '-ar', '16000', '-ac', '1', str(out_wav)
    ]
    subprocess.run(cmd, check=True, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)


def transcribe(audio_path: Path) -> str:
    result = model.transcribe(str(audio_path), fp16=False)
    return result['text']


def summarize(text: str) -> str:
    return text[:280] + ('...' if len(text) > 280 else '')


def classify_mood(text: str) -> str:
    lowered = text.lower()
    if any(k in lowered for k in ['hype', 'excited', 'crazy', 'wow', 'insane']):
        return 'high-energy'
    if any(k in lowered for k in ['sad', 'quiet', 'soft']):
        return 'calm'
    return 'neutral'


def classify_role(text: str) -> str:
    if '?' in text[:120]:
        return 'setup'
    if any(k in text.lower() for k in ['then', 'after that', 'later']):
        return 'escalation'
    return 'beat'


def ensure_qdrant_collection(vector_size: int):
    try:
        qclient.get_collection(QDRANT_COLLECTION)
        return
    except Exception:
        pass
    qclient.recreate_collection(
        collection_name=QDRANT_COLLECTION,
        vectors_config=qmodels.VectorParams(size=vector_size, distance=qmodels.Distance.COSINE)
    )
    log(f"qdrant collection {QDRANT_COLLECTION} created")


def upsert_embeddings(clip_rows, embeddings):
    ensure_qdrant_collection(len(embeddings[0]))
    points = []
    for row, vec in zip(clip_rows, embeddings):
        clip_id, batch_id, filename, title, summary, transcript, mood, role, topic, duration = row
        points.append(
            qmodels.PointStruct(
                id=clip_id,
                vector=vec.tolist(),
                payload={
                    "batch_id": batch_id,
                    "filename": filename,
                    "title": title,
                    "summary": summary,
                    "mood": mood,
                    "role": role,
                    "topic": topic,
                    "duration_seconds": duration,
                }
            )
        )
    qclient.upsert(collection_name=QDRANT_COLLECTION, points=points)


def process_batch(batch_id: str, name: str, zip_path: Path):
    conn = connect_db()
    try:
        update_status(conn, batch_id, 'processing')
        workdir = PROCESS_DIR / batch_id
        if workdir.exists():
            shutil.rmtree(workdir)
        workdir.mkdir(parents=True, exist_ok=True)
        extract_zip(zip_path, workdir, logger=log)

        videos = list(media_files(workdir))
        if not videos:
            raise ValueError("zip contains no supported video files")

        clip_rows = []
        texts = []
        for video in videos:
            clip_id = str(uuid.uuid4())
            audio = workdir / f"{clip_id}.wav"
            try:
                extract_audio(video, audio)
                transcript = transcribe(audio)
            except Exception as e:
                log(f"failed {video.name}: {e}")
                transcript = ""
            summary = summarize(transcript)
            mood = classify_mood(transcript)
            role = classify_role(transcript)
            duration = duration_seconds(video)
            clip_rows.append((clip_id, batch_id, video.name, video.stem, summary, transcript, mood, role, 'general', duration))
            texts.append(transcript if transcript else video.name)

        if clip_rows:
            cur = conn.cursor()
            cols = "id, batch_id, filename, title, summary, transcript, mood, role, topic, duration_seconds, created_at"
            values_ph = ','.join([ph(i) for i in range(1, 12)])
            insert_sql = f"INSERT INTO clips ({cols}) VALUES ({values_ph})"
            now = datetime.now(timezone.utc)
            rows_with_time = [row + (now,) for row in clip_rows]
            cur.executemany(insert_sql, rows_with_time)
            conn.commit()

        if texts:
            embeddings = embedder.encode(texts)
            k = 1 if len(texts) == 1 else min(len(texts), max(2, len(texts)//2))
            km = KMeans(n_clusters=k, n_init=10)
            labels = km.fit_predict(embeddings)
            ordering = sorted(range(len(texts)), key=lambda i: (labels[i], clip_rows[i][9]))
            cur = conn.cursor()
            sid = str(uuid.uuid4())
            cur.execute(
                f"INSERT INTO storylines (id, batch_id, title, type, created_at) VALUES ({ph(1)}, {ph(2)}, {ph(3)}, {ph(4)}, {ph(5)})",
                (sid, batch_id, f"AI Sequence for {name}", 'ai_sequence', datetime.now(timezone.utc))
            )
            for pos, idx in enumerate(ordering):
                cur.execute(
                    f"INSERT INTO storyline_clips (storyline_id, clip_id, position) VALUES ({ph(1)}, {ph(2)}, {ph(3)})",
                    (sid, clip_rows[idx][0], pos)
                )
            conn.commit()
            ordered_rows = [clip_rows[idx] for idx in ordering]
            ordered_embeddings = [embeddings[idx] for idx in ordering]
            upsert_embeddings(ordered_rows, ordered_embeddings)

        cur = conn.cursor()
        cur.execute(
            f"UPDATE batches SET status='complete', clip_count=(SELECT count(*) FROM clips WHERE batch_id = {ph(1)}), updated_at = {current_timestamp()} WHERE id = {ph(2)}",
            (batch_id, batch_id)
        )
        conn.commit()
    finally:
        conn.close()
    log(f"batch {batch_id} complete")


def duration_seconds(video_path: Path) -> float:
    try:
        import ffmpeg
        probe = ffmpeg.probe(str(video_path))
        streams = [s for s in probe['streams'] if s['codec_type'] == 'video']
        if streams:
            return float(streams[0].get('duration', 0))
    except Exception:
        pass
    return 0.0


def main():
    while True:
        item = rdb.blpop('jobs:batch', timeout=0)
        if not item:
            time.sleep(1)
            continue
        _, payload = item
        try:
            data = json.loads(payload)
            batch_id = data['id']
            name = data.get('name', batch_id)
            zip_path = data['zip_path']
            if not isinstance(batch_id, str) or not isinstance(zip_path, str):
                raise ValueError("job id and zip_path must be strings")
        except Exception as e:
            log(f"invalid job payload: {e}")
            continue
        log(f"processing batch {batch_id}")
        try:
            process_batch(batch_id, name, Path(zip_path))
        except Exception as e:
            log(f"batch {batch_id} failed: {e}")
            try:
                conn = connect_db()
                update_status(conn, batch_id, 'failed')
            except Exception as status_error:
                log(f"failed to mark batch {batch_id} failed: {status_error}")
            finally:
                try:
                    conn.close()
                except Exception:
                    pass


if __name__ == "__main__":
    main()
