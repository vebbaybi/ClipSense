import json
import subprocess
import sys
import tempfile
import unittest
from pathlib import Path

POLICY = Path(__file__).with_name('security-policy.py')

class SecurityPolicyTests(unittest.TestCase):
    def run_policy(self, kind, report, *flags):
        with tempfile.TemporaryDirectory() as directory:
            path = Path(directory) / 'report.json'
            path.write_text(json.dumps(report))
            result = subprocess.run([sys.executable, str(POLICY), kind, str(path), *flags], capture_output=True, text=True)
            return result.returncode, path.read_text()

    def test_incomplete_scan_fails(self):
        self.assertNotEqual(self.run_policy('npm', {})[0], 0)
        self.assertNotEqual(self.run_policy('python', {'dependencies':[{'skip_reason':'not found'}]})[0], 0)

    def test_release_debt_blocks(self):
        self.assertNotEqual(self.run_policy('npm', {'metadata':{}, 'vulnerabilities':{'test':{'severity':'high'}}}, '--release')[0], 0)
        self.assertEqual(self.run_policy('npm', {'metadata':{}, 'vulnerabilities':{}}, '--release')[0], 0)

    def test_image_secrets_block_and_report_is_sanitized(self):
        code, report = self.run_policy('container', {'SchemaVersion':2, 'Results':[{'Secrets':[{'RuleID':'fixture', 'Match':'sensitive-sentinel', 'Code':{'Lines':['sensitive-sentinel']}}]}]})
        self.assertNotEqual(code, 0)
        self.assertNotIn('sensitive-sentinel', report)

if __name__ == '__main__':
    unittest.main()
