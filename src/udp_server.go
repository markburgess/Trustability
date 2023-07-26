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
	"fmt"
	"log"
	"net"
	"time"
	"TT"
)

// *******************************************************************

func main() {


	udpServer, err := net.ListenPacket("udp", ":1053")

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Promising unconditionally to attend to promised messages and impositions from anyone...")

	defer udpServer.Close()

	for {
		buf := make([]byte, 1024)

		_, addr, err := udpServer.ReadFrom(buf)

		fmt.Println("Blocking wait received imposition:",string(buf))

		if err != nil {
			continue
		}

		go response(udpServer, addr, buf)
	}

}

// ************************************************************************

func response(udpServer net.PacketConn, addr net.Addr, buf []byte) {

	time := time.Now().Format(time.ANSIC)

	responseStr := fmt.Sprintf("time received: %v. Your message was: %v!", time, string(buf))

	// Use RandomAccept rate as a reliability / attention rate

	if TT.RandomAccept(0.8) { 

		fmt.Println("Replying as promised generically to all:",string(buf))
		
		udpServer.WriteTo([]byte(responseStr), addr)
	}
}