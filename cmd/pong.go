// Copyright © 2018 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"bytes"
	"io"
	"log"
	"net"

	"github.com/spf13/cobra"
)

type PongOptions struct {
	Listen string
}

var pongOpts PongOptions

// pongCmd represents the pong command
var pongCmd = &cobra.Command{
	Use:   "pong",
	Short: "server side which send pong as response when receive a ping",
	Run: func(cmd *cobra.Command, args []string) {
		Pong(&pongOpts)
	},
}

func init() {
	RootCmd.AddCommand(pongCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// pongCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	pongCmd.Flags().StringVarP(&pongOpts.Listen, "listen", "l", "127.0.0.1:8804", "listen address")

}
func Pong(opts *PongOptions) {
	lis, err := net.Listen("tcp", opts.Listen)
	if err != nil {
		log.Fatalln(err)
	}
	for {
		conn, err := lis.Accept()
		if err != nil {
			if e, ok := err.(*net.OpError); ok && e.Temporary() {
				continue
			}
			log.Fatalln(err)
		}
		log.Println("INFO client connected", conn.RemoteAddr())
		go pong(conn)
	}
}

func pong(c net.Conn) {
	PING := []byte("PING")
	PONG := []byte("PONG")

	buf := make([]byte, 4)
	for {
		n, e := c.Read(buf)
		if e == io.EOF {
			log.Println("INFO client exited", c.RemoteAddr())
			break
		}

		if n != len(buf) {
			log.Println("WARN data corrupted")
			continue
		}
		if bytes.Compare(buf, PING) != 0 {
			log.Println("unexpected data:", string(buf))
			continue
		}

		_, e = c.Write(PONG)
		if e != nil {
			log.Println("ERROR write failed:", e)
			break
		}
	}
}
