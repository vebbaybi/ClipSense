"""Inventory is evidence, not release approval. Scanner failures always block."""
import datetime
import json
import sys
from pathlib import Path

kind, filename = sys.argv[1:3]
data = json.loads(Path(filename).read_text())
findings = []
if kind == 'npm':
    if 'metadata' not in data:
        raise SystemExit('npm scan incomplete')
    findings = [v for v in data.get('vulnerabilities', {}).values() if v['severity'] in ('high', 'critical')]
elif kind == 'python':
    if 'dependencies' not in data:
        raise SystemExit('Python scan incomplete')
    if any(d.get('skip_reason') for d in data['dependencies']):
        raise SystemExit('Python scan skipped installed dependencies')
    findings = [v for d in data['dependencies'] for v in d.get('vulns', [])]
elif kind == 'container':
    if 'SchemaVersion' not in data:
        raise SystemExit('container scan incomplete')
    if any(result.get('Secrets') for result in data.get('Results', [])):
        for result in data.get('Results', []):
            result['Secrets'] = [{k:v for k,v in secret.items() if k in ('RuleID', 'Category', 'Severity', 'Title', 'StartLine', 'EndLine')} for secret in result.get('Secrets', [])]
        Path(filename).write_text(json.dumps(data))
        raise SystemExit('Image secret finding requires review; release and integration blocked')
    findings = [v for result in data.get('Results', []) for v in result.get('Vulnerabilities', []) if v.get('Severity') in ('HIGH', 'CRITICAL')]
else:
    raise SystemExit('unknown scanner')
print(json.dumps({'scanner':kind, 'release_blocking_findings':len(findings), 'release_accepted':not findings}))
# Explicit bootstrap inventory window, not indefinite suppression or risk acceptance.
# Foundation can be reviewed; no release is authorized during this debt window.
deadline = datetime.date(2026, 10, 6)
if findings and ('--release' in sys.argv or datetime.date.today() >= deadline):
    raise SystemExit('Security debt blocks release, and integration after 2026-10-06')
