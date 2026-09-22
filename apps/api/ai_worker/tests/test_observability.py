import io
import json
import logging
import unittest
from prometheus_client import generate_latest
from kaufmanlogger import CONTEXT, KaufmanFormatter, event, job_context, logger, stage


class ObservabilityTests(unittest.TestCase):
    def test_context_and_redaction(self):
        identifier = 'f2c726a4-46ed-4ae1-9481-b3616b66cbef'
        context = job_context({'id':identifier, 'request_id':'bad\nsecret', 'correlation_id':identifier, 'token':'secret'})
        token = CONTEXT.set(context)
        try:
            record = logging.LogRecord('worker', logging.ERROR, '', 0, 'job.failed', (), None)
            record.safe_fields = {'password':'secret', 'error':'redis://secret', 'stage':'secret', 'transcript':'private'}
            encoded = KaufmanFormatter().format(record)
            data = json.loads(encoded)
            self.assertEqual(data['correlation_id'], identifier)
            self.assertEqual(data['batch_id'], identifier)
            self.assertNotIn('request_id', data)
            self.assertNotIn('secret', encoded)
            self.assertNotIn('private', encoded)
        finally:
            CONTEXT.reset(token)

    def test_stage_failure_and_metrics_exclude_identifiers(self):
        with self.assertRaises(ValueError):
            with stage('ffmpeg'):
                raise ValueError('sensitive filename')
        metrics = generate_latest().decode()
        self.assertIn('clipsense_stage_duration_seconds_count{outcome="failed",stage="ffmpeg"}', metrics)
        self.assertNotIn('sensitive filename', metrics)
        self.assertNotIn('batch_id=', metrics)
        self.assertNotIn('request_id=', metrics)

    def test_unrecognized_fields_and_event_are_dropped(self):
        stream = io.StringIO()
        handler = logging.StreamHandler(stream)
        handler.setFormatter(KaufmanFormatter())
        logger.addHandler(handler)
        try:
            event('attacker.event', password='secret')
            self.assertEqual(stream.getvalue(), '')
            event('job.failed', exception='secret', duration_ms=float('nan'))
            self.assertNotIn('secret', stream.getvalue())
            json.loads(stream.getvalue())
        finally:
            logger.removeHandler(handler)
