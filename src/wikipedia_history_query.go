
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

	curr := GetEpisodeHead(subject)

	fmt.Println("Starting",curr)

	for next := GetNextEpisode(curr); next != "none"; next = GetNextEpisode(next) {

		fmt.Println(next)
		list = append(list,next)
		users := GetEpisodeUsers(next)
		fmt.Println(users)
	}

	return list
}

// ********************************************************************************

func GetEpisodeHead(subject string) string {

	var err error
	var cursor A.Cursor

	instring := "FOR n in follows FILTER n._to == 'topic/"+ subject +"' RETURN n._from"

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

	instring := "FOR n in follows FILTER n._from == '"+ current +"' && n.semantics == 'THEN' RETURN n._to"

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

	instring := "FOR n in follows FILTER n._to == '"+ current +"' && n.semantics == 'INFL'  RETURN n._from"

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
