begin;
set local statement_timeout = '5min';
select plan(14);

select has_table('public', 'gameplay_command_receipts', 'Task 2 creates command receipts');
select has_table('public', 'gameplay_events', 'Task 2 creates ordered events');
select has_table('public', 'gameplay_event_outbox', 'Task 2 creates the delivery outbox');
select has_column('public', 'world_turn_states', 'canonical_state_json', 'Task 2 adds canonical aggregate state');

insert into public.account (id, name, info, location, email)
values (9101, 'task2-player', '', '', 'task2-player@example.test');
insert into public.games (id, name, status, style, rules_json, exclusions_json, starting_settlers, starting_gold)
values (9100, 'task2-game', 'started', 'standard_ctp2', '[]'::jsonb, '{}'::jsonb, 1, 100);
insert into public.game_players (id, game_id, account_id, country, ping_ms, controller_kind)
values (9102, 9100, 9101, 'Task2', 0, 'human');
insert into public.world_snapshots (
    game_id, game_status, generator_version, style, rules_json, exclusions_json,
    starting_settlers, starting_gold, map_size, climate, player_order_json, spawns_json,
    tiles_json, ruleset_canonical_json, ruleset_content_hash, ruleset_version
) values (
    9100, 'started', 'task2-v1', 'standard_ctp2', '[]'::jsonb, '{}'::jsonb,
    1, 100, 'small', 'temperate', '[9102]'::jsonb, '[]'::jsonb, '[]'::jsonb,
    '{"version":"ctp2-v1"}'::jsonb,
    '0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef',
    'ctp2-v1'
);
insert into public.world_turn_states (
    game_id, turn_number, current_player_index, active_player_order_json,
    recalculation_count, rng_state, ratings_json, pollution_json, last_recalculated_at,
    state_schema_version, revision, last_event_sequence, canonical_state_json
) values (
    9100, 0, 0, '[9102]'::jsonb, 0, '1', '{}'::jsonb, '{}'::jsonb, now(),
    1, 0, 1, '{"phase":"setup"}'::jsonb
);

select lives_ok(
    $$insert into public.gameplay_command_receipts (game_id, actor_player_id, idempotency_key, payload_sha256, result_json, resulting_revision, resulting_event_sequence)
      values (9100, 9102, 'receipt-1', '0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef', '{}'::jsonb, 0, 1)$$,
    'first receipt is accepted'
);
select throws_ok(
    $$insert into public.gameplay_command_receipts (game_id, actor_player_id, idempotency_key, payload_sha256, result_json, resulting_revision, resulting_event_sequence)
      values (9100, 9102, 'receipt-1', '0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef', '{}'::jsonb, 0, 1)$$,
    '23505', null, 'duplicate receipt key is rejected'
);

insert into public.gameplay_events (game_id, sequence, revision, ordinal, event_type, payload_json)
values (9100, 1, 0, 0, 'created', '{}'::jsonb);
select throws_ok(
    $$insert into public.gameplay_events (game_id, sequence, revision, ordinal, event_type, payload_json)
      values (9100, 1, 0, 1, 'duplicate-sequence', '{}'::jsonb)$$,
    '23505', null, 'duplicate event sequence is rejected'
);
select throws_ok(
    $$update public.game_players set controller_kind = 'robot' where id = 9102$$,
    '23514', null, 'invalid controller kind is rejected'
);
select throws_ok(
    $$update public.games set status = 'ended' where id = 9100$$,
    '23514', null, 'ended status without terminal metadata is rejected'
);
select throws_ok(
    $$update public.games set ended_at = now(), ended_by = 9102, end_reason = 'invalid' where id = 9100$$,
    '23514', null, 'non-ended status with terminal metadata is rejected'
);
select throws_ok(
    $$update public.world_snapshots set ruleset_version = 'mutated' where game_id = 9100$$,
    'P0001', 'world snapshots are immutable', 'snapshot updates are rejected'
);

insert into public.account (id, name, info, location, email)
values (9103, 'task2-unpinned-player', '', '', 'task2-unpinned-player@example.test');
insert into public.games (id, name, status, style, rules_json, exclusions_json, starting_settlers, starting_gold)
values (9104, 'task2-unpinned-game', 'started', 'standard_ctp2', '[]'::jsonb, '{}'::jsonb, 1, 100);
insert into public.game_players (id, game_id, account_id, country, ping_ms, controller_kind)
values (9105, 9104, 9103, 'Unpinned', 0, 'human');
select throws_ok(
    $$insert into public.world_turn_states (
        game_id, turn_number, current_player_index, active_player_order_json,
        recalculation_count, rng_state, ratings_json, pollution_json, last_recalculated_at,
        state_schema_version, revision, last_event_sequence, canonical_state_json
    ) values (
        9104, 0, 0, '[9105]'::jsonb, 0, '1', '{}'::jsonb, '{}'::jsonb, now(),
        1, 0, 0, '{"phase":"setup"}'::jsonb
    )$$,
    'P0001', 'v1 gameplay state requires pinned immutable ruleset',
    'v1 aggregate without an immutable snapshot pin is rejected'
);
select throws_ok(
    $$update public.world_turn_states set revision = 2, last_event_sequence = 1 where game_id = 9100$$,
    'P0001', 'v1 gameplay revision must advance exactly one', 'v1 revision jumps are rejected'
);
select throws_ok(
    $$update public.world_turn_states set revision = 1, last_event_sequence = 0 where game_id = 9100$$,
    'P0001', 'v1 gameplay event sequence must not decrease', 'v1 event sequence regression is rejected'
);

select * from finish();
rollback;