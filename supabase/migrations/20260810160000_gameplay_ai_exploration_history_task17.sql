-- Task 17 persists only accepted AI movement steps needed to prevent visible-route revisits.
-- It intentionally carries no canonical state, player projection, external execution data, or foreign-player data.
create table if not exists "public"."gameplay_ai_exploration_history" (
    "game_id" bigint not null,
    "actor_player_id" bigint not null,
    "sequence" bigint not null,
    "command_type" text not null,
    "destination_tile_id" text not null,
    "created_at" timestamp with time zone not null default now(),
    constraint "gameplay_ai_exploration_history_pkey" primary key ("game_id", "actor_player_id", "sequence"),
    constraint "gameplay_ai_exploration_history_sequence_positive" check ("sequence" > 0),
    constraint "gameplay_ai_exploration_history_command_type_nonempty" check (btrim("command_type") <> ''),
    constraint "gameplay_ai_exploration_history_destination_nonempty" check (btrim("destination_tile_id") <> ''),
    constraint "gameplay_ai_exploration_history_game_player_fkey" foreign key ("game_id", "actor_player_id")
        references "public"."game_players"("game_id", "id") on update cascade on delete cascade
);

create index if not exists gameplay_ai_exploration_history_actor_sequence_idx
    on public.gameplay_ai_exploration_history using btree (game_id, actor_player_id, sequence);

alter table "public"."gameplay_ai_exploration_history" enable row level security;
