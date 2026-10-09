# Memory first

A rule for every session in a project that has long-term memory. It exists because the failure it prevents
is expensive and looks like competence: an agent that re-derives a decision, or re-discovers a gotcha, and
reports it as new.

## Before acting

1. **Search memory for the task, not for the project.** `memory_query` with the words in the request, and
   `memory_recent` for what the last session was doing. A hit on `gotchas/*` or `decisions/*` is not
   background reading — it is an instruction to apply now.
2. **If the task touches a billing, pricing or vendor API, fetch the documentation.** Memory holds what was
   true when the page was written; a price is not a fact that survives a year.
3. **Read the page, not the snippet.** A snippet that looks right is how a session ends up implementing
   half a decision.

## While acting

4. **One question per search.** "memory" as a query returns pages about memory; "retention sweep deletes
   pinned pages" returns the page that says pinned pages are exempt.
5. **When a page was useful, say so** (`memory_feedback: helpful`). When it was wrong or stale, say that
   instead (`wrong` / `stale`), because a page nobody rates is a page whose decay is a coin flip.

## Before finishing

6. **Write down what would have saved you time**: a gotcha that cost more than ten minutes, a decision with
   a reason that is not in the code, a procedure that has steps in an order. With the source and the date —
   a page that cannot be attributed cannot be trusted or corrected.
7. **A page that contradicts an existing one is a bug in one of them.** Rewrite, do not append.
8. **Never paste a credential into memory.** `redact.Redact` runs on the write path for a reason; the report
   carries the kind and the length, never the value.

## Why this is a rule and not advice

The cost of skipping step 1 is paid by whoever reads the session afterwards, which is usually the same
person. The cost of skipping step 6 is paid every session until somebody writes it down.
