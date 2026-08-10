-- Task 2 gameplay persistence is forward-only. Legacy world rows remain schema v0;
-- no migration rewrites them as v1 because they have no immutable imported ruleset pin.

alter table "public"."world_snapshots"
    add column if not exists "ruleset_canonical_json" jsonb,
    add column if not exists "ruleset_content_hash" text,
    add column if not exists "ruleset_version" text;

alter table "public"."world_turn_states"
    add column if not exists "state_schema_version" integer not null default 0,
    add column if not exists "revision" bigint not null default 0,
    add column if not exists "last_event_sequence" bigint not null default 0,
    add column if not exists "canonical_state_json" jsonb;

do $$
begin
    if not exists (select 1 from pg_constraint where conname = 'world_snapshots_ruleset_pin_complete') then
        alter table "public"."world_snapshots"
            add constraint "world_snapshots_ruleset_pin_complete"
            check (("ruleset_canonical_json" is null and "ruleset_content_hash" is null and "ruleset_version" is null)
                or ("ruleset_canonical_json" is not null and "ruleset_content_hash" ~ '^[0-9a-f]{64}$' and btrim("ruleset_version") <> ''));
    end if;
    if not exists (select 1 from pg_constraint where conname = 'world_turn_states_state_schema_version_valid') then
        alter table "public"."world_turn_states"
            add constraint "world_turn_states_state_schema_version_valid"
            check ("state_schema_version" in (0, 1));
    end if;
    if not exists (select 1 from pg_constraint where conname = 'world_turn_states_revision_nonnegative') then
        alter table "public"."world_turn_states"
            add constraint "world_turn_states_revision_nonnegative" check ("revision" >= 0);
    end if;
    if not exists (select 1 from pg_constraint where conname = 'world_turn_states_last_event_sequence_nonnegative') then
        alter table "public"."world_turn_states"
            add constraint "world_turn_states_last_event_sequence_nonnegative" check ("last_event_sequence" >= 0);
    end if;
    if not exists (select 1 from pg_constraint where conname = 'world_turn_states_v1_canonical_state_required') then
        alter table "public"."world_turn_states"
            add constraint "world_turn_states_v1_canonical_state_required"
            check ("state_schema_version" = 0 or ("canonical_state_json" is not null and jsonb_typeof("canonical_state_json") = 'object'));
    end if;
end $$;

-- A v1 aggregate always refers to a pin stored with the immutable snapshot. The
-- trigger is necessary because a CHECK cannot inspect world_snapshots.
create or replace function public.enforce_world_turn_state_v1_ruleset_pin()
returns trigger
language plpgsql
as $$
begin
    if new.state_schema_version = 1 and not exists (
        select 1 from public.world_snapshots snapshot
        where snapshot.game_id = new.game_id
          and snapshot.ruleset_canonical_json is not null
          and btrim(snapshot.ruleset_content_hash) <> ''
          and btrim(snapshot.ruleset_version) <> ''
    ) then
        raise exception 'v1 gameplay state requires pinned immutable ruleset';
    end if;
    if tg_op = 'UPDATE' and old.state_schema_version = 1 then
        if new.revision <> old.revision + 1 then
            raise exception 'v1 gameplay revision must advance exactly one';
        end if;
        if new.last_event_sequence < old.last_event_sequence then
            raise exception 'v1 gameplay event sequence must not decrease';
        end if;
    end if;
    return new;
end;
$$;

drop trigger if exists "world_turn_states_v1_invariants_trigger" on "public"."world_turn_states";
create trigger "world_turn_states_v1_invariants_trigger"
    before insert or update on "public"."world_turn_states"
    for each row execute function public.enforce_world_turn_state_v1_ruleset_pin();

-- Snapshots are immutable after creation, including their rules pin.
create or replace function public.reject_world_snapshot_update()
returns trigger
language plpgsql
as $$
begin
    raise exception 'world snapshots are immutable';
end;
$$;

drop trigger if exists "world_snapshots_immutable_trigger" on "public"."world_snapshots";
create trigger "world_snapshots_immutable_trigger"
    before update on "public"."world_snapshots"
    for each row execute function public.reject_world_snapshot_update();

alter table "public"."games"
    add column if not exists "ended_at" timestamp with time zone,
    add column if not exists "ended_by" bigint,
    add column if not exists "end_reason" text;

alter table "public"."game_players"
    add column if not exists "controller_kind" text not null default 'human'::text,
    add column if not exists "controller_changed_at" timestamp with time zone;

create unique index if not exists game_players_game_id_id_unique_idx
    on public.game_players using btree (game_id, id);

alter table "public"."games"
    drop constraint if exists "games_terminal_metadata_complete";

alter table "public"."games"
    add constraint "games_terminal_metadata_complete"
    check (("status" = 'ended' and "ended_at" is not null and "ended_by" is not null and "end_reason" is not null and btrim("end_reason") <> '')
        or ("status" <> 'ended' and "ended_at" is null and "ended_by" is null and "end_reason" is null));

do $$
begin
    if not exists (select 1 from pg_constraint where conname = 'games_ended_by_game_players_fkey') then
        alter table "public"."games"
            add constraint "games_ended_by_game_players_fkey"
            foreign key ("id", "ended_by") references "public"."game_players"("game_id", "id")
            on update cascade on delete restrict;
    end if;
    if not exists (select 1 from pg_constraint where conname = 'game_players_controller_kind_valid') then
        alter table "public"."game_players"
            add constraint "game_players_controller_kind_valid"
            check ("controller_kind" in ('human', 'ai'));
    end if;
end $$;

create table if not exists "public"."gameplay_command_receipts" (
    "game_id" bigint not null,
    "actor_player_id" bigint not null,
    "idempotency_key" text not null,
    "payload_sha256" text not null,
    "result_json" jsonb not null,
    "resulting_revision" bigint not null,
    "resulting_event_sequence" bigint not null,
    "created_at" timestamp with time zone not null default now(),
    constraint "gameplay_command_receipts_pkey" primary key ("game_id", "actor_player_id", "idempotency_key"),
    constraint "gameplay_command_receipts_payload_sha256_valid" check ("payload_sha256" ~ '^[0-9a-f]{64}$'),
    constraint "gameplay_command_receipts_resulting_revision_nonnegative" check ("resulting_revision" >= 0),
    constraint "gameplay_command_receipts_resulting_event_sequence_nonnegative" check ("resulting_event_sequence" >= 0),
    constraint "gameplay_command_receipts_game_player_fkey" foreign key ("game_id", "actor_player_id")
        references "public"."game_players"("game_id", "id") on update cascade on delete cascade
);

create table if not exists "public"."gameplay_events" (
    "game_id" bigint not null,
    "sequence" bigint not null,
    "revision" bigint not null,
    "ordinal" integer not null,
    "event_type" text not null,
    "payload_json" jsonb not null,
    "created_at" timestamp with time zone not null default now(),
    constraint "gameplay_events_pkey" primary key ("game_id", "sequence"),
    constraint "gameplay_events_revision_ordinal_unique" unique ("game_id", "revision", "ordinal"),
    constraint "gameplay_events_sequence_positive" check ("sequence" > 0),
    constraint "gameplay_events_revision_nonnegative" check ("revision" >= 0),
    constraint "gameplay_events_ordinal_nonnegative" check ("ordinal" >= 0),
    constraint "gameplay_events_type_nonempty" check (btrim("event_type") <> ''),
    constraint "gameplay_events_game_fkey" foreign key ("game_id") references "public"."games"("id") on update cascade on delete cascade
);

create table if not exists "public"."gameplay_event_outbox" (
    "game_id" bigint not null,
    "sequence" bigint not null,
    "created_at" timestamp with time zone not null default now(),
    "delivered_at" timestamp with time zone,
    "delivery_attempts" integer not null default 0,
    constraint "gameplay_event_outbox_pkey" primary key ("game_id", "sequence"),
    constraint "gameplay_event_outbox_event_fkey" foreign key ("game_id", "sequence")
        references "public"."gameplay_events"("game_id", "sequence") on update cascade on delete cascade,
    constraint "gameplay_event_outbox_attempts_nonnegative" check ("delivery_attempts" >= 0)
);

create index if not exists gameplay_events_game_id_revision_idx
    on public.gameplay_events using btree (game_id, revision, ordinal);
create index if not exists gameplay_event_outbox_undelivered_idx
    on public.gameplay_event_outbox using btree (game_id, sequence)
    where delivered_at is null;

alter table "public"."gameplay_command_receipts" enable row level security;
alter table "public"."gameplay_events" enable row level security;
alter table "public"."gameplay_event_outbox" enable row level security;
