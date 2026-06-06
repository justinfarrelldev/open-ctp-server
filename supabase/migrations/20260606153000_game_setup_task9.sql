alter table "public"."games"
    add column if not exists "updated_at" timestamp with time zone not null default now(),
    add column if not exists "started_at" timestamp with time zone,
    add column if not exists "style" text not null default 'standard_ctp2'::text,
    add column if not exists "rules_json" jsonb not null default '[]'::jsonb,
    add column if not exists "exclusions_json" jsonb not null default '{}'::jsonb,
    add column if not exists "status" text not null default 'setup'::text,
    add column if not exists "starting_settlers" integer not null default 1,
    add column if not exists "starting_gold" integer not null default 100;

update "public"."games"
set
    "updated_at" = coalesce("updated_at", "created_at", now()),
    "style" = coalesce(nullif("style", ''), 'standard_ctp2'::text),
    "rules_json" = coalesce("rules_json", '[]'::jsonb),
    "exclusions_json" = coalesce("exclusions_json", '{}'::jsonb),
    "status" = coalesce(nullif("status", ''), 'setup'::text),
    "starting_settlers" = coalesce("starting_settlers", 1),
    "starting_gold" = coalesce("starting_gold", 100);

alter table "public"."game_players"
    add column if not exists "updated_at" timestamp with time zone not null default now();

update "public"."game_players"
set
    "updated_at" = coalesce("updated_at", "joined_at", "created_at", now());

do $$
begin
    if not exists (
        select 1
        from pg_constraint
        where conname = 'games_starting_settlers_nonnegative'
    ) then
        alter table "public"."games"
            add constraint "games_starting_settlers_nonnegative"
            check ("starting_settlers" >= 0);
    end if;

    if not exists (
        select 1
        from pg_constraint
        where conname = 'games_starting_gold_nonnegative'
    ) then
        alter table "public"."games"
            add constraint "games_starting_gold_nonnegative"
            check ("starting_gold" >= 0);
    end if;
end $$;

drop view if exists "public"."game";

create view "public"."game" as
select
    "games"."id",
    "games"."password_protected",
    "games"."password",
    "games"."name",
    "games"."owner_account_id",
    "games"."lobby_id",
    "games"."is_locked",
    "games"."is_public",
    "games"."style",
    "games"."rules_json",
    "games"."exclusions_json",
    "games"."status",
    "games"."starting_settlers",
    "games"."starting_gold",
    "games"."created_at",
    "games"."updated_at",
    "games"."started_at"
from "public"."games";

grant delete on table "public"."game" to "anon";
grant insert on table "public"."game" to "anon";
grant select on table "public"."game" to "anon";
grant update on table "public"."game" to "anon";
grant delete on table "public"."game" to "authenticated";
grant insert on table "public"."game" to "authenticated";
grant select on table "public"."game" to "authenticated";
grant update on table "public"."game" to "authenticated";
grant delete on table "public"."game" to "service_role";
grant insert on table "public"."game" to "service_role";
grant select on table "public"."game" to "service_role";
grant update on table "public"."game" to "service_role";
