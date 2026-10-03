# Task 089 source notes — wonder icon/description/movie metadata classification

Retail source evidence:

- `ctp2_source_code/ctp2_code/gs/newdb/WonderRecord.cpp:508-521` parses `DefaultIcon` as an Icon DB record, `Movie` as a WonderMovie DB record, and `Description` as a string ID.
- `ctp2_source_code/ctp2_code/gs/newdb/WonderRecord.cpp:998-1006` exposes `GetDefaultIcon()` and `GetMovie()` by resolving through icon/movie databases.
- `ctp2_source_code/ctp2_code/ui/interface/wondermoviewin.cpp:49-85` uses wonder/icon movie and gameplay text data to display a wonder movie window after completion.
- `ctp2_source_code/ctp2_code/ui/interface/greatlibrarywindow.cpp:409-451` uses wonder `DefaultIcon`/icon metadata to populate Great Library still/movie/gameplay/history files.
- `ctp2_source_code/ctp2_code/ui/interface/EditQueue.cpp:1535-1538` uses wonder `DefaultIcon` for production queue display.

Classification:

- Wonder `DefaultIcon`, `Description`, and `Movie` are presentation/UI metadata for server gameplay. They do not drive legality, canonical gameplay state mutation, or deterministic simulation outcomes.

Validation:

- `go test ./internal/gameplay/rules -run TestRetailFieldInventorySnapshot -count=1`
