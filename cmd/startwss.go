/*
   reconws is websocket client that automatically reconnects
   Copyright (C) 2019 Timothy Drysdale <timothy.d.drysdale@gmail.com>

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU Affero General Public License as
   published by the Free Software Foundation, either version 3 of the
   License, or (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU Affero General Public License for more details.

   You should have received a copy of the GNU Affero General Public License
   along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

package reconws

import (
	"math/rand"
	"net/url"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/jpillora/backoff"
	log "github.com/sirupsen/logrus"
)

type WsMessage struct {
	Data []byte
	Type int
}

// connects (retrying/reconnecting if necessary) to websocket server at url

type ReconWs struct {
	Url             string
	In              chan WsMessage
	Out             chan WsMessage
	Stop            chan struct{}
	Wg              *sync.WaitGroup
	ConnectedAt     time.Time
	ForwardIncoming bool
	Err             error

	Retry RetryConfig
	//Stats?

	// internal usage only
	close     chan struct{}
	connected chan struct{}
}

type RetryConfig struct {
	Factor  float64
	Min     time.Duration
	Max     time.Duration
	Timeout time.Duration
}

func New() *ReconWs {
	r := &ReconWs{
		Url:             "",
		In:              make(chan WsMessage),
		Out:             make(chan WsMessage),
		close:           make(chan struct{}),
		Stop:            make(chan struct{}),
		ForwardIncoming: true,
		Err:             nil,
		Retry: RetryConfig{Factor: 1.5,
			Min:     time.Second,
			Max:     60 * time.Second,
			Timeout: 5 * time.Second,
			Jitter:  true},
	}
	return r
}

// run this in a separate goroutine so that the connection can be
// ended from where it was initialised, by close((* ReconWs).Stop)
func (r *ReconWs) Reconnect() {

	boff := &backoff.Backoff{
		Min:    r.Retry.Min,
		Max:    r.Retry.Max,
		Factor: r.Retry.Factor,
		Jitter: r.Retry.Jitter,
	}

	rand.Seed(time.Now().UTC().UnixNano())

	// try dialling ....

	r.connected = make(chan struct{}) //reset our connection indicator

	for {

		nextRetryWait = boff.Duration()

		//refresh these indicators each attempt
		running := make(chan struct{})
		stopped := make(chan struct{})

		go func() {
			r.Reconnect()  //returns on error or if closed with close(r.Close)
			close(stopped) //signal there was an issue
		}()

		select {

		case <-r.connected: //good!
			boff.Reset()
			//let the connection operate until an issue arises:
			select {

			case <-stopped: // there was a problem with the connection
				continue //try reconnecting

			case <-r.Stop: // requested to stop
				close(r.close) //close the connection (don't wait)
				return         //don't try reconnecting
			}

		case <-r.Stop: //requested to stop
			close(r.close)
			return //don't try reconnecting

		case <-time.After(r.HandshakeTimeout):
			// handshake too slow, close and retry
			close(r.Close)
			continue
		}

		time.Sleep(nextRetryWait)

	}

}

// Dial the websocket server once.
// If dial fails then return immediately
// If dial succeeds then handle message traffic until
// (r *ReconWs).Close is closed (presumably by the caller)
// or if there is a reading or writing error
func (r *ReconWs) Dial() {

	var err error

	defer func() {
		r.Wg.Done()
		r.Err = err
	}()

	if r.Url == "" {
		log.Error("Can't dial an empty Url")
		return
	}

	u, err := url.Parse(r.Url)

	if err != nil {
		log.Error("Url:", err)
		return
	}

	if u.Scheme != "ws" && u.Scheme != "wss" {
		log.Error("Url needs to start with ws or wss")
		return
	}

	if u.User != nil {
		log.Error("Url can't contain user name and password")
		return
	}

	// start dialing ....

	log.WithField("To", u).Debug("Connecting")

	c, _, err := websocket.DefaultDialer.Dial(u, nil)

	if err != nil {
		log.WithField("error", err).Error("Dialing")
		return
	}

	defer c.Close()

	r.ConnectedAt = time.Now()
	close(r.connected) //signal that we've connected

	log.WithField("To", u).Info("Connected")

	// handle our reading tasks

	done := make(chan struct{})

	go func() {
		defer close(done) // signal to writing task to exit if we exit first
		for {

			data, mt, err := c.ReadMessage()

			// Check for errors, e.g. caused by writing task closing conn
			// because we've been instructed to exit
			if err != nil {
				log.WithField("error", err).Error("Reading")
				return
			}
			// optionally forward messages
			if r.ForwardIncoming {
				r.In <- WsMessage{Data: data, Type: mt}
			}
		}
	}()

	// handle our writing tasks

	for {
		select {
		case <-done:
			return

		case msg := <-r.Out:

			err := c.WriteMessage(msg.Type, msg.Data)
			if err != nil {
				log.WithField("error", err).Fatal("Writing")
				return
			}
		case <-r.close: // r.HandshakeTimout has probably expired

			// Cleanly close the connection by sending a close message
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.WithField("error", err).Error("Closing")
			} else {
				log.WithField("name", name).Info("Closed")
			}
			return
		}
	}
}
