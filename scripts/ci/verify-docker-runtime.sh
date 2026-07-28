#!/usr/bin/env bash
set -Eeuo pipefail

mode="${1:-}"
evidence_dir="${EVIDENCE_DIR:-artifacts/docker-runtime}"
summary_file="$evidence_dir/run-summary.md"
mkdir -p "$evidence_dir"

compose() {
  docker compose "$@"
}

append_summary() {
  printf '%s\n' "$*" | tee -a "$summary_file" >> "${GITHUB_STEP_SUMMARY:-/dev/null}"
}

record_service_state() {
  local label="$1"
  {
    echo "timestamp=$(date -u +%FT%TZ) label=$label"
    compose ps --all
    for service in postgres redis qdrant api worker web; do
      container_id="$(compose ps -q "$service" 2>/dev/null || true)"
      if [[ -n "$container_id" ]]; then
        docker inspect --format \
          'service={{index .Config.Labels "com.docker.compose.service"}} id={{.Id}} status={{.State.Status}} health={{if .State.Health}}{{.State.Health.Status}}{{else}}none{{end}} restart_count={{.RestartCount}} exit_code={{.State.ExitCode}} started={{.State.StartedAt}} finished={{.State.FinishedAt}}' \
          "$container_id"
      fi
    done
  } | tee -a "$evidence_dir/service-transitions.txt"
}

wait_for_health() {
  local service="$1"
  local expected="${2:-healthy}"
  local attempts="${3:-60}"
  local container_id state
  for ((attempt=1; attempt<=attempts; attempt++)); do
    container_id="$(compose ps -q "$service" 2>/dev/null || true)"
    if [[ -n "$container_id" ]]; then
      state="$(docker inspect --format '{{if .State.Health}}{{.State.Health.Status}}{{else}}{{.State.Status}}{{end}}' "$container_id")"
      printf '%s service=%s attempt=%s state=%s\n' "$(date -u +%FT%TZ)" "$service" "$attempt" "$state" |
        tee -a "$evidence_dir/health-transitions.txt"
      if [[ "$state" == "$expected" ]]; then
        return 0
      fi
    fi
    sleep 5
  done
  return 1
}

request() {
  local name="$1"
  local url="$2"
  local allowed="$3"
  local body_file="$evidence_dir/http-${name}.body"
  local metrics_file="$evidence_dir/http-results.txt"
  local result status duration
  result="$(curl --silent --show-error --output "$body_file" --write-out '%{http_code} %{time_total}' \
    --max-time 10 "$url")"
  status="${result%% *}"
  duration="${result##* }"
  printf '%s name=%s url=%s status=%s duration_seconds=%s allowed=%s\n' \
    "$(date -u +%FT%TZ)" "$name" "$url" "$status" "$duration" "$allowed" |
    tee -a "$metrics_file"
  if [[ ! "$status" =~ ^($allowed)$ ]]; then
    echo "Unexpected HTTP status for $name: $status" >&2
    return 1
  fi
}

collect_diagnostics() {
  set +e
  record_service_state "diagnostics"
  compose logs --no-color --timestamps > "$evidence_dir/compose.log" 2>&1
  compose images > "$evidence_dir/compose-images.txt" 2>&1
  docker volume ls --filter "label=com.docker.compose.project=${COMPOSE_PROJECT_NAME:-clipsense-ci}" \
    > "$evidence_dir/volumes.txt" 2>&1
  for service in postgres redis qdrant api worker web; do
    container_id="$(compose ps -q "$service" 2>/dev/null)"
    if [[ -n "$container_id" ]]; then
      docker inspect --format '{{json .State}}' "$container_id" \
        > "$evidence_dir/inspect-${service}-state.json" 2>&1
    fi
  done
  docker system df > "$evidence_dir/docker-system-df.txt" 2>&1
  compose down --timeout 30 > "$evidence_dir/compose-down.txt" 2>&1
  remaining="$(compose ps -aq 2>/dev/null)"
  if [[ -n "$remaining" ]]; then
    echo "Containers remained after Compose shutdown: $remaining" |
      tee -a "$evidence_dir/compose-down.txt"
    return 1
  fi
  echo "Clean full-stack shutdown verified." | tee -a "$evidence_dir/compose-down.txt"
}

case "$mode" in
  config)
    start="$(date +%s)"
    compose config --quiet
    {
      echo "source_commit=${SOURCE_SHA:-$(git rev-parse HEAD)}"
      echo "validated_at=$(date -u +%FT%TZ)"
      echo "services:"
      compose config --services
      echo "images:"
      compose config --images
      echo "volumes:"
      compose config --volumes
      echo "published_ports:"
      compose config --format json |
        jq -r '.services | to_entries[] | .key as $service | (.value.ports // [])[] | "\($service) \(.published):\(.target)"'
      echo "healthchecks:"
      compose config --format json |
        jq -r '.services | to_entries[] | select(.value.healthcheck != null) | .key'
      echo "dependency_conditions:"
      compose config --format json |
        jq -r '.services | to_entries[] | .key as $service | (.value.depends_on // {}) | to_entries[] | "\($service) -> \(.key): \(.value.condition)"'
      echo "build_contexts:"
      compose config --format json |
        jq -r '.services | to_entries[] | select(.value.build != null) | "\(.key) context=\(.value.build.context) dockerfile=\(.value.build.dockerfile)"'
      echo "duration_seconds=$(($(date +%s)-start))"
    } | tee "$evidence_dir/compose-metadata.txt"
    append_summary ""
    append_summary "## Compose configuration"
    append_summary ""
    append_summary "- Result: passed"
    append_summary "- Services: postgres, redis, qdrant, api, worker, web"
    ;;
  build)
    start="$(date +%s)"
    docker buildx bake --file docker-compose.yml api worker web \
      --load \
      --set "api.tags=${COMPOSE_PROJECT_NAME:-clipsense-ci}-api" \
      --set "worker.tags=${COMPOSE_PROJECT_NAME:-clipsense-ci}-worker" \
      --set "web.tags=${COMPOSE_PROJECT_NAME:-clipsense-ci}-web" \
      --set '*.cache-from=type=gha' \
      --set '*.cache-to=type=gha,mode=max' \
      --progress plain 2>&1 |
      tee "$evidence_dir/build.log"
    duration="$(($(date +%s)-start))"
    {
      echo "source_commit=${SOURCE_SHA:-$(git rev-parse HEAD)}"
      echo "duration_seconds=$duration"
      for service in api worker web; do
        image_id="${COMPOSE_PROJECT_NAME:-clipsense-ci}-${service}"
        docker image inspect --format \
          'service='"$service"' id={{.Id}} size_bytes={{.Size}} created={{.Created}} repo_digests={{json .RepoDigests}}' \
          "$image_id"
      done
    } | tee "$evidence_dir/image-inventory.txt"
    compose run --rm --no-deps worker sh -ec \
      'python -m py_compile main.py zip_safety.py && python -c "import main; import zip_safety" && ffmpeg -version | head -n 1' |
      tee "$evidence_dir/worker-build-smoke.txt"
    append_summary ""
    append_summary "## Image build"
    append_summary ""
    append_summary "- Result: API, worker, and web images built"
    append_summary "- Duration: ${duration}s"
    append_summary "- Worker Python compile/import and FFmpeg smoke: passed"
    ;;
  runtime)
    runtime_start="$(date +%s)"
    record_service_state "before-start"
    compose up --detach --no-build
    record_service_state "started"
    for service in postgres redis qdrant api web; do
      wait_for_health "$service" healthy 120
      record_service_state "${service}-healthy"
    done

    request "api-live" "http://localhost:8080/api/health/live" "200"
    request "api-ready" "http://localhost:8080/api/health/ready" "200"
    request "api-compatibility" "http://localhost:8080/api/health" "200"
    request "web-root" "http://localhost:3000/" "2[0-9][0-9]|3[0-9][0-9]|401|403"
    request "web-dashboard" "http://localhost:3000/dashboard" "2[0-9][0-9]|3[0-9][0-9]|401|403"
    request "web-batch-new" "http://localhost:3000/dashboard/batches/new" "2[0-9][0-9]|3[0-9][0-9]|401|403"
    request "web-batch-detail" "http://localhost:3000/dashboard/batches/test-id" "2[0-9][0-9]|3[0-9][0-9]|401|403"
    request "web-settings" "http://localhost:3000/dashboard/settings" "2[0-9][0-9]|3[0-9][0-9]|401|403"

    compose exec -T postgres pg_isready -U clipsense -d clipsense |
      tee "$evidence_dir/postgres-check.txt"
    compose exec -T redis redis-cli ping | tee "$evidence_dir/redis-check.txt"
    curl --fail --silent --show-error http://localhost:6333/readyz |
      tee "$evidence_dir/qdrant-check.txt"
    compose exec -T worker sh -ec '
      python --version
      python -c "import main; import zip_safety; print(\"worker imports: ok\")"
      ffmpeg -version | head -n 1
      python -c "import os, redis, psycopg, urllib.request; redis.from_url(os.environ[\"REDIS_URL\"]).ping(); psycopg.connect(os.environ[\"DATABASE_URL\"]).close(); urllib.request.urlopen(\"http://%s:%s/readyz\" % (os.environ[\"QDRANT_HOST\"], os.environ[\"QDRANT_PORT\"]), timeout=5).read(); print(\"worker dependencies: ok\")"
      du -sh /root/.cache 2>/dev/null || true
    ' | tee "$evidence_dir/worker-runtime-smoke.txt"

    worker_booted=false
    for attempt in {1..180}; do
      blocked_clients="$(compose exec -T redis redis-cli --raw info clients |
        awk -F: '/^blocked_clients:/ {gsub(/\r/, "", $2); print $2}')"
      printf '%s attempt=%s redis_blocked_clients=%s\n' \
        "$(date -u +%FT%TZ)" "$attempt" "${blocked_clients:-unknown}" |
        tee -a "$evidence_dir/worker-startup.txt"
      if [[ "${blocked_clients:-0}" -ge 1 ]]; then
        worker_booted=true
        break
      fi
      sleep 5
    done
    [[ "$worker_booted" == "true" ]]
    compose exec -T worker sh -ec 'du -sh /root/.cache 2>/dev/null || true' |
      tee -a "$evidence_dir/worker-startup.txt"
    record_service_state "worker-smoke-complete"

    worker_id="$(compose ps -q worker)"
    worker_restart_before="$(docker inspect --format '{{.RestartCount}}' "$worker_id")"
    sleep 10
    worker_running="$(docker inspect --format '{{.State.Running}}' "$worker_id")"
    worker_restart_after="$(docker inspect --format '{{.RestartCount}}' "$worker_id")"
    [[ "$worker_running" == "true" && "$worker_restart_before" == "$worker_restart_after" ]]
    if compose logs --no-color worker | grep -Eq 'ModuleNotFoundError|ImportError|Traceback|uncontrolled initialization'; then
      echo "Worker logs contain a prohibited startup error." >&2
      exit 1
    fi

    compose stop --timeout 10 redis
    record_service_state "redis-stopped"
    request "degraded-live" "http://localhost:8080/api/health/live" "200"
    request "degraded-ready" "http://localhost:8080/api/health/ready" "503"
    compose start redis
    wait_for_health redis healthy 60
    request "recovered-ready" "http://localhost:8080/api/health/ready" "200"
    record_service_state "redis-recovered"

    api_id="$(compose ps -q api)"
    docker kill --signal=SIGTERM "$api_id" > "$evidence_dir/api-sigterm.txt"
    for _ in {1..30}; do
      api_running="$(docker inspect --format '{{.State.Running}}' "$api_id")"
      [[ "$api_running" == "false" ]] && break
      sleep 1
    done
    api_running="$(docker inspect --format '{{.State.Running}}' "$api_id")"
    api_exit="$(docker inspect --format '{{.State.ExitCode}}' "$api_id")"
    printf 'running=%s exit_code=%s\n' "$api_running" "$api_exit" |
      tee -a "$evidence_dir/api-sigterm.txt"
    [[ "$api_running" == "false" && "$api_exit" == "0" ]]
    record_service_state "api-graceful-termination"

    duration="$(($(date +%s)-runtime_start))"
    append_summary ""
    append_summary "## Runtime integration"
    append_summary ""
    append_summary "- Result: passed"
    append_summary "- Startup and verification duration: ${duration}s"
    append_summary "- Dependencies: PostgreSQL, Redis, and Qdrant healthy"
    append_summary "- API: live, ready, and compatibility endpoints passed"
    append_summary "- Routes: all five canonical routes avoided 404/500"
    append_summary "- Worker: imports, FFmpeg, dependencies, and stability passed"
    append_summary "- Redis degradation: readiness failed while liveness remained healthy, then recovered"
    append_summary "- API SIGTERM: clean exit code 0"
    ;;
  diagnostics)
    collect_diagnostics
    ;;
  *)
    echo "Usage: $0 {config|build|runtime|diagnostics}" >&2
    exit 2
    ;;
esac
