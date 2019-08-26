package cmd

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"
)

func startHttp(closed <-chan struct{}, wg *sync.WaitGroup, listen url.URL, feedmap FeedMap) {
	defer wg.Done()

	port, err := strconv.Atoi(listen.Port())
	if err != nil {
		panic("Error Converting port into int")

	}

	wg.Add(1)
	fmt.Printf("\n Listening on :%d\n", port)
	srv := startHttpServer(wg, port, feedmap)

	for {
		select {
		case <-closed:
			fmt.Printf("Starting to close HTTP SERVER %v\n", wg)
			if err := srv.Shutdown(context.TODO()); err != nil {
				log.Panicf("failure/timeout shutting down the http server gracefully: %v", err)
			}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second) //TODO make configurable
			defer cancel()

			srv.SetKeepAlivesEnabled(false)
			if err := srv.Shutdown(ctx); err != nil {
				log.Fatalf("Could not gracefully shutdown the server: %v\n", err)
			}
			fmt.Printf("Exiting START HTTP SERVER %v\n", wg)
			return
		default:
		} // select
	} // for

} // startHttp

//mux := http.NewServeMux()
//mux.Handler("/request", requesthandler)
//http.ListenAndServe(":9000", nil)

func startHttpServer(wg *sync.WaitGroup, port int, feedmap FeedMap) *http.Server {
	defer wg.Done()
	addr := fmt.Sprintf(":%d", port)
	srv := &http.Server{Addr: addr}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { muxingHandler(w, r, feedmap) })

	wg.Add(1)
	go func() {
		defer wg.Done()

		// returns ErrServerClosed on graceful close
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			//https://stackoverflow.com/questions/39320025/how-to-stop-http-listenandserve
			log.Fatalf("ListenAndServe(): %s", err)
		}
		fmt.Printf("Exiting HTTPServer %v\n", wg)
	}()

	// returning reference so caller can call Shutdown()
	return srv
}

func muxingHandler(w http.ResponseWriter, r *http.Request, feedmap FeedMap) {
	fmt.Printf("\n-------------------------------------------------------------------------------------\nhttp://handler called\n")
	if channelSlice, ok := feedmap[r.URL.Path]; !ok {
		fmt.Printf(`\n*****************************************************************************\n
Not going to send this stream anywhere so goodbye\n
*********************************************\n`)
		return
	} else {

		var b bytes.Buffer // A Buffer needs no initialization.
		const chunkSize = 1024000
		chunk := make([]byte, chunkSize)
		//b.Write([]byte("Hello "))
		//fmt.Fprintf(&b, "world!")
		//b.WriteTo(os.Stdout)
		i := 0
		for {
			n, err := io.ReadFull(r.Body, chunk)
			if err != nil {
				return //assume capture has stopped
			}
			//fmt.Printf("Chunk %d", i)
			i = i + 1
			b.Write(chunk[:n])
			if n < chunkSize {
				fmt.Println("\nunderead\n")
			}
			m := b.Len()
			fragment := make([]byte, m)
			b.Read(fragment)
			packet := Packet{Data: fragment}
			fmt.Printf("Received %d\n", m)
			for _, channel := range channelSlice {
				channel <- packet
			}
			//}
			b.Reset()

		}
	}
}

//	buf, err := ioutil.ReadAll(r.Body)
//	if err != nil {
//		log.Fatal("request", err)
//		fmt.Printf("http://error with data %v", err)
//	}
//	packet := Packet{Data: buf}
//	fmt.Printf("http://got data to send\n")
//	if channelSlice, ok := feedmap[r.URL.Path]; ok {
//		for _, channel := range channelSlice {
//			channel <- packet
//			fmt.Printf("http://sent that data :-)")
//		}
//	} else {
//		fmt.Printf("didn't find %s in feedmap", r.URL.Path)
//	}
//}
