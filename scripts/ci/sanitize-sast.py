"""Retain rule/location/severity evidence without source excerpts or captures."""
import json
import sys
from pathlib import Path

path = Path(sys.argv[1])
report = json.loads(path.read_text())
for result in report.get('results', []):
    extra = result.get('extra', {})
    result['extra'] = {key:value for key,value in extra.items() if key in ('message', 'severity', 'fingerprint', 'metadata')}
path.write_text(json.dumps(report))
