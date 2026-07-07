# opencode-go

Go SDK for the [opencode server](https://opencode.ai/docs/server/) HTTP API.

Zero dependencies (standard library only). Types are hand-written against the
server's OpenAPI 3.1 spec (`GET /doc`, opencode 1.15.x).

## Install

```sh
go get github.com/orvice/opencode-go
```

## Quick start

Start a server:

```sh
opencode serve --port 4096
```

Create a session and send a prompt:

```go
package main

import (
	"context"
	"fmt"
	"log"

	opencode "github.com/orvice/opencode-go"
)

func main() {
	ctx := context.Background()

	client, err := opencode.NewClient() // defaults to http://127.0.0.1:4096
	if err != nil {
		log.Fatal(err)
	}

	sess, err := client.Session.Create(ctx, &opencode.SessionCreateParams{
		Title: "my session",
	})
	if err != nil {
		log.Fatal(err)
	}

	resp, err := client.Session.PromptText(ctx, sess.ID, "Explain this codebase")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(resp.Text())
}
```

## Configuration

```go
client, err := opencode.NewClient(
	opencode.WithBaseURL("http://127.0.0.1:4096"),
	opencode.WithPassword("secret"),            // HTTP basic auth; defaults to $OPENCODE_SERVER_PASSWORD
	opencode.WithDirectory("/path/to/project"), // default `directory` query param on every request
)
```

When `OPENCODE_SERVER_PASSWORD` is set the client authenticates automatically;
the username defaults to `opencode` (override with `WithUsername` or
`$OPENCODE_SERVER_USERNAME`).

## Events (SSE)

```go
stream, err := client.Event.Subscribe(ctx) // GET /event
if err != nil {
	log.Fatal(err)
}
defer stream.Close()

for stream.Next() {
	ev := stream.Current()
	switch ev.Type {
	case opencode.EventTypeMessagePartDelta:
		var p opencode.EventMessagePartDeltaProperties
		ev.DecodeProperties(&p)
		fmt.Print(p.Delta)
	case opencode.EventTypeSessionIdle:
		return
	}
}
if err := stream.Err(); err != nil {
	log.Fatal(err)
}
```

`client.Global.Events(ctx)` subscribes to `/global/event`, which multiplexes
events from all server instances.

Combine with `Session.PromptAsync` for streaming responses: send the prompt
asynchronously, then render `message.part.delta` events until `session.idle`.
See [examples/basic](examples/basic/main.go).

## Services

| Service | Endpoints |
|---------|-----------|
| `client.Session` | create/list/get/update/delete, children, todo, status, diff, init, fork, abort, share, summarize, revert, permissions, messages, **Prompt / PromptAsync / Command / Shell** |
| `client.Event` / `client.Global` | `/event`, `/global/event` SSE streams, `/global/health` |
| `client.Config` | get, patch, providers |
| `client.Provider` / `client.Auth` | provider list, auth methods, OAuth flow, set/remove credentials |
| `client.Find` | text search (ripgrep), fuzzy file search, workspace symbols |
| `client.File` | list directory, read file, git status |
| `client.App` | agents, commands, log, path, VCS info |
| `client.Project` | list, current |
| `client.TUI` | prompt control, toasts, dialogs, control protocol |
| `client.MCP` / `client.LSP` / `client.Formatter` / `client.Tool` | MCP status/add/connect, LSP status, formatter status, experimental tool registry |
| `client.Instance` | dispose |

## Errors

Non-2xx responses are returned as `*opencode.Error`:

```go
if _, err := client.Session.Get(ctx, "ses_x"); err != nil {
	var apiErr *opencode.Error
	if errors.As(err, &apiErr) {
		fmt.Println(apiErr.StatusCode, apiErr.Message)
	}
}
```

## Testing

```sh
go test ./...
```

Integration tests run against a real server when `OPENCODE_INTEGRATION_URL` is set:

```sh
opencode serve --port 4096 &
OPENCODE_INTEGRATION_URL=http://127.0.0.1:4096 go test -run TestIntegration -v
```

## License

MIT
