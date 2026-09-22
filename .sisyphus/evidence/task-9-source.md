# Task 9 source review — 2026-08-29T02:15:27Z

Pinned upstream commit: `civctp2/civctp2@0668dd199c2519ff8ab7d2871e7c7a89e21c8e52`

Reviewed source evidence provided for Task 9:
- `Player.cpp#L6899-L6979`: `Player::GameOver` awards `GAME_OVER_WON_CONQUERED_WORLD` to the sole remaining non-dead player.
- `Player.cpp#L7021-L7029`, `#L1210-L1225`, `#L1540-L1587`: player death depends on losing cities plus the first-city/unit conditions; city/army removal triggers the dead-player check.
- `Player.cpp#L7034-L7177`: `StartDeath` and `RemoveDeadPlayers` mark dead players, remove remaining assets, and remove them from turn order/player lists.
- `network.cpp#L1073-L1153`: network removal converts the slot to robot and may end the current turn; it is continuity, not defeat.
- `slicfunc.cpp#L4442-L4445`, `civapp.cpp#L2592-L2603`, `#L2746-L2754`: quit-to-lobby / app end-game are UI/session exits, not the conquest/elimination terminal path.
- `net_action.cpp#L343-L377`, `TurnCnt.cpp#L293-L323`: terminal checks occur on state transition / end-turn / new-year paths, and simultaneous-network completion may be deferred.

Implementation consequence:
- Catalog terminal-oracle support was kept conservative and typed around declared terminal reasons.
- No resignation-as-defeat success path was fabricated.
- Rows without reviewed immutable artifacts remain provisional/fail-closed instead of receiving invented winners or hashes.
