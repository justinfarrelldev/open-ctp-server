# Task 038 source/API check: space-launch command gateway route

Retail/source-backed mechanic:

- Task 037 added source-backed execution for retail `SpaceLaunch` units.
- The server's public command surface must expose behavior-bearing mechanics through legal projected actions and HTTP command routes rather than hidden-state shortcuts.

Server alignment:

- The HTTP gameplay command gateway now accepts `/game/{id}/space-launch`.
- The route maps to canonical `engine.CommandSpaceLaunch` and passes only the public payload fields `unit_id` and `target_tile_id`.
- The route test verifies that server-derived authority is preserved and the canonical command kind/payload reach the command executor.
