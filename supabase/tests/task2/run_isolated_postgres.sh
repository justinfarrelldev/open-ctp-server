#!/usr/bin/env bash
set -euo pipefail

repo_root=$(CDPATH= cd -- "$(dirname -- "$0")/../../.." && pwd)
migrations_dir="$repo_root/supabase/migrations"
tests_dir="$repo_root/supabase/tests/task2"
container_name="task2-migrations-$$"
container_id=""

cleanup() {
    if [ -n "$container_id" ]; then
        docker rm --force --volumes "$container_id" >/dev/null 2>&1 || true
    fi
}
trap cleanup EXIT HUP INT TERM

bootstrap_sql='CREATE ROLE anon; CREATE ROLE authenticated; CREATE ROLE service_role; CREATE SCHEMA extensions; CREATE SCHEMA graphql; CREATE SCHEMA vault; CREATE SCHEMA pgsodium; CREATE PUBLICATION supabase_realtime;'

apply_sql() {
    docker exec -i "$container_id" psql --username postgres --dbname "$1" --set ON_ERROR_STOP=1 < "$2" >/dev/null
}

apply_migrations() {
    database=$1
    last_migration=${2:-}
    start_after=${3:-}
    applying=false
    if [ -z "$start_after" ]; then
        applying=true
    fi
    for migration in "$migrations_dir"/*.sql; do
        name=$(basename "$migration")
        if [ "$name" = "0_schema.sql" ]; then
            continue
        fi
        if [ -n "$start_after" ] && [ "$applying" = false ]; then
            if [ "$name" = "$start_after" ]; then
                applying=true
            fi
            continue
        fi
        apply_sql "$database" "$migration"
        if [ -n "$last_migration" ] && [ "$name" = "$last_migration" ]; then
            return
        fi
    done
}

container_id=$(docker run --detach --publish 127.0.0.1::5432 --env POSTGRES_HOST_AUTH_METHOD=trust --name "$container_name" postgres:16-alpine)
port=$(docker port "$container_id" 5432/tcp | awk -F: '{print $NF}')
if [ -z "$port" ] || [ "$port" = "54322" ]; then
    printf 'Docker did not allocate a safe dynamic PostgreSQL host port: %s\n' "$port" >&2
    exit 1
fi
until docker exec "$container_id" pg_isready --username postgres --dbname postgres >/dev/null 2>&1; do
    sleep 1
done

docker exec -i "$container_id" psql --username postgres --dbname postgres --set ON_ERROR_STOP=1 --command "$bootstrap_sql" >/dev/null
apply_migrations postgres
apply_sql postgres "$tests_dir/isolated_fresh_assertions.sql"

docker exec "$container_id" createdb --username postgres task2_legacy
apply_migrations task2_legacy 20260606183000_world_persistence_task7.sql
apply_sql task2_legacy "$tests_dir/legacy_v0_fixture.sql"
apply_migrations task2_legacy "" 20260606183000_world_persistence_task7.sql
apply_sql task2_legacy "$tests_dir/isolated_legacy_assertions.sql"

printf 'Task 2 fresh and legacy migrations passed on isolated PostgreSQL host port %s.\n' "$port"
