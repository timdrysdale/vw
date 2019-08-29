package cmd

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
)

var filename string

func init() {
	rootCmd.AddCommand(checkCmd)
	checkCmd.Flags().StringVar(&filename, "file", "", "file with TS stream in 188-byte chunk format")
}

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "count and report TS packets between delimiters",
	Long:  `Reports the number of packets betwen delimiters that are inserted after every write by file writers (set debug=true in writer when recording)`,
	Run: func(cmd *cobra.Command, args []string) {

		if filename == "" {
			fmt.Println("Please specify a file (--file=<file>)")
			return
		}

		delim, err := hex.DecodeString("800f20008047") //47 is seen at end due to ReadBytes getting the next sync
		fmt.Printf("Deliminator %v\n", delim)
		syncTS := byte('G')

		file, err := os.Open(filename)
		check(err)

		defer file.Close()

		r := bufio.NewReader(file)

		//ignore first line
		_, err = r.ReadBytes(syncTS)
		if err != nil {
			if err != io.EOF {
				fmt.Println(err)
			}
			fmt.Printf("Reached EOF without finding a sync\n")
			return
		}
		frameSize := 0
		badFrames := 0
		delims := 0

		for {

			line, err := r.ReadBytes(syncTS)
			if err != nil {
				if err != io.EOF {
					fmt.Println(err)
				}
				break
			}
			if len(line) == 6 {
				if bytes.Equal(line, delim) {
					delims = delims + 1
					if frameSize%188 != 0 {
						badFrames = badFrames + 1
					}
					frameSize = 0
				}
			} else {
				frameSize = frameSize + len(line) //keep track of bytes read since last delimiter
			}

		}
		fmt.Printf("Found %d deliminators\n", delims)
		fmt.Printf("Estimate %d bad frames\n", badFrames)
	},
}
