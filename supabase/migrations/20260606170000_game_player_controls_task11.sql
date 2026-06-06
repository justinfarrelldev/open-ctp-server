alter table "public"."game_players"
    add column if not exists "country" text not null default ''::text,
    add column if not exists "ping_ms" integer not null default 0;

update "public"."game_players"
set
    "country" = coalesce("country", ''::text),
    "ping_ms" = coalesce("ping_ms", 0);

do $$
begin
    if not exists (
        select 1
        from pg_constraint
        where conname = 'game_players_ping_ms_nonnegative'
    ) then
        alter table "public"."game_players"
            add constraint "game_players_ping_ms_nonnegative"
            check ("ping_ms" >= 0);
    end if;
end $$;

create unique index if not exists game_players_game_id_country_unique_idx
    on public.game_players using btree (game_id, lower(country))
    where btrim(country) <> '';
