
package main

import (
	"fmt"
	"flag"
	"os"
	"context"
	"time"
	"TT"
	A "github.com/arangodb/go-driver"

)

// ********************************************************************************

var G TT.Analytics

// ********************************************************************************

func main() {
	
        flag.Usage = usage
        flag.Parse()
        args := flag.Args()

	if ! (len(args) > 0) {
                usage()
                os.Exit(1);
        }

	// ***********************************************************
	
	TT.InitializeSmartSpaceTime()

	var dbname string = "SemanticSpacetime"
	var dburl string = "http://localhost:8529"
	var user string = "root"
	var pwd string = "mark"

	G = TT.OpenAnalytics(dbname,dburl,user,pwd)

	users := GetEpisodeChain(args[0])
	
	baddies := GetEpisodeUsersBySignal("contentious")

	fmt.Println("\nContentious users:",len(baddies),"of",len(users),"\n   ",baddies)
}

//**************************************************************

func usage() {

        fmt.Fprintf(os.Stderr, "usage: go run wikipedia_history_query.go [subject_topic]\n")
        flag.PrintDefaults()
        os.Exit(2)
}

// ********************************************************************************

func GetEpisodeChain(subject string) []string {

	var list []string
	var all_users []string

	// Here just looking at all the adjacency relations ADJ_* of type Near
	// could add a filter, e.g. FOR n in Near FILTER n.semantics == "ADJ_NODE"

	fmt.Println("Starting",subject)

	repeat_users := make(map[string]int)

	for next := GetEpisodeHead(subject); next != "none"; next = GetNextEpisode(next) {

		this_ep := make(map[string]int)
		list = append(list,next)
		ep_users := GetEpisodeUsers(next)

		fmt.Println("\nTopic:",next,"(",len(ep_users),"ep_users",")")

		for u := range ep_users {
			this_ep[ep_users[u]]++
			repeat_users[ep_users[u]]++
		}

		for unique := range this_ep {
			fmt.Println(" involved ",unique, "making", this_ep[unique],"contributions")
		}
	}

	fmt.Println("\nRepeat user summary:\n")

	for unique := range repeat_users {
		all_users = append(all_users,unique)
		if repeat_users[unique] > 1 {
			fmt.Println(" involved ",unique, "making", repeat_users[unique],"contributions")
		}
	}

	return all_users
}

// ********************************************************************************

func GetEpisodeHead(subject string) string {

	var err error
	var cursor A.Cursor

	f := TT.LINKTYPES[TT.GR_FOLLOWS]

	instring := "FOR n in "+f+" FILTER n._to == 'topic/"+ subject +"' RETURN n._from"

	// This might take a long time, so we need to extend the timeout

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(time.Hour*8))

	defer cancel()

	cursor,err = G.S_db.Query(ctx,instring,nil)

	if err != nil {
		fmt.Printf("Query failed: %v", err)
		os.Exit(1)
	}

	defer cursor.Close()

	for {
		var key string

		_,err = cursor.ReadDocument(nil,&key)

		if A.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			fmt.Printf("Doc returned: %v", err)
		} else {
			return key
		}
	}

return "none"
}

// ********************************************************************************

func GetNextEpisode(current string) string {

	var err error
	var cursor A.Cursor

	// Here just looking at all the adjacency relations ADJ_* of type Near
	// could add a filter, e.g. FOR n in Near FILTER n.semantics == "ADJ_NODE"

	f := TT.LINKTYPES[TT.GR_FOLLOWS]

	instring := "FOR n in "+f+" FILTER n._from == '"+ current +"' && n.semantics == 'THEN' RETURN n._to"

	// This might take a long time, so we need to extend the timeout

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(time.Hour*8))

	defer cancel()

	cursor,err = G.S_db.Query(ctx,instring,nil)

	if err != nil {
		fmt.Printf("Query failed: %v", err)
		os.Exit(1)
	}

	defer cursor.Close()

	for {
		var key string

		_,err = cursor.ReadDocument(nil,&key)

		if A.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			fmt.Printf("Doc returned: %v", err)
		} else {
			return key
		}
	}

return "none"
}

// ********************************************************************************

func GetEpisodeUsers(current string) []string {

	var err error
	var cursor A.Cursor
	var list []string

	// Here just looking at all the adjacency relations ADJ_* of type Near
	// could add a filter, e.g. FOR n in Near FILTER n.semantics == "ADJ_NODE"

	f := TT.LINKTYPES[TT.GR_FOLLOWS]

	instring := "FOR n in "+f+" FILTER n._to == '"+ current +"' && n.semantics == 'INFL'  RETURN n._from"

	// This might take a long time, so we need to extend the timeout

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(time.Hour*8))

	defer cancel()

	cursor,err = G.S_db.Query(ctx,instring,nil)

	if err != nil {
		fmt.Printf("Query failed: %v", err)
		os.Exit(1)
	}

	defer cursor.Close()

	for {
		var key string

		_,err = cursor.ReadDocument(nil,&key)

		if A.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			fmt.Printf("Doc returned: %v", err)
		} else {
			list = append(list,key)
		}
	}

return list
}

// ********************************************************************************

func GetEpisodeUsersBySignal(sig string) []string {

	var err error
	var cursor A.Cursor
	var list []string

	// Here just looking at all the adjacency relations ADJ_* of type Near
	// could add a filter, e.g. FOR n in Near FILTER n.semantics == "ADJ_NODE"

	f := TT.LINKTYPES[TT.GR_EXPRESSES]

	instring := "FOR n in "+f+" FILTER n._to == 'signal/"+ sig +"' RETURN n._from"

	// This might take a long time, so we need to extend the timeout

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(time.Hour*8))

	defer cancel()

	cursor,err = G.S_db.Query(ctx,instring,nil)

	if err != nil {
		fmt.Printf("Query failed: %v", err)
		os.Exit(1)
	}

	defer cursor.Close()

	for {
		var key string

		_,err = cursor.ReadDocument(nil,&key)

		if A.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			fmt.Printf("Doc returned: %v", err)
		} else {
			list = append(list,key)
		}
	}

return list
}
