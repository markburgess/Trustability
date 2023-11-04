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

	users := GetEpisodeChain(args[0])
	
	baddies := GetEpisodeUsersBySignal("contentious")

	fmt.Println("\nContentious users:",len(baddies),"of",len(users),"=",100*float64(len(baddies))/float64(len(users)),"% \n   ")//,baddies)

	summ := TT.GetEpisodeData(G,args[0])

	fmt.Println(" Length of article",summ.L)
	fmt.Println(" Average <N>",summ.N)
	fmt.Println(" Incidents % length",summ.I*100)
	fmt.Println(" discussion/L %",summ.W*100)
	fmt.Println(" mistrust policy s/H %",summ.M*100)
	fmt.Println(" av duration per episode (days)",summ.TG)
	fmt.Println(" bot fraction %",summ.BF*100)

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

	var total_duration float64 = 0
	var total_episodes float64 = 0

	for next := GetEpisodeHead(subject); next != "none"; next = GetNextEpisode(next) {

		this_ep := make(map[string]int)
		list = append(list,next)
		ep_users := GetEpisodeUsers(next)

		node := TT.GetFullNode(G,next)
		start_time := time.Unix(0,node.Begin)
		end_time := time.Unix(0,node.End)
		duration := float64(node.End - node.Begin) / float64(TT.NANO*24*3600)
		total_duration += duration
		total_episodes++

		interval := float64(node.Gap) / float64(TT.NANO*24*3600)

		if interval >= 0 {
			fmt.Println("\n gap",interval,"days\n")
		}

		fmt.Println("\nTopic:",next,"(",len(ep_users),"ep_users",")")
		fmt.Println(" occurred between",start_time.UTC(),"and",end_time.UTC())
		fmt.Println(" duration ",duration,"days")

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

	fmt.Println("Average duration per episode", total_duration/total_episodes)

	return all_users
}

// ********************************************************************************

func GetEpisodeHead(subject string) string {

	var err error
	var cursor A.Cursor

	instring := "FOR n in Follows FILTER n._from == 'topic/"+ subject +"' && n.semantics == 'THEN' RETURN n._to"

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
