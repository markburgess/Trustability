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
//     go run udp_server.go
// then 
//     go run udp_client.go
//
// ****************************************************************************

package main

import (
	"net"
	"os"
	"TT"
	"time"
	"fmt"
)

// *******************************************************************

func main() {

	//

	var dbname string = "SemanticSpacetime"
	var url string = "http://localhost:8529"
	var user string = "root"
	var pwd string = "mark"

	g := TT.OpenAnalytics(dbname,url,user,pwd)

	//

	fmt.Println("UDP: We have not established a promise protocol with extended conditional dependences, i.e. this is a connectionless shot in the dark")

	udpServer, err := net.ResolveUDPAddr("udp", ":1053")

	if err != nil {
		println("ResolveUDPAddr failed:", err.Error())
		os.Exit(1)
	}

	ctx:= TT.PromiseContext_Begin(g,"udp_service") // periodigram?

	endpoint, err := net.DialUDP("udp", nil, udpServer)

	if err != nil {
		println("Failed to establish endpoint:", err.Error())
		os.Exit(1)
	}

	defer endpoint.Close()

	_, err = endpoint.Write([]byte("This is a UDP process message from S"))

	if err != nil {
		println("Write failed:", err.Error())
		os.Exit(1)
	}

	fmt.Println("=========IDENTITY=================")
	localAddr := endpoint.LocalAddr().(*net.UDPAddr)
	remoteAddr := endpoint.RemoteAddr().(*net.UDPAddr)
	fmt.Println("Local IP",localAddr)
	fmt.Println("Remote IP",remoteAddr)
	fmt.Println("=========IDENTITY=================")

	received := make([]byte, 1024)

	// busy waiting = mistrusting the absence of clients
	// Use RandomAccept rate as a reliability / attention rate

	if TT.RandomAccept(0.8) {

		println("Waiting for a response...")

		// Set the read deadline to 10 seconds
		errtimer := endpoint.SetReadDeadline(time.Now().Add(10 * time.Second))

		if errtimer != nil {
			println("Unable to set timeout")
		}

		_, err = endpoint.Read(received)
		
		AssessPromiseOutcome() // this has to be specific to each agent and process

		if err != nil {

			println("Server left me hanging, read failed from R:", err.Error())

		} else {

			println("R replied with:", string(received))
		}
	}

	TT.PromiseContext_End(g,ctx)

	println("Learn/update trustworthiness of R by S from return value e")
	println("Update policy trust/attentiveness to R by S")
}

// *******************************************************************

func AssessPromiseOutcome() {
/*
Extreme events
	EstimateRiskForPromise() - importance MEGA, MEDIUM, LOW

Testy events
        Confidence()

Regular events
	UpdateTrustWorthInAgentPromise() - how much up or down?
	UpdateTrustWorthInAgent()
	UpdateTrustForAgentPromise()
*/
}