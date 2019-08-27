package cmd

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/google/uuid"
)

func HandleConnections(closed <-chan struct{}, wg *sync.WaitGroup, clientActionsChan chan clientAction, messagesFromMe chan message, host *url.URL) {
	defer wg.Done()

	fn := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		var HTTPUpgrader ws.HTTPUpgrader

		// chrome needs this to connect
		HTTPUpgrader.Protocol = func(str string) bool { return true }

		conn, _, _, err := HTTPUpgrader.Upgrade(r, w)

		if err != nil {
			log.Fatalf("WS upgrade failed because %v\n", err)
			return
		}

		//subscribe this new client
		messagesForMe := make(chan message, 2)
		var name = uuid.New().String()
		var topic = r.URL.Path

		client := clientDetails{name: name, topic: topic, messagesChan: messagesForMe}

		clientActionsChan <- clientAction{action: clientAdd, client: client}

		defer func() {
			clientActionsChan <- clientAction{action: clientDelete, client: client}
			fmt.Printf("Disconnected %v, deleting from topics\n", client)
		}()

		var localWG sync.WaitGroup

		localWG.Add(2)
		// read from client
		go func() {
			defer localWG.Done()
			for {

				msg, op, err := wsutil.ReadClientData(conn)
				if err == nil {
					messagesFromMe <- message{sender: client, op: op, data: msg}
				} else {
					log.Printf("Error on read because %v\n", err)
					return
				}
			}
		}()

		//write to client
		go func() {
			defer conn.Close()
			defer localWG.Done()
			for {
				select {
				case msg := <-messagesForMe:
					err = wsutil.WriteServerMessage(conn, msg.op, msg.data)
					if err != nil {
						log.Printf("Fatal error on write because %v", err)
						return
					}

				case <-closed:
					return
				} //select
			} //for
		}() //func

		localWG.Wait()
	}) //end of fun definition

	addr := strings.Join([]string{host.Hostname(), ":", host.Port()}, "")
	log.Printf("Starting listener on %s\n", addr)
	err := http.ListenAndServe(addr, fn)
	if err != nil {
		log.Fatal(err)
	}

}
