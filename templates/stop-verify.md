# Stop-verify checklist

Before saying a task is done. Each line is a claim that has to be checked, not remembered.

- [ ] **The command was run, and the output was read.** "It should work" is not a result.
- [ ] **The exit code was checked.** A test run piped into `tail` hides the failure and keeps the pipeline
      green; a command chained with `&&` is the fix.
- [ ] **The artifact is at the path that was reported.** `ls` it, and read back the first lines if it is
      generated.
- [ ] **The claim matches the measurement.** If the README says 95% coverage, that number came from a run
      of the command next to it, on this machine, this week.
- [ ] **What could not be verified is named.** "Not verified against a live server" is a result; silence is
      a claim nobody can check.
- [ ] **Nothing secret was written.** Not into the repository, not into memory, not into the report: a
      finding carries the kind and the length, never the value.
- [ ] **The next reader can find this.** Commit message says why, not what; the file path is in the report.

## The two ways this is usually skipped

Running the test **after** the commit, and reading the last three lines of output because the head looked
fine. Both have happened in this repository's own history: a `go test | tail` chained into a commit let a
broken file land, and a workflow YAML that did not parse was pushed because it was read rather than parsed.
