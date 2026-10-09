---
name: loop-fix-until-clean
description: Run a fix-until-clean loop over a project - CI failures, review findings, lint debt - with a stated done-condition, an iteration ceiling, and a maker/checker split. Use when asked to iterate until no bugs remain, or when a fix produces more findings.
---

# Loop until clean

An unbounded "keep going until it is perfect" loop burns a session and rarely ends. What makes the loop work
is three things decided before it starts.

## 1. Done-condition, written down

Not "no bugs", which cannot be reached, but a check that can be run:

- every command in the CI workflow exits 0 locally, with the same toolchain the CI uses;
- the checker the project already has (a linter, `check-diagrams.mjs`, `make verify`) passes;
- every claim in the README was produced by a command that exists.

## 2. Iteration ceiling and state on disk

Fix, verify, repeat — but cap the rounds (three is usually enough to find out whether the loop is converging
or thrashing) and write state after each round: what failed, what was changed, what was verified. A loop
that lives only in a conversation loses everything on a compaction.

## 3. Maker and checker are different eyes

The session that wrote the fix is the worst judge of it. The cheap version: reproduce the failure **before**
the fix, with the exact command the CI runs, and run it **after** — a fix that cannot be shown to reproduce
the failure is a change, not a fix.

## The two failure modes worth naming

- **Guessing at a checker instead of reading it.** Twice in one session a repository's own
  `check-diagrams.mjs` rejected a commit because the reference was written `diagrams/x.mmd` and the regex
  wanted `./diagrams/x.mmd`. The checker's source is ten lines; read it first.
- **Chaining without `&&`.** `pytest | tail -3` followed by `git commit` lets a failing suite land, and
  `git push` prints success when there was nothing to push. Chain with `&&`, then confirm with
  `git log --oneline -1`.
