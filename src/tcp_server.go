
package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"time"
	"strings"
	"TT"
)

const (
	HOST = "localhost"
	PORT = "8080"
	TYPE = "tcp"
)

// ***************************************************************

func main() {

	fmt.Println("Promising unconditionally to attend to promised messages and impositions from anyone...but not necessarily to accept impositions")

	listen, err := net.Listen(TYPE, HOST+":"+PORT)

	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	//

	var dbname string = "SemanticSpacetime"
	var url string = "http://localhost:8529"
	var user string = "root"
	var pwd string = "mark"

	g := TT.OpenAnalytics(dbname,url,user,pwd)

	// 

	defer listen.Close()

	for count := 1; count < 10; count++ {
		conn, err := listen.Accept()
		if err != nil {
			log.Fatal(err)
			os.Exit(1)
		}

		go handleRequest(conn,g,count)
	}
}

// ***************************************************************

func handleRequest(conn net.Conn, g TT.Analytics, count int) {

	fmt.Println("=========IDENTITY=================")
	localAddr := conn.LocalAddr().(*net.TCPAddr)
	remoteAddr := conn.RemoteAddr().(*net.TCPAddr)
	fmt.Println("Local IP",localAddr)
	fmt.Println("Remote IP",remoteAddr)
	fmt.Println("=========IDENTITY=================")

	// On the server side, each connection is an imposition so we're naturally
	// less trusting on the server side. Server is unaware of client intentions

	ctx:= TT.PromiseContext_Begin(g,"tcp_serviceprovider") // periodigram?

	received := make([]byte, 1024)

	_, err := conn.Read(received)

	if err != nil {
		log.Fatal(err)
		conn.Close()
	}

	//fmt.Println("Delaying.....reply",count)
	time.Sleep(time.Duration(count*2)*time.Millisecond*300)

	time := time.Now().Format(time.ANSIC)

	responseStr := fmt.Sprintf("roundtrip \"%s\", received at: %v", string(received), time)

	conn.Write([]byte(responseStr))

	fmt.Println("responding with:", responseStr)

	e := TT.PromiseContext_End(g,ctx)

	// Do we know what was promised? Or how to express it?

	promised_upper_bound := 1.6 // response time in seconds
	trust_interval := 1.0       // monitor interval in seconds

	V := TT.AssessPromiseOutcome(g,e,AssessResult(string(received)),promised_upper_bound,trust_interval)

	// On the server side, the port is random so strip it off

	s := strings.Split(fmt.Sprintf("/tmp/client_%v",remoteAddr),":")
	TT.AppendFileValue(s[0],V)

	conn.Close()
}

// ***************************************************************

// Each agent needs to provide a function to return a value in a
// fixed set TT.const - here we're assessing the client's message

func AssessResult(res string) float64 {

	// we need to place a value on what was delivered b+ intersect our b-

	// This includes amount and quality. It assumes we know what was promised,
	// because monitoring systems don't generally know that, and tend to look
	// at implicit measures that are only peripheral to the promised outcome

	// One reason to trust is that we don't really know what we want, or
	// what is being offered - but this is often used as an excuse not to trust

	fmt.Printf("SELF-ASSESSING RETURN(%s)\n",res)

	// since this is an imposition, we might want to assess relevance/intent

	// e.g. is this a reasonable request, or does it look like DOS or manipulation..

	if strings.Contains(res,"client_excellent") {
		return TT.ASSESS_EXCELLENT
	}

	if strings.Contains(res,"client_ok") {
		return TT.ASSESS_PAR
	}

	if strings.Contains(res,"client_weak") {
		return TT.ASSESS_WEAK
	}

	return TT.ASSESS_SUBPAR
}