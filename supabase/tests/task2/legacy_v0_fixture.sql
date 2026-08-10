-- Synthetic pre-Task-2 fixture. This file runs only after reset through 20260606183000.
do $$
begin
    insert into public.account (id, name, info, location, email)
    values (8101, 'legacy-player', '', '', 'legacy-player@example.test');

    insert into public.games (id, name, status, style, rules_json, exclusions_json, starting_settlers, starting_gold)
    values (8100, 'legacy-game', 'started', 'standard_ctp2', '[]'::jsonb, '{}'::jsonb, 1, 100);

    insert into public.game_players (id, game_id, account_id, country, ping_ms)
    values (8102, 8100, 8101, 'Legacy', 0);

    insert into public.world_snapshots (
        game_id, game_status, generator_version, style, rules_json, exclusions_json,
        starting_settlers, starting_gold, seed, wetness, map_size, climate,
        player_order_json, spawns_json, tiles_json
    ) values (
        8100, 'started', 'legacy-world-v0', 'standard_ctp2', '[]'::jsonb, '{}'::jsonb,
        1, 100, 77, 2, 'small', 'temperate',
        '[8102]'::jsonb, '[]'::jsonb, '[]'::jsonb
    );

    insert into public.world_turn_states (
        game_id, turn_number, current_player_index, active_player_order_json,
        recalculation_count, rng_state, ratings_json, pollution_json, last_recalculated_at
    ) values (
        8100, 3, 0, '[8102]'::jsonb, 2, '77',
        '{"overall":1}'::jsonb, '{"global_pollution":0}'::jsonb, now()
    );
end;
$$;