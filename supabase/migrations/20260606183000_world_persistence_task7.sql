create table if not exists "public"."world_snapshots" (
    "game_id" bigint primary key,
    "created_at" timestamp with time zone not null default now(),
    "updated_at" timestamp with time zone not null default now(),
    "game_status" text not null,
    "generator_version" text not null,
    "style" text not null,
    "rules_json" jsonb not null,
    "exclusions_json" jsonb not null,
    "starting_settlers" integer not null,
    "starting_gold" integer not null,
    "seed" bigint,
    "wetness" integer,
    "map_size" text not null,
    "climate" text not null,
    "player_order_json" jsonb not null,
    "spawns_json" jsonb not null,
    "tiles_json" jsonb not null,
    constraint "world_snapshots_game_id_fkey"
        foreign key ("game_id")
        references "public"."games"("id")
        on update cascade
        on delete cascade,
    constraint "world_snapshots_starting_settlers_nonnegative"
        check ("starting_settlers" >= 0),
    constraint "world_snapshots_starting_gold_nonnegative"
        check ("starting_gold" >= 0)
);

alter table "public"."world_snapshots" enable row level security;

create table if not exists "public"."world_turn_states" (
    "game_id" bigint primary key,
    "created_at" timestamp with time zone not null default now(),
    "updated_at" timestamp with time zone not null default now(),
    "turn_number" integer not null,
    "current_player_index" integer not null,
    "active_player_order_json" jsonb not null,
    "recalculation_count" integer not null,
    "rng_state" text not null,
    "ratings_json" jsonb not null,
    "pollution_json" jsonb not null,
    "last_recalculated_at" timestamp with time zone not null,
    constraint "world_turn_states_game_id_fkey"
        foreign key ("game_id")
        references "public"."world_snapshots"("game_id")
        on update cascade
        on delete cascade
);

alter table "public"."world_turn_states" enable row level security;

grant delete on table "public"."world_snapshots" to "anon";
grant insert on table "public"."world_snapshots" to "anon";
grant references on table "public"."world_snapshots" to "anon";
grant select on table "public"."world_snapshots" to "anon";
grant trigger on table "public"."world_snapshots" to "anon";
grant truncate on table "public"."world_snapshots" to "anon";
grant update on table "public"."world_snapshots" to "anon";
grant delete on table "public"."world_snapshots" to "authenticated";
grant insert on table "public"."world_snapshots" to "authenticated";
grant references on table "public"."world_snapshots" to "authenticated";
grant select on table "public"."world_snapshots" to "authenticated";
grant trigger on table "public"."world_snapshots" to "authenticated";
grant truncate on table "public"."world_snapshots" to "authenticated";
grant update on table "public"."world_snapshots" to "authenticated";
grant delete on table "public"."world_snapshots" to "service_role";
grant insert on table "public"."world_snapshots" to "service_role";
grant references on table "public"."world_snapshots" to "service_role";
grant select on table "public"."world_snapshots" to "service_role";
grant trigger on table "public"."world_snapshots" to "service_role";
grant truncate on table "public"."world_snapshots" to "service_role";
grant update on table "public"."world_snapshots" to "service_role";

grant delete on table "public"."world_turn_states" to "anon";
grant insert on table "public"."world_turn_states" to "anon";
grant references on table "public"."world_turn_states" to "anon";
grant select on table "public"."world_turn_states" to "anon";
grant trigger on table "public"."world_turn_states" to "anon";
grant truncate on table "public"."world_turn_states" to "anon";
grant update on table "public"."world_turn_states" to "anon";
grant delete on table "public"."world_turn_states" to "authenticated";
grant insert on table "public"."world_turn_states" to "authenticated";
grant references on table "public"."world_turn_states" to "authenticated";
grant select on table "public"."world_turn_states" to "authenticated";
grant trigger on table "public"."world_turn_states" to "authenticated";
grant truncate on table "public"."world_turn_states" to "authenticated";
grant update on table "public"."world_turn_states" to "authenticated";
grant delete on table "public"."world_turn_states" to "service_role";
grant insert on table "public"."world_turn_states" to "service_role";
grant references on table "public"."world_turn_states" to "service_role";
grant select on table "public"."world_turn_states" to "service_role";
grant trigger on table "public"."world_turn_states" to "service_role";
grant truncate on table "public"."world_turn_states" to "service_role";
grant update on table "public"."world_turn_states" to "service_role";
