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
// e.g.
//      go run wikipedia_history_query.go Arnold_Bax
// ***********************************************************

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

	topics := GetStorylineForSubject(subject)

	history := GetHistoryForSubject(subject)

/*
	TT.FractionateSentences(subject) // finds STM_NGRAM_RANK[n]

	order := ScoreTopicsBasedOnFractions()

	fmt.Println("OUT",order)


*/

	fmt.Println("STORY\n")

	for i := 0; i < len(topics); i++ {
		fmt.Println(i,":",topics[i],"\n")
	}

	fmt.Println("History\n")

	for i := 0; i < len(history); i++ {
		fmt.Println(i,history[i])
	}

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

// Cast a wide net .. with all ngrams

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

	head := GetStoryHead(subject,"LEADS_TO")

	if head == "" {
		return nil
	}

	return GetStoryTail(head,200)
}

// ********************************************************************************

func GetHistoryForSubject(subject string) []string {

	head := GetStoryHead(subject,"THEN")

	if head == "" {
		return nil
	}

	return GetStoryTail(head,200)
}

// ********************************************************************************

func GetStoryHead(subject,kind string) string {

	q := "FOR n in Follows FILTER n._from == 'topic/"+ subject +"' && n.semantics == '"+kind+"' RETURN n._to"
	// This might take a long time, so we need to extend the timeout

	var err error
	var cursor A.Cursor

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
			return key
		}
	}

	return ""
}

// ********************************************************************************

func GetStoryTail(startnode string, event_horizon int) []string {

	linktype := "Follows"
	
	q := fmt.Sprintf("FOR n IN 1..%d OUTBOUND '%s' %s OPTIONS { dfs: 'true'} RETURN n", event_horizon,startnode,linktype)
	
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(time.Hour*8))
	
	defer cancel()
	
	cursor,err := G.S_db.Query(ctx,q,nil)
	
	if err != nil {
		fmt.Printf("Query failed: %v", err)
		os.Exit(1)
	}
	
	defer cursor.Close()
	
	var list []string

	list = append(list,startnode)

	for {
		var key TT.Node

		_,err = cursor.ReadDocument(nil,&key)
		
		if A.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			fmt.Printf("Doc returned: %v", err)
		} else {
			list = append(list,key.Data)
		}
	}
	
	return list
}