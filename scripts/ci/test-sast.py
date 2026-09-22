"""Prove the exact production rules reject known unsafe code in each ecosystem."""
import json
import subprocess

result = subprocess.run(['semgrep', 'scan', '--metrics=off', '--config', '.semgrep.yml', '--error', '--json', 'scripts/ci/sast-fixtures'], capture_output=True, text=True)
report = json.loads(result.stdout)
found = {finding['check_id'].split('.')[-1] for finding in report.get('results', [])}
expected = {'python-dynamic-execution', 'javascript-dynamic-execution', 'go-disabled-tls-validation'}
assert result.returncode == 1 and not report.get('errors') and expected <= found, (result.returncode, found)
print('PASS: production SAST rules reject unsafe Go, Python and TypeScript fixtures')
