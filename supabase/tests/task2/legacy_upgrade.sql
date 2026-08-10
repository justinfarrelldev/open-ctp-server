begin;
set local statement_timeout = '5min';
select plan(7);

select has_column('public', 'world_turn_states', 'state_schema_version', 'upgrade adds an explicit state schema version');
select has_column('public', 'world_turn_states', 'canonical_state_json', 'upgrade adds canonical state storage');
select is((select state_schema_version from public.world_turn_states where game_id = 8100), 0, 'legacy turn state remains v0');
select is((select revision from public.world_turn_states where game_id = 8100), 0::bigint, 'legacy revision defaults to zero');
select is((select last_event_sequence from public.world_turn_states where game_id = 8100), 0::bigint, 'legacy event sequence defaults to zero');
select is((select turn_number from public.world_turn_states where game_id = 8100), 3, 'legacy turn data survives the upgrade');
select is((select generator_version from public.world_snapshots where game_id = 8100), 'legacy-world-v0', 'legacy immutable snapshot survives the upgrade');

select * from finish();
rollback;