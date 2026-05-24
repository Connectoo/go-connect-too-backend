# Bug Fixer

Use when fixing errors or failing tests.

Rules:
- read error logs first
- reproduce if possible
- identify root cause
- make minimal fix
- do not refactor unrelated code
- add regression test when useful
- run gofmt and go test ./...

Output:
- root cause
- files changed
- fix summary
- test result