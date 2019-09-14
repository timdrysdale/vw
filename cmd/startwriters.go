package cmd

import (
	"os"
	"sync"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"github.com/timdrysdale/hub"
)

func startWriters(closed <-chan struct{}, wg *sync.WaitGroup, writers ToFile, h *hub.Hub) {
	defer wg.Done()
	for _, writer := range writers.Writers {
		wg.Add(1)
		name := "fileWriter(" + uuid.New().String()[:3] + "):"
		go writerClient(closed, wg, writer, name, h)
		log.Printf("%s spawned", name)
	}
}

func writerClient(closed <-chan struct{}, wg *sync.WaitGroup, writer Writer, name string, h *hub.Hub) {
	//debugDelimiter provides an unambiguous marker in the packet stream, being a packet you would not expect to see ordinarily
	//Thus SYNC,Transport Error Indicator,ReservedPID of 15,Adaptation Byte Only,No Extra Fields, Discontinuity state
	//https://en.wikipedia.org/wiki/MPEG_transport_stream#Packet_identifier_(PID)
	debugDelimiter := []byte("<TIMOTHY>") //hex.DecodeString("47800f200080")
	//if err != nil {
	//	fmt.Printf("ERROR: debugDelimeter in startwriter.go/writerClient() is bad\n")
	//}
	//
	defer wg.Done()
	messagesForMe := make(chan hub.Message, 10)

	for _, input := range writer.InputNames {

		client := &hub.Client{Hub: h,
			Name:  name,
			Send:  messagesForMe,
			Stats: hub.NewClientStats(),
			Topic: input}

		log.WithFields(log.Fields{"name": name, "input": input}).Info("Subscribing")
		h.Register <- client

		defer func() {
			h.Unregister <- client
			log.WithField("name", name).Fatal("Disconnected")
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
				_, err := file.Write(msg.Data)
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
