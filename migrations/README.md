# Database migrations

All PostgreSQL and TimescaleDB schema changes belong here. Do not create or alter tables from application code.

Start the local database, then apply all migrations with:

```bash
go run ./cmd/migrate up
```

Create a new numbered pair for every schema change:

```text
000009_descriptive_name.up.sql
000009_descriptive_name.down.sql
```
