# Measurements

Each number in the README has a command behind it. This is the record of a run.

## Environment

| | |
|---|---|
| Machine | 12-core x86-64 (AMD64), Ubuntu 24.04, Docker 29.1.3 |
| Go | 1.27.2 |
| Date | 2026-10-13 (UTC) |

Every host port in this file is the one the repository publishes: bound to loopback and offset from
the default, so a machine already running a Postgres, a Redis or three other stacks does not have to
stop them first.


## The redactor is the package with the tests

```
$ go test ./... -cover
ok  github.com/lasttoss/agent-memory-workflow  0.003s  coverage: 95.8% of statements
```

**Result.** 95.8% of statements, from 7 top-level tests, with no network and no credentials:

```
$ go test ./... -count=1 -v | grep -c '^--- PASS: Test'
7
```
