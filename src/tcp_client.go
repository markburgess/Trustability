//
// Copyright © Mark Burgess, ChiTek-i (2023)
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
//
// e.g. in one CLI window start the server
//     go run tcp_server.go
// then 
//     go run tcp_client.go
//
// ****************************************************************************

package main

import (
	"flag"
	"net"
	"fmt"
	"os"
	"TT"
	"strings"
)

const (
	HOST = "localhost"
	PORT = "8080"
	TYPE = "tcp"
)

var LATENCY []TT.PromiseHistory

// ***************************************************************

func main() {

	flag.Parse()
	args := flag.Args()

	fmt.Println("Command line with",len(args),"argument(s)")

	var sendbuf string

	if len(args) == 1 {
		sendbuf = args[0]

	} else {
		fmt.Println("No message specified")
		return
	}

	// Prologue

	fmt.Println("We have established a promise protocol with extended conditional dependences, i.e. a connection..")

	//

	var dbname string = "SemanticSpacetime"
	var url string = "http://localhost:8529"
	var user string = "root"
	var pwd string = "mark"

	g := TT.OpenAnalytics(dbname,url,user,pwd)

	// Trusting DNS

	tcpServer, err := net.ResolveTCPAddr(TYPE, HOST+":"+PORT)

	if err != nil {
		println("ResolveTCPAddr failed:", err.Error())
		os.Exit(1)
	}

	conn, err := net.DialTCP(TYPE, nil, tcpServer)

	if err != nil {
		println("Dial failed:", err.Error())
		os.Exit(1)
	}

	ctx:= TT.PromiseContext_Begin(g,"tcp_service") // periodigram?

	fmt.Println("1. S delivers request onto R, conditionally on prearranged promise protocol bundle")

	fmt.Println("=========IDENTITY=================")
	localAddr := conn.LocalAddr().(*net.TCPAddr)
	remoteAddr := conn.RemoteAddr().(*net.TCPAddr)
	fmt.Println("Local IP",localAddr)
	fmt.Println("Remote IP",remoteAddr)
	fmt.Println("=========IDENTITY=================")

	_, err = conn.Write([]byte(sendbuf))

	if err != nil {
		println("Write data failed:", err.Error())
		os.Exit(1)
	}

	fmt.Println("2. S promises to accept reply from R as part of prearranged protocol bundle")

	received := make([]byte, 2048)

	_, err = conn.Read(received)

	if err != nil {
		println("Read data failed:", err.Error())
		os.Exit(1)
	}

	e := TT.PromiseContext_End(g,ctx)

	// Do we know what was promised? Or how to express it? Our SLO?

	promised_upper_bound := 1.6 // response time in seconds
	trust_interval := 1.0       // monitor interval in seconds

	V := TT.AssessPromiseOutcome(g,e,AssessResult(string(received)),promised_upper_bound, trust_interval)

	s := fmt.Sprintf("/tmp/server_%v",remoteAddr)
	TT.AppendFileValue(s,V)

	conn.Close()
}

// ***************************************************************

// Each agent needs to provide a function to return a value in a
// fixed set TT.const -- here we're assessing the server's promise

func AssessResult(res string) float64 {

	// we need to place a value on what was delivered b+ intersect our b-

	// This includes amount and quality. It assumes we know what was promised,
	// because monitoring systems don't generally know that, and tend to look
	// at implicit measures that are only peripheral to the promised outcome

	// One reason to trust is that we don't really know what we want, or
	// what is being offered - but this is often used as an excuse not to trust

	//fmt.Printf("SELF-ASSESSING RETURN(%s)\n",res)

	if strings.Contains(res,"server_excellent") {
		return TT.ASSESS_EXCELLENT
	}

	if strings.Contains(res,"server_ok") {
		return TT.ASSESS_PAR
	}

	if strings.Contains(res,"server_weak") {
		return TT.ASSESS_WEAK
	}

	if strings.Contains(res,"Too soon") {

		fmt.Println("\n !! Connection rejected by rate limiting")
		os.Exit(0)
	}

	return TT.ASSESS_SUBPAR
}