# Task 024 source check: suppress legacy terraform unit commands in AI

Retail/source context:

- Retail terrain transformation is represented in `terrimprove.cdb` through terraform terrain-improvement records and target terrain fields.
- The server's retail public-works implementation already projects and executes terraform improvements through validated `build-improvement` actions using runtime terrain-improvement definitions and transform material/time costs.
- The older synthetic `terraform` unit command is a narrow compatibility route; full-game continuity showed projected terraform submissions could be rejected by engine eligibility checks after planning.

Implementation notes:

- AI continues to consume validated public-works `build-improvement` terraform candidates.
- AI no longer submits legacy `terraform` unit commands, preventing invalid autonomous submissions while preserving the retail public-works terraform path.
