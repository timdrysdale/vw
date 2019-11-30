package cmd

import (
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/timdrysdale/agg"
	"github.com/timdrysdale/hub"
	"github.com/timdrysdale/rwc"
)

func (app *App) internalAPI(topic string) {

	c := &hub.Client{Hub: app.Hub.Hub,
		Name:  "admin",
		Send:  make(chan hub.Message),
		Stats: hub.NewClientStats(),
		Topic: topic,
	}

	app.Hub.Register <- c

	for {
		select {
		case message, ok := <-c.Send:

			if !ok {
				// The hub closed the channel.
				return
			}

			reply, err := app.handleAdminMessage(message.Data)

			if err == nil {
				c.Hub.Broadcast <- hub.Message{Sender: *c, Data: reply, Type: websocket.TextMessage, Sent: time.Now()} //mmmm type needed here == too much coupling ...!!
			} else {
				c.Hub.Broadcast <- hub.Message{Sender: *c, Data: []byte(err.Error()), Type: websocket.TextMessage, Sent: time.Now()}
			}

		case <-app.Closed:
			return
		}
	}
}

type Command struct {
	Verb  string
	What  string
	Which string
	Rule  *json.RawMessage
}

type RuleStream struct {
	Stream string
	Feeds  []string
}

var errBadCommand = errors.New("Unrecognised Command")

// JSON API - note change to singular stream and destination
//
// {"verb":"add","what":"destination","rule":{"stream":"video0","destination":"wss://<some.relay.server>/in/video0","id":"0"}}
// {"verb":"add","what":"stream","rule":{"stream":"video0","feeds":["video0","audio0"]}}
//
// {"verb":"list","what":"stream","which":"<name>"}
// {"verb":"list","what":"destination","which":"<id>">}
//
// {"verb":"list","what":"stream","which":"all"}
// {"verb":"list","what":"destination","which":"all"}
//
// {"verb":"delete","what":"stream","which":"<which>"}
// {"verb":"delete","what":"destination","which":"<id>">}
//
// {"verb":"delete","what":"stream","which":"all"}
// {"verb":"delete","what":"destination","which":"all"}

func (app *App) handleAdminMessage(msg []byte) ([]byte, error) {

	var cmd Command //map[string]*json.RawMessage
	var reply []byte

	err := json.Unmarshal(msg, &cmd)

	switch cmd.What {
	case "destination":
		switch cmd.Verb {
		case "add":
			var rule rwc.Rule
			err = json.Unmarshal(*cmd.Rule, &rule)
			rule.Stream = strings.TrimPrefix(rule.Stream, "/") //to match trimming we do in handleStreamAdd
			app.Websocket.Add <- rule
			reply, err = json.Marshal(rule)
		case "delete":
			switch cmd.Which {
			case "":
				err = errBadCommand
			case "all":
				app.Websocket.Delete <- "deleteAll"
				reply, err = json.Marshal("deleteAll")
			default:
				app.Websocket.Delete <- cmd.Which
				reply, err = json.Marshal(cmd.Which)
			}
		case "list":
			switch cmd.Which {
			case "":
				err = errBadCommand
			case "all":
				reply, err = json.Marshal(app.Websocket.Rules)
			default:
				reply, err = json.Marshal(app.Websocket.Rules[cmd.Which])
			}
		default:
			err = errBadCommand
		}
	case "stream":
		switch cmd.Verb {
		case "add":
			var rule agg.Rule
			err = json.Unmarshal(*cmd.Rule, &rule)
			rule.Stream = strings.TrimPrefix(rule.Stream, "/") //to match trimming we do in handleStreamAdd
			app.Hub.Add <- rule
			reply, err = json.Marshal(rule)
		case "delete":
			switch cmd.Which {
			case "":
				err = errBadCommand
			case "all":
				app.Hub.Delete <- "deleteAll"
				reply, err = json.Marshal("deleteAll")
			default:
				app.Hub.Delete <- cmd.Which
				reply, err = json.Marshal(cmd.Which)
			}
		case "list":
			switch cmd.Which {
			case "":
				err = errBadCommand
			case "all":
				reply, err = json.Marshal(app.Hub.Rules)
			default:
				reply, err = json.Marshal(app.Hub.Rules[cmd.Which])
			}
		default:
			err = errBadCommand
		}
	default:
		err = errBadCommand
	}

	return reply, err

}

// REST-like API
// destination: POST {"stream":"video0","destination":"wss://<some.relay.server>/in/video0","id":"0"} /api/destinations
// stream: POST {"stream":"/stream/front/large","feeds":["video0","audio0"]} /api/streams
// GET /api/streams/all
// GET /api/destinations/all
// DELETE /api/streams</stream_name>
// DELETE /api/destinations</id>
// DELETE /api/streams/all
// DELETE /api/destinations/all
