// Command basic demonstrates creating a session, streaming events and
// sending a prompt against a local opencode server.
//
// Usage:
//
//	opencode serve --port 4096
//	go run ./examples/basic
package main

import (
	"context"
	"fmt"
	"log"

	opencode "github.com/orvice/opencode-go"
)

func main() {
	ctx := context.Background()

	client, err := opencode.NewClient()
	if err != nil {
		log.Fatal(err)
	}

	health, err := client.Global.Health(ctx)
	if err != nil {
		log.Fatalf("is the server running? opencode serve --port 4096: %v", err)
	}
	fmt.Printf("connected to opencode %s\n", health.Version)

	sess, err := client.Session.Create(ctx, &opencode.SessionCreateParams{
		Title: "opencode-go example",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("session: %s (%s)\n", sess.ID, sess.Title)
	defer client.Session.Delete(ctx, sess.ID)

	// Subscribe to events, then send the prompt asynchronously and stream
	// the assistant's text deltas as they arrive.
	stream, err := client.Event.Subscribe(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer stream.Close()

	err = client.Session.PromptAsync(ctx, sess.ID, opencode.PromptParams{
		Parts: []opencode.PartInput{opencode.NewTextPart("Say hello in one short sentence.")},
	})
	if err != nil {
		log.Fatal(err)
	}

	for stream.Next() {
		ev := stream.Current()
		switch ev.Type {
		case opencode.EventTypeMessagePartDelta:
			var p opencode.EventMessagePartDeltaProperties
			if err := ev.DecodeProperties(&p); err == nil && p.Field == "text" {
				fmt.Print(p.Delta)
			}
		case opencode.EventTypeSessionIdle:
			fmt.Println()
			return
		case opencode.EventTypeSessionError:
			var p opencode.EventSessionErrorProperties
			ev.DecodeProperties(&p)
			if p.Error != nil {
				log.Fatalf("session error: %s: %s", p.Error.Name, p.Error.Data.Message)
			}
			return
		}
	}
	if err := stream.Err(); err != nil {
		log.Fatal(err)
	}
}
