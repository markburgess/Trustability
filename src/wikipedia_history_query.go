
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

	list := GetEpisodeChain(args[0])
	TT.Println(list)
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

	// Here just looking at all the adjacency relations ADJ_* of type Near
	// could add a filter, e.g. FOR n in Near FILTER n.semantics == "ADJ_NODE"

	fmt.Println("Starting",subject)

	repeat_users := make(map[string]int)

	for next := GetEpisodeHead(subject); next != "none"; next = GetNextEpisode(next) {

		list = append(list,next)
		users := GetEpisodeUsers(next)

		fmt.Println("\nTopic:",next,"(",len(users),"users",")")

		for u := range users {
			repeat_users[users[u]]++
		}

		for u := range users {
			fmt.Println(" involved ",users[u], "making", repeat_users[users[u]],"contributions")
		}
	}

	return list
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
