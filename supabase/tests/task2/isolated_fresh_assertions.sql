\set ON_ERROR_STOP on

DO $$
DECLARE
    hash text := '0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef';
BEGIN
    IF to_regclass('public.gameplay_command_receipts') IS NULL
        OR to_regclass('public.gameplay_events') IS NULL
        OR to_regclass('public.gameplay_event_outbox') IS NULL
        OR to_regclass('public.gameplay_ai_exploration_history') IS NULL THEN
        RAISE EXCEPTION 'fresh migration did not create gameplay persistence tables';
    END IF;

    INSERT INTO public.account (id, name, info, location, email)
    VALUES (9101, 'task2-player', '', '', 'task2-player@example.test');
    INSERT INTO public.games (id, name, status, style, rules_json, exclusions_json, starting_settlers, starting_gold)
    VALUES (9100, 'task2-game', 'started', 'standard_ctp2', '[]'::jsonb, '{}'::jsonb, 1, 100);
    INSERT INTO public.game_players (id, game_id, account_id, country, ping_ms, controller_kind)
    VALUES (9102, 9100, 9101, 'Task2', 0, 'human');
    INSERT INTO public.world_snapshots (
        game_id, game_status, generator_version, style, rules_json, exclusions_json,
        starting_settlers, starting_gold, map_size, climate, player_order_json, spawns_json,
        tiles_json, ruleset_canonical_json, ruleset_content_hash, ruleset_version
    ) VALUES (
        9100, 'started', 'task2-v1', 'standard_ctp2', '[]'::jsonb, '{}'::jsonb,
        1, 100, 'small', 'temperate', '[9102]'::jsonb, '[]'::jsonb, '[]'::jsonb,
        '{"version":"ctp2-v1"}'::jsonb, hash, 'ctp2-v1'
    );
    INSERT INTO public.world_turn_states (
        game_id, turn_number, current_player_index, active_player_order_json,
        recalculation_count, rng_state, ratings_json, pollution_json, last_recalculated_at,
        state_schema_version, revision, last_event_sequence, canonical_state_json
    ) VALUES (
        9100, 0, 0, '[9102]'::jsonb, 0, '1', '{}'::jsonb, '{}'::jsonb, now(),
        1, 0, 1, '{"phase":"setup"}'::jsonb
    );
    INSERT INTO public.gameplay_command_receipts (
        game_id, actor_player_id, idempotency_key, payload_sha256, result_json,
        resulting_revision, resulting_event_sequence
    ) VALUES (9100, 9102, 'receipt-1', hash, '{}'::jsonb, 0, 1);
    INSERT INTO public.gameplay_events (game_id, sequence, revision, ordinal, event_type, payload_json)
    VALUES (9100, 1, 0, 0, 'created', '{}'::jsonb);
    INSERT INTO public.gameplay_ai_exploration_history (
        game_id, actor_player_id, sequence, command_type, destination_tile_id
    ) VALUES (9100, 9102, 1, 'move', 'tile-1-1');

    BEGIN
        UPDATE public.game_players SET controller_kind = 'robot' WHERE id = 9102;
        RAISE EXCEPTION 'invalid controller kind was accepted';
    EXCEPTION WHEN check_violation THEN
        NULL;
    END;
    BEGIN
        UPDATE public.world_turn_states SET revision = 2 WHERE game_id = 9100;
        RAISE EXCEPTION 'revision jump was accepted';
    EXCEPTION WHEN raise_exception THEN
        IF SQLERRM <> 'v1 gameplay revision must advance exactly one' THEN
            RAISE;
        END IF;
    END;
    BEGIN
        UPDATE public.world_snapshots SET ruleset_version = 'mutated' WHERE game_id = 9100;
        RAISE EXCEPTION 'snapshot mutation was accepted';
    EXCEPTION WHEN raise_exception THEN
        IF SQLERRM <> 'world snapshots are immutable' THEN
            RAISE;
        END IF;
    END;
END;
$$;
