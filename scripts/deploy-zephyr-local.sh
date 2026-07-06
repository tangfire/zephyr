#!/bin/sh
set -eu

DEPLOY_ACTION="${DEPLOY_ACTION:-deploy}"
DEPLOY_DIR="${ZEPHYR_DEPLOY_DIR:-/opt/zephyr}"
BUILD_BIN="${ZEPHYR_BUILD_BIN:-.woodpecker-build/zephyr}"
HEALTH_URL="${ZEPHYR_HEALTH_URL:-http://127.0.0.1:8095/healthz}"

host_systemctl() {
  docker run --rm --privileged --pid=host -v /:/host docker:28-cli \
    sh -lc "nsenter -t 1 -m -u -i -n -p -- systemctl $*"
}

host_healthcheck() {
  attempts="${1:-30}"
  i=1
  while [ "$i" -le "$attempts" ]; do
    if docker run --rm --network host docker:28-cli \
      sh -lc "wget -qO- --timeout=5 '$HEALTH_URL'" >/tmp/zephyr-health.out 2>/tmp/zephyr-health.err; then
      cat /tmp/zephyr-health.out
      rm -f /tmp/zephyr-health.out /tmp/zephyr-health.err
      return 0
    fi
    sleep 2
    i=$((i + 1))
  done
  echo "Peapod health check failed: $HEALTH_URL" >&2
  cat /tmp/zephyr-health.err >&2 2>/dev/null || true
  rm -f /tmp/zephyr-health.out /tmp/zephyr-health.err
  return 1
}

case "$DEPLOY_ACTION" in
  deploy|restart|status) ;;
  *)
    echo "unsupported DEPLOY_ACTION=$DEPLOY_ACTION (expected deploy, restart or status)" >&2
    exit 1
    ;;
esac

if [ "$DEPLOY_ACTION" = "status" ]; then
  host_systemctl status zephyr --no-pager -l
  host_healthcheck 1
  exit 0
fi

if [ "$DEPLOY_ACTION" = "restart" ]; then
  host_systemctl restart zephyr
  host_healthcheck 30
  exit 0
fi

test -d "$DEPLOY_DIR"
test -f "$DEPLOY_DIR/.env.host"
test -x "$BUILD_BIN"
test -d frontend/dist

owner_group="$(stat -c '%u:%g' "$DEPLOY_DIR" 2>/dev/null || echo '1000:1000')"
stamp="$(date +%Y%m%d%H%M%S)"
backup_dir="$DEPLOY_DIR/.deploy/backups/$stamp"
mkdir -p "$backup_dir"

if [ -x "$DEPLOY_DIR/zephyr" ]; then
  cp "$DEPLOY_DIR/zephyr" "$backup_dir/zephyr"
fi
if [ -d "$DEPLOY_DIR/frontend/dist" ]; then
  mkdir -p "$backup_dir/frontend"
  tar -C "$DEPLOY_DIR/frontend/dist" -cf - . | tar -xf - -C "$backup_dir/frontend"
fi

tar \
  --exclude .env \
  --exclude .env.host \
  --exclude 'data' \
  --exclude 'frontend/node_modules' \
  --exclude 'frontend/dist' \
  --exclude '.woodpecker-build' \
  --exclude 'zephyr' \
  --exclude '*.bak*' \
  -cf - . | tar -xf - -C "$DEPLOY_DIR"

rm -rf "$DEPLOY_DIR/frontend/dist"
mkdir -p "$DEPLOY_DIR/frontend/dist"
tar -C frontend/dist -cf - . | tar -xf - -C "$DEPLOY_DIR/frontend/dist"

install -m 0755 "$BUILD_BIN" "$DEPLOY_DIR/zephyr.next"
mv -f "$DEPLOY_DIR/zephyr.next" "$DEPLOY_DIR/zephyr"

chown -R "$owner_group" \
  "$DEPLOY_DIR/.git" \
  "$DEPLOY_DIR/.woodpecker" \
  "$DEPLOY_DIR/config" \
  "$DEPLOY_DIR/docs" \
  "$DEPLOY_DIR/examples" \
  "$DEPLOY_DIR/frontend" \
  "$DEPLOY_DIR/scripts" \
  "$DEPLOY_DIR/static" \
  "$DEPLOY_DIR/Dockerfile" \
  "$DEPLOY_DIR/README.md" \
  "$DEPLOY_DIR/docker-compose.yml" \
  "$DEPLOY_DIR/go.mod" \
  "$DEPLOY_DIR/go.sum" \
  "$DEPLOY_DIR/main.go" \
  "$DEPLOY_DIR/monitoring.go" \
  "$DEPLOY_DIR/monitoring_test.go" \
  "$DEPLOY_DIR/pipeline_summary_test.go" \
  "$DEPLOY_DIR/store.go" \
  "$DEPLOY_DIR/zephyr" 2>/dev/null || true
chown "$owner_group" "$DEPLOY_DIR" 2>/dev/null || true
chmod 0755 "$DEPLOY_DIR" 2>/dev/null || true

printf '%s\n' "${CI_COMMIT_SHA:-unknown}" > "$DEPLOY_DIR/.deploy/current-source-sha"
printf '%s deploy %s pipeline=%s\n' "$(date -Is)" "${CI_COMMIT_SHA:-unknown}" "${CI_PIPELINE_NUMBER:-manual}" >> "$DEPLOY_DIR/.deploy/deploy-history.log"

host_systemctl restart zephyr
host_healthcheck 30
