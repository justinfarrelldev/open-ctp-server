alter table public.gameplay_events
  add column if not exists event_id text;

update public.gameplay_events
set event_id = null
where event_id is not null and btrim(event_id) = '';
