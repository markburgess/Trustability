
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

	subject := args[0]
	fmt.Println("Search string",subject)

	TT.FractionateSentences(subject) // finds STM_NGRAM_RANK[n]

	order := ScoreTopicsBasedOnFractions()

	fmt.Println("OUT",order)

	topics := GetStorylineForSubject(subject)

	fmt.Println(topics)
}

//**************************************************************

func usage() {

        fmt.Fprintf(os.Stderr, "usage: go run wikipedia_topic_query.go [subject_topic]\n")
        flag.PrintDefaults()
        os.Exit(2)
}

// ********************************************************************************

func ScoreTopicsBasedOnFractions() []string {

	var list []string

	for n := 3; n < 6; n++ {
		for ngram := range TT.STM_NGRAM_RANK[n] {
			list = append(list,GetTopicsInheriting(ngram)...)
		}
	}

	return list
}

// ********************************************************************************

func GetTopicsInheriting(frag string) []string {

	can := TT.CanonifyName(frag)

	q := "FOR n in Contains FILTER n._to == 'ngram/"+can+"' && n.semantics == 'CONTAINS' RETURN n._from"

	fmt.Println("SS",q)

	// This might take a long time, so we need to extend the timeout

	var err error
	var cursor A.Cursor
	var list []string

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(time.Hour*8))

	defer cancel()

	cursor,err = G.S_db.Query(ctx,q,nil)

	if err != nil {
		return list
	}

	defer cursor.Close()

	for {
		var node string

		_,err = cursor.ReadDocument(nil,&node)

		if A.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			fmt.Printf("Doc returned: %v", err)
		} else {
			list = append(list,node)
		}
	}

	return list

}

// ********************************************************************************

func GetStorylineForSubject(subject string) []string {

	q := "FOR n in Follows FILTER n._from == 'topic/"+ subject +"' && n.semantics == 'LEADS_TO' RETURN n._to"

	// This might take a long time, so we need to extend the timeout

	var err error
	var cursor A.Cursor
	var list []string

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(time.Hour*8))

	defer cancel()

	cursor,err = G.S_db.Query(ctx,q,nil)

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

