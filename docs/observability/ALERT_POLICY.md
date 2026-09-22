# Alert Policy

Prometheus evaluates provisioned rules. Local alerts have no external delivery
receiver; CS-048 delivery/product analytics remain open. Use the dashboard and
Prometheus Alerts view during development. No production paging/SLO claim.

| Condition | Action |
|---|---|
| Target unavailable for 2 minutes | Check container state, listener, network and recent deployment |
| API dependency unavailable for 2 minutes | Inspect Postgres/Redis health and recovery events |
| Any cleanup failure in 5 minutes | Reconcile owning batch before deleting anything |
| Any job failure in 5 minutes | Follow batch correlation and failed stage |
| Zero processing filesystem free bytes | Stop intake and recover space safely |

Two minutes is a provisional local scrape/startup debounce, not a measured service
objective. Job/cleanup rules intentionally expose individual failures as warnings.
Storage warning percentage, auth spike, sustained failure ratio and queue-not-draining
thresholds **require baseline measurement**. A queue can legitimately remain nonempty
during expensive inference; do not fabricate a five-minute production deadline.
Measure representative job durations and arrival rates before enabling such rules.
Oldest-job age is not safely measurable from current list payloads alone.

Each rule links the incident runbook. Review rules after workload measurement.
