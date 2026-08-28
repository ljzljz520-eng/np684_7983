# BUG_REPRO

The following failures were observed while validating the initial project state.
Each section records what failed, how to reproduce it, and the complete command output.
They are preserved intentionally; only failing build gates are omitted from the generated Dockerfile.

## Failure 1: Go test (.)

- Observed problem: `Go test (.)` failed in the initial project state.
- Working directory: `.`
- Command: `cd /app && GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off go test -count=1 ./...`
- Exit status: `1`

```text
--- FAIL: TestTicketCodeConsumedOnlyOnce (0.01s)
    ticketdesk_test.go:53: successes=2 duplicates=0
FAIL
FAIL	ticketdesk	0.014s
?   	ticketdesk/cmd/ticketdesk	[no test files]
ok  	ticketdesk/internal/batch	0.015s
?   	ticketdesk/internal/model	[no test files]
ok  	ticketdesk/internal/httpapi	0.007s
ok  	ticketdesk/internal/report	0.001s
ok  	ticketdesk/internal/store	0.009s
ok  	ticketdesk/internal/validate	0.001s
ok  	ticketdesk/internal/worker	0.008s
FAIL
```

## Architecture reproduction

### linux/amd64
- Go toolchain version: exit `0`
- Go build (.): exit `0`
- Go test (.): exit `1`
- Go run smoke (cmd/ticketdesk): exit `0`
### linux/arm64
- Go toolchain version: exit `0`
- Go build (.): exit `0`
- Go test (.): exit `1`
- Go run smoke (cmd/ticketdesk): exit `0`
