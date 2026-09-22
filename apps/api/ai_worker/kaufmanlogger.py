"""Allowlisted operational events over standard logging; never serialize payloads."""
import contextlib
import contextvars
import functools
import json
import logging
import os
import re
import shutil
import time
import uuid
from datetime import datetime, timezone

from prometheus_client import Counter, Gauge, Histogram, start_http_server

CONTEXT = contextvars.ContextVar('diagnostics', default={})
EVENTS = frozenset('service.started service.failed job.claimed job.invalid job.failed job.completed processing.started processing.completed processing.stage.started processing.stage.completed processing.stage.failed dependency.degraded dependency.recovered upload.rejected upload.cleanup.failed'.split())
STAGES = frozenset('validation ffmpeg transcription embedding qdrant persistence classification models processing'.split())
event_count = Counter('clipsense_events_total', 'Allowlisted operational events', ['event'])
stage_duration = Histogram('clipsense_stage_duration_seconds', 'Major worker stages', ['stage', 'outcome'], buckets=(.1, 1, 5, 30, 60, 300, 900, 3600))
job_duration = Histogram('clipsense_job_duration_seconds', 'Claimed job processing duration', buckets=(1, 5, 30, 60, 300, 900, 3600))
disk_available = Gauge('clipsense_storage_available_bytes', 'Free bytes on processing filesystem')
disk_total = Gauge('clipsense_storage_total_bytes', 'Capacity of processing filesystem')
for _event in EVENTS:
    event_count.labels(_event)


def valid_id(value):
    if not isinstance(value, str) or len(value) != 36:
        return None
    try:
        return value if str(uuid.UUID(value)) == value else None
    except ValueError:
        return None


def job_context(data):
    return {key: value for key, value in {
        'batch_id': valid_id(data.get('id')),
        'request_id': valid_id(data.get('request_id')),
        'correlation_id': valid_id(data.get('correlation_id')) or str(uuid.uuid4()),
    }.items() if value}


class KaufmanFormatter(logging.Formatter):
    def format(self, record):
        name = record.msg if record.msg in EVENTS else 'service.failed'
        version = os.getenv('BUILD_SHA', '')
        environment = os.getenv('APP_ENV', 'production')
        output = dict(timestamp=datetime.now(timezone.utc).isoformat(), level=record.levelname,
                      service='worker', environment=environment if environment in ('development', 'test', 'production') else 'production',
                      version=version if re.fullmatch('[a-f0-9]{40}', version) else 'unknown', event=name)
        for key, value in CONTEXT.get().items():
            if key in ('batch_id', 'request_id', 'correlation_id') and valid_id(value):
                output[key] = value
        fields = getattr(record, 'safe_fields', {})
        if fields.get('stage') in STAGES:
            output['stage'] = fields['stage']
        if fields.get('dependency') in ('redis', 'database', 'qdrant'):
            output['dependency'] = fields['dependency']
        for key in ('duration_ms', 'retry_count'):
            value = fields.get(key)
            if type(value) in (int, float) and 0 <= value <= 10**12:
                output[key] = value
        # No record.args, exception text/stack, arbitrary extras, or library payloads.
        return json.dumps(output, allow_nan=False)


logger = logging.getLogger('clipsense.kaufman')
logger.setLevel(logging.INFO)
logger.propagate = False
if not logger.handlers:
    handler = logging.StreamHandler()
    handler.setFormatter(KaufmanFormatter())
    logger.addHandler(handler)


def event(name, **fields):
    if name not in EVENTS:
        return
    level = logging.ERROR if name.endswith('.failed') else logging.INFO
    logger.log(level, name, extra={'safe_fields': fields})
    event_count.labels(name).inc()


@contextlib.contextmanager
def stage(name):
    if name not in STAGES:
        raise ValueError('unknown processing stage')
    start = time.monotonic()
    outcome = 'failed'
    event('processing.stage.started', stage=name)
    try:
        yield
        outcome = 'completed'
    finally:
        elapsed = time.monotonic() - start
        stage_duration.labels(name, outcome).observe(elapsed)
        event('processing.stage.' + outcome, stage=name, duration_ms=int(elapsed * 1000))


def measured(name):
    def decorate(function):
        @functools.wraps(function)
        def call(*args, **kwargs):
            with stage(name):
                return function(*args, **kwargs)
        return call
    return decorate


def safety_event(message):
    # Existing safety helper emits fixed strings; unknown input is never forwarded.
    if message in ('event=archive_cleanup_failed', 'event=upload_cleanup_failed'):
        event('upload.cleanup.failed')
    elif message in ('event=archive_rejected', 'event=upload_validation_failed'):
        event('upload.rejected')


def start_metrics():
    start_http_server(9092, addr=os.getenv('METRICS_HOST', '127.0.0.1'))
    def available():
        return shutil.disk_usage(os.getenv('PROCESS_DIR', '/tmp')).free
    def total():
        return shutil.disk_usage(os.getenv('PROCESS_DIR', '/tmp')).total
    disk_available.set_function(available)
    disk_total.set_function(total)
