alter table "public"."lobbies"
    add column if not exists "updated_at" timestamp with time zone not null default now(),
    add column if not exists "last_activity_at" timestamp with time zone not null default now(),
    add column if not exists "closed_at" timestamp with time zone;

update "public"."lobbies"
set
    "updated_at" = coalesce("updated_at", "created_at", now()),
    "last_activity_at" = coalesce("last_activity_at", "created_at", now()),
    "closed_at" = case
        when "closed_at" is not null then "closed_at"
        when "is_closed" then coalesce("updated_at", "created_at", now())
        else null
    end;

create or replace view "public"."lobby" as
select
    "lobbies"."id",
    "lobbies"."name",
    "lobbies"."owner_name",
    "lobbies"."owner_account_id",
    "lobbies"."is_closed",
    "lobbies"."is_muted",
    "lobbies"."is_public",
    "lobbies"."created_at",
    "lobbies"."updated_at",
    "lobbies"."last_activity_at",
    "lobbies"."closed_at"
from "public"."lobbies";
