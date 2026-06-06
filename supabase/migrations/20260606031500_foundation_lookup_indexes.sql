CREATE INDEX sessions_account_id_idx ON public.sessions USING btree (account_id);
CREATE INDEX lobbies_owner_account_id_idx ON public.lobbies USING btree (owner_account_id);
CREATE INDEX lobby_members_account_id_idx ON public.lobby_members USING btree (account_id);
CREATE INDEX lobby_messages_lobby_id_idx ON public.lobby_messages USING btree (lobby_id);
CREATE INDEX lobby_messages_account_id_idx ON public.lobby_messages USING btree (account_id);
CREATE INDEX games_lobby_id_idx ON public.games USING btree (lobby_id);
CREATE INDEX games_owner_account_id_idx ON public.games USING btree (owner_account_id);
CREATE INDEX game_players_account_id_idx ON public.game_players USING btree (account_id);
