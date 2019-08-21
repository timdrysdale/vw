package cmd

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"sync"

	"github.com/phayes/freeport"
	"github.com/spf13/cobra"
	"nhooyr.io/websocket"
)

// TODO:
//     * how should we handle messages exceeding bufsize?

var audioCommand string
var bufsize int64
var destination string
var source string
var verbose bool
var videoCommand string

func init() {
	rootCmd.PersistentFlags().StringVarP(&audioCommand, "audioCommand", "-a", "", "Command to obtain audio stream, including the token `stream_to_vw`")
	rootCmd.MarkFlagRequired("audioCommand")

	rootCmd.PersistentFlags().Int64VarP(&bufsize, "bufsize", "b", 65535, "buffer size (max message size) [DEFAULT is 65535 bytes]")

	rootCmd.PersistentFlags().StringVarP(&destination, "destination", "d", "", "ws[s]://<ip>:<port> of the destination websocket server")
	rootCmd.MarkFlagRequired("destination")

	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "print connection and message logs [DEFAULT is quiet]")

	rootCmd.PersistentFlags().StringVarP(&videoCommand, "videoCommand", "-v", "", "Command to obtain video stream, including the token `stream_to_vw`")
	rootCmd.MarkFlagRequired("videoCommand")

	port, err := freeport.GetFreePort()
	if err != nil {
		log.Fatal(err)
	}

	source = strings.Join([]string{"http://127.0.0.1:", port.string()}, "")
	audioCommand = strings.Replace(audioCommand, "stream_to_vw", source, -1) //allow multiple replacements to avoid surprising users
	videoCommand = strings.Replace(videoCommand, "stream_to_vw", source, -1) //allow multiple replacements to avoid surprising users

}

var rootCmd = &cobra.Command{
	Use:   "vw",
	Short: "VW video websockets transporter (geddit)",
	Long:  `VW initialises video and audio capture via syscall, receives stream via http to avoid pipe latency issues, then forwards to a websocket server`,
	Run: func(cmd *cobra.Command, args []string) {

		// see https://www.alexedwards.net/blog/validation-snippets-for-go#url-validation)
		d, err := url.Parse(destination)
		if err != nil {
			panic(err)
		} else if d.Scheme == "" || d.Host == "" {
			fmt.Println("error: destination must be an absolute URL")
			return
		} else if d.Scheme != "ws" && d.Scheme != "wss" {
			fmt.Println("error: destination must begin with ws or wss")
			return
		}

		if verbose {
			fmt.Printf("audio: %v\nvideo: %v\n", audioCommand, videoCommand)
			fmt.Printf("source: %v\ndestination: %v", source, destination)

		}

		var wg sync.WaitGroup
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		msg := make(chan []byte)
		closed := make(chan struct{})

		go func() {
			for _ = range c {

				close(closed)
				wg.Wait()
				os.Exit(1)

			}
		}()

		wg.Add(1)
		go HandleSource(closed, msg, &wg, t)
		go HandleDestination(closed, msg, &wg, r)
		wg.Wait()
	},
}

func Execute() {

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func HandleDestination(closed <-chan struct{}, msg chan<- []byte, wg *sync.WaitGroup, t *url.URL) {
	defer wg.Done()

	var buf = make([]byte, bufsize)
	var n int

	if verbose {
		fmt.Println("connecting to", t.String())
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	c, _, err := websocket.Dial(ctx, t.String(), websocket.DialOptions{})

	if err != nil {
		fmt.Println(err)
		return
	}

	defer c.Close(websocket.StatusInternalError, fmt.Sprintf("Internal error with incoming websocket %s", t.String()))

	for {
		select {
		default:

			typ, r, err := c.Reader(ctx)

			if err != nil {

				fmt.Println("tx: io.Reader", err)
				return
			}

			if typ != websocket.MessageBinary {
				fmt.Println("Not binary")
				//return
			}

			n, err = r.Read(buf)
			if verbose {
				fmt.Println("Got: ", n)
			}
			if err != nil {
				if err != io.EOF {
					fmt.Println("Read:", err)
				}

			}

			msg <- buf[:n]

		case <-closed:
			c.Close(websocket.StatusNormalClosure, "")
			return
		}
	}
}

func HandleSource(closed <-chan struct{}, msg <-chan []byte, wg *sync.WaitGroup, t *url.URL) {
	//	for pkt := range msg {
	//		fmt.Println(len(pkt))
	//	}
	//}

	defer wg.Done()

	var n int

	if servers {

		fn := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			// just a normal http thing here, but binary.
			c, err := websocket.Accept(w, r, websocket.AcceptOptions{InsecureSkipVerify: true})
			if err != nil {
				log.Println(err)
				return
			}
			defer c.Close(websocket.StatusInternalError, "the sky is falling")

			ctx, cancel := context.WithCancel(r.Context())
			ctx = c.CloseRead(ctx) //since we won't read
			defer cancel()

			for {
				select {
				case buf := <-msg:

					w, err := c.Writer(ctx, websocket.MessageBinary)

					if err != nil {
						fmt.Println("io.Writer", err)
						return
					}
					if verbose {
						fmt.Println("buf size is", len(buf))
					}

					n, err = w.Write(buf)

					if verbose {
						fmt.Println("wrote buf of length", n)
					}

					if n != len(buf) {
						fmt.Println("Mismatch write lengths, overflow?")
						return
					}

					if err != nil {
						if err != io.EOF {
							fmt.Println("Write:", err)
						}
					}

					err = w.Close() // do every write to flush frame
					if verbose {
						fmt.Println("closed writer")
					}
					if err != nil {
						fmt.Println("Closing Write failed:", err)
					}

				case <-closed:
					fmt.Println("Been told to finish up")
					c.Close(websocket.StatusNormalClosure, "")
					return
				}
			}
		})
		addr := strings.Join([]string{t.Hostname(), ":", t.Port()}, "")
		log.Printf("Starting listener on %s\n", addr)
		err := http.ListenAndServe(addr, fn)
		log.Fatal(err)
	} else {
		if verbose {
			fmt.Println("connecting to", t.String())
		}
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		c, _, err := websocket.Dial(ctx, t.String(), websocket.DialOptions{})

		if err != nil {
			fmt.Println(err)
			return
		}

		defer c.Close(websocket.StatusInternalError, fmt.Sprintf("Internal error with incoming websocket %s", t.String()))

		ctx = c.CloseRead(ctx) //since we won't read

		for {
			select {
			case buf := <-msg:

				w, err := c.Writer(ctx, websocket.MessageBinary)

				if err != nil {
					fmt.Println("io.Writer", err)
					return
				}
				if verbose {
					fmt.Println("buf size is", len(buf))
				}

				n, err = w.Write(buf)

				if verbose {
					fmt.Println("wrote buf of length", n)
				}

				if n != len(buf) {
					fmt.Println("Mismatch write lengths, overflow?\n")

				}

				if err != nil {
					if err != io.EOF {
						fmt.Println("Write:", err)
					}
				}

				err = w.Close() // do every write to flush frame
				if verbose {
					fmt.Println("closed writer")
				}
				if err != nil {
					fmt.Println("Closing Write failed:", err)
				}

			case <-closed:
				c.Close(websocket.StatusNormalClosure, "")
				return
			}
		}

	}
}

//func HandleReceiverOld(msg <-chan []byte, wg *sync.WaitGroup, server bool, bufsize int64, t *url.URL) {
//	// NOTE that nhooyr websockets must read messages sent to them else
//	// control frames are not processed ... so bidirectional is needed
//	defer wg.Done()
//
//	//var buf = make([]byte, 65535+1) //1000kbps rate at 30fps is just over 4200byte/s
//	var n int
//
//	//ctx, cancel := context.WithTimeout(context.Background(), 18*time.Second)
//	ctx, cancel := context.WithCancel(context.Background())
//	defer cancel()
//
//	c, _, err := websocket.Dial(ctx, t.String(), websocket.DialOptions{})
//	c.SetReadLimit(bufsize)
//	if err != nil {
//		fmt.Println(err)
//		return
//	}
//
//	defer c.Close(websocket.StatusInternalError, fmt.Sprintf("Internal error with websocket client %s", t.String()))
//	fmt.Println("In Handle Receiver")
//
//	//assume we don't get any messages so we must keep extending the readDeadline
//	//ticker := time.NewTicker(30 * time.Second)
//	//go func() {
//	//	for t := range ticker.C {
//	//		//ctx.SetReadDeadline(time.Minute)
//	//		c.setReadTimeout <- context.Background()
//	//	}
//	//}()
//	// read but ignore any messages we do receive
//	//go func() {
//	//	typ, r, err := c.Reader(ctx)
//	//
//	//	if err != nil {
//	//
//	//		fmt.Println("io.Reader", err)
//	//		return
//	//	}
//	//
//	//	if typ != websocket.MessageBinary {
//	//		fmt.Println("Not binary")
//	//		//return
//	//	}
//	//
//	//	n, err = r.Read(buf)
//	//	if err != nil {
//	//		if err != io.EOF {
//	//			fmt.Println("Read:", err)
//	//		}
//	//
//	//	} else {
//	//		fmt.Println("Got: ", n)
//	//	}
//	//
//	//}()
//
//	for pkt := range msg {
//		fmt.Printf("%T %d\n", pkt, len(pkt))
//		fmt.Println("attempting to read from channel")
//		//func (c *Conn) Writer(ctx context.Context, typ MessageType) (io.WriteCloser, error)
//		w, err := c.Writer(ctx, websocket.MessageBinary)
//		fmt.Println("got writer")
//		if err != nil {
//
//			fmt.Println("io.Writer", err)
//			return
//		}
//		fmt.Println("pkt size is", len(pkt))
//		//nn := int64(math.Min(float64(len(pkt)), float64(65535)))
//		n, err = w.Write(pkt[:511])
//
//		fmt.Println("wrote pkt length", n)
//
//		if err != nil {
//			if err != io.EOF {
//				fmt.Println("Write:", err)
//			}
//		}
//
//		err = w.Close()
//		fmt.Println("closed writer")
//		if err != nil {
//			fmt.Println("Closing Write failed:", err)
//		}
//
//	}
//
//	c.Close(websocket.StatusNormalClosure, "")
//	return
//
//}
