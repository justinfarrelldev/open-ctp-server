\set ON_ERROR_STOP on

DO $$
BEGIN
    IF (SELECT state_schema_version FROM public.world_turn_states WHERE game_id = 8100) <> 0 THEN
        RAISE EXCEPTION 'legacy state schema version changed';
    END IF;
    IF (SELECT revision FROM public.world_turn_states WHERE game_id = 8100) <> 0 THEN
        RAISE EXCEPTION 'legacy revision changed';
    END IF;
    IF (SELECT last_event_sequence FROM public.world_turn_states WHERE game_id = 8100) <> 0 THEN
        RAISE EXCEPTION 'legacy event sequence changed';
    END IF;
    IF (SELECT turn_number FROM public.world_turn_states WHERE game_id = 8100) <> 3 THEN
        RAISE EXCEPTION 'legacy turn data changed';
    END IF;
    IF (SELECT generator_version FROM public.world_snapshots WHERE game_id = 8100) <> 'legacy-world-v0' THEN
        RAISE EXCEPTION 'legacy world snapshot changed';
    END IF;
END;
$$;
