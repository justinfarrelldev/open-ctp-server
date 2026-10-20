# Overview

This is the open Call to Power server. The goal of this server is to make a full Call to Power game server that can host Call to Power games. See the README.md for more details.

## CTP2 source code

Right now, CTP2 is being focused on. The source code from the original source code release in 2003 is in the ctp2 folder.

# Plans

Plans live in the `plans` folder on the top-level of the project. Follow plans exactly. While implementing a feature, create a checklist in the checklists folder (in `plans`) that is full of checklist bullet points that are kept UP-TO-DATE with the implementation status.

## Methodology

Always use test-driven development with red-green-refactor. Never, ever skip any phases of red-green-refactor.

## Commits

Use atomic commits. You are not allowed to modify or skip the pre-commit hook.

## Reviews

After all planned work has been implemented, you must do a thorough code review of the work you have done to ensure it fully implements what was specified in a maintainable, safe manner. This is not a toy project - spare no expense on maintainability and readability over the long-term. Make sure to verify mechanics with the CTP2 source code (in ctp2/ctp2_released_game + ctp2_source_code). Obvious bugs in the original source should be fixed - for example, do not keep features that would lead to crashing.
