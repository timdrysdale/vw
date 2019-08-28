package cmd

import (
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/google/uuid"
)

func startWriters(closed <-chan struct{}, wg *sync.WaitGroup, writers ToFile, clientActionsChan chan clientAction) {
	defer wg.Done()
	for _, writer := range writers.Writers {
		wg.Add(1)
		name := "fileWriter(" + uuid.New().String()[:3] + "):"
		go writerClient(closed, wg, writer, name, clientActionsChan)
		log.Printf("%s spawned", name)
	}
}

func writerClient(closed <-chan struct{}, wg *sync.WaitGroup, writer Writer, name string, clientActionsChan chan clientAction) {
	//debugDelimiter provides an unambiguous marker in the packet stream, being a packet you would not expect to see ordinarily
	//Thus SYNC,Transport Error Indicator,ReservedPID of 15,Adaptation Byte Only,No Extra Fields, Discontinuity state
	//https://en.wikipedia.org/wiki/MPEG_transport_stream#Packet_identifier_(PID)
	debugDelimiter, err := hex.DecodeString("47800f200080")
	if err != nil {
		fmt.Printf("ERROR: debugDelimeter in startwriter.go/writerClient() is bad\n")
	}

	defer wg.Done()
	messagesForMe := make(chan message, 10)

	for i, input := range writer.InputNames {

		client := clientDetails{name: name, topic: input, messagesChan: messagesForMe}
		fmt.Printf("\n%d: %s subscribing to %s\n", i, name, input)
		clientActionsChan <- clientAction{action: clientAdd, client: client}

		defer func() {
			clientActionsChan <- clientAction{action: clientDelete, client: client}
			fmt.Printf("Disconnected %v, deleting from topics\n", client)
		}()
	}
	for {
		_, err := os.Stat(writer.File)
		if err == nil {
			log.Fatalf("%s found %s already exists\n", name, writer.File)
			return
		}
		file, err := os.Create(writer.File)
		if err != nil {
			log.Fatalf("%s could not open %s for writing\n", name, writer.File)
			return
		}
		defer file.Close()
		if writer.Debug == true {
			//give an indication at start of file that it has been
			//modified with delimiters
			_, err := file.Write(debugDelimiter)
			if err != nil {
				log.Fatalf("%s failed writing to %s because %s\n", name, writer.File, err)
				return
			}
		}
		for {
			select {
			case <-closed:
				return
			case msg := <-messagesForMe:
				_, err := file.Write(msg.data)
				if err != nil {
					log.Fatalf("%s failed writing to %s because %s\n", name, writer.File, err)
					return
				}
				if writer.Debug == true {
					_, err := file.Write(debugDelimiter)
					if err != nil {
						log.Fatalf("%s failed writing to %s because %s\n", name, writer.File, err)
						return
					}
				}
			}
		}
	}
}
