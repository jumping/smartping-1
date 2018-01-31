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
	"time"

	"github.com/spf13/cobra"
)

// pingCmd represents the ping command
var pingCmd = &cobra.Command{
	Use:   "ping",
	Short: "Ping the server",
	Run: func(cmd *cobra.Command, args []string) {
		Ping(&opts)
	},
}

// PingOptions is the options to configure ping behavior
type PingOptions struct {
	Server string
	Min    time.Duration
	Max    time.Duration
	Factor float64
}

var opts PingOptions

func init() {
	RootCmd.AddCommand(pingCmd)

	pingCmd.Flags().StringVarP(&opts.Server, "server", "s", "127.0.0.1:8804", "server address to ping")
	pingCmd.Flags().DurationVarP(&opts.Min, "min", "m", 1*time.Second, "the min ping interval")
	pingCmd.Flags().DurationVarP(&opts.Max, "max", "M", 60*time.Minute, "the max ping interval")
	pingCmd.Flags().Float64VarP(&opts.Factor, "factor", "f", 2, "the growth factor of hearbeat")
}

func Ping(opts *PingOptions) {
	heartbeat := opts.Min
	factor := opts.Factor
	PING := []byte("PING")
	PONG := []byte("PONG")
	buf := make([]byte, 4)
	seq := 0

	left := opts.Min
	right := opts.Max

	dial := func() net.Conn {
		for {
			conn, err := net.Dial("tcp", opts.Server)
			if err != nil {
				log.Println("ERROR dial failed:", err)
				time.Sleep(time.Second)
				continue
			}
			return conn
		}
	}
	redial := func(old net.Conn) net.Conn {
		if err := old.Close(); err != nil {
			log.Println("ERROR close failed:", err)
		}
		return dial()
	}

	conn := dial()
	for {
		tick := time.NewTicker(heartbeat)
		select {
		case <-tick.C:
			seq += 1
			begin := time.Now()
			conn.SetWriteDeadline(begin.Add(heartbeat))
			log.Printf("INFO Send PING(%v), heartbeat:%v, left:%v, right:%v, factor:%v\n", seq, heartbeat, left, right, factor)
			n, e := conn.Write(PING)
			if e != nil {
				log.Println("ERROR ping ", e)

				conn = redial(conn)

				right = heartbeat
				next := time.Duration(float64(left+right) / factor)
				heartbeat = next

				tick.Stop()
				continue
			}
			if n != len(PING) {
				log.Println("WARN write corrupted")
				tick.Stop()
				continue
			}

			conn.SetReadDeadline(time.Now().Add(heartbeat))
			n, e = conn.Read(buf)
			if e == io.EOF {
				log.Println("INFO Server close the connection")
				tick.Stop()
				return
			}

			if e != nil {
				log.Println("ERROR pong ", e)
				conn = redial(conn)
				right = heartbeat
				next := time.Duration(float64(left+right) / factor)
				heartbeat = next
				tick.Stop()
				continue
			}
			if bytes.Compare(buf, PONG) != 0 {
				log.Println("WARN got response unexpected")
				tick.Stop()
				continue
			}
			cost := time.Now().Sub(begin)
			log.Printf("INFO Recv PONG(%v), heartbeat:%v, left:%v, rigth:%v, factor:%v, cost:%v\n", seq, heartbeat, left, right, factor, cost)

			left = heartbeat
			next := time.Duration(float64(left) * factor)
			heartbeat = next
			if heartbeat > right {
				heartbeat = right
			}
			if left == right {
				right = opts.Max
			}
			tick.Stop()
		}
	}
}
