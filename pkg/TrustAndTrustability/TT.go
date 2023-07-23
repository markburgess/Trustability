//
// Copyright © Mark Burgess
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

// ***************************************************************************
//*
//* Cellibrium v3 in golang - on Arango
//*
// ***************************************************************************

package TT

import (
	"strings"
	"context"
	"fmt"
	"regexp"
	"path"
	"os"
	"runtime"
	"hash/fnv"
	"time"
	"math"
	"math/rand"
	"sort"
	"io/ioutil"

	A "github.com/arangodb/go-driver"
	"github.com/arangodb/go-driver/http"
)

// ***************************************************************************
// Some datatypes
// ***************************************************************************

type Name string
type List []string
type Neighbours []int

// ****************************************************************************

type ConnectionSemantics struct {

	LinkType string  // Association type
	From     string  // Node key pointed to

	// Used in aggregation

	FwdSrc   string
	BwdSrc   string
}

type SemanticLinkSet map[string][]ConnectionSemantics

type Cone map[int]SemanticLinkSet

// ****************************************************************************

type Analytics struct {

// Container db

S_db   A.Database

// Graph model

S_graph A.Graph

// 3 levels of nodes and supernodes

S_frags A.Collection  // fractionated Ngrams
S_nodes A.Collection  // whole semantic events
S_hubs  A.Collection  // collective patterns

// 4 primary link types

S_Follows   A.Collection
S_Contains  A.Collection
S_Expresses A.Collection
S_Near      A.Collection

// Chain memory 
previous_event_key Node
previous_event_slice []Node
}

// ***************************************************************************
// Example and for use as histograms
// ***************************************************************************

type KeyValue struct {

	K  string  `json:"_key"`
	V  float64 `json:"value"`
}

// ***************************************************************************

type PromiseContext struct {

	Time time.Time
	Name string
}

type PromiseHistory struct {

	// Use this as an event tracker, CFEngine style

	PromiseId string     `json:"_key"`

	// Three points for derivative

	Q         float64    `json:"q"`
	Q1        float64    `json:"q1"`
	Q2        float64    `json:"q2"`

	Q_av      float64    `json:"q_av"`
	Q_var     float64    `json:"q_var"`

	T         int64      `json:"lastT"`
	T1        int64      `json:"lastT1"`
	T2        int64      `json:"lastT2"`

	Dt_av     float64    `json:"dT"`
        Dt_var    float64    `json:"dT_var"`

	V         float64    `json:"V"`
	AntiT     float64    `json:"antiT"`

	Units     string     `json:"units"`
}

// ****************************************************************************

// We want to standardize the representation of assessments to quantify as a potential

const ASSESS_EXCELLENT = 1.0
const ASSESS_PAR = 0.5
const ASSESS_WEAK = 0.25
const ASSESS_SUBPAR = 0.0

const NOT_EXIST = 0

// ****************************************************************************

type Assessment struct {

	// Use this as an outcome assessment -> [-1,+1], level of promise keeping

	Key     string     `json:"_key"`
	Id      string     `json:"agent"`
	Outcome float64    `json:"outcome"`
}

// ***************************************************************************
// ngram fractionation method
// ***************************************************************************

// Short term memory class

type Narrative struct {

	rank float64
	text string
	index int
}

type Score struct {

	Key   string
	Score float64
}

// ***************************************************************************

// Promise bindings in English. This domain knowledge saves us a lot of training analysis

var FORBIDDEN_ENDING = []string{"but", "and", "the", "or", "a", "an", "its", "it's", "their", "your", "my", "of", "as", "are", "is", "be", "with", "using", "that", "who", "to" ,"no", "because","at","but","yes","no","yeah","yay"}

var FORBIDDEN_STARTER = []string{"and","or","of","the","it","because","in","that","these","those","is","are","was","were","but","yes","no","yeah","yay"}

// ***************************************************************************

var WORDCOUNT int = 0
var LEGCOUNT int = 0
var KEPT int = 0
var SKIPPED int = 0
var ALL_SENTENCE_INDEX int = 0

var THRESH_ACCEPT float64 = 0
var TOTAL_THRESH float64 = 0

// ************** SOME INTRINSIC SPACETIME SCALES ****************************

const MAXCLUSTERS = 7

var LEG_WINDOW int = 100  // sentences per leg

var ATTENTION_LEVEL float64 = 1.0
var SENTENCE_THRESH float64 = 100 // chars

const REPEATED_HERE_AND_NOW  = 1.0 // initial primer
const INITIAL_VALUE = 0.5

const MEANING_THRESH = 20      // reduce this if too few samples
const FORGET_FRACTION = 0.001  // this amount per sentence ~ forget over 1000 words

// ****************************************************************************

var LTM_EVERY_NGRAM_OCCURRENCE [MAXCLUSTERS]map[string][]int

var EXCLUSIONS []string

var STM_NGRAM_RANK [MAXCLUSTERS]map[string]float64


// ****************************************************************************
// Knowledge graph SST structures
// ****************************************************************************

type Node struct {

	Key     string  `json:"_key"`     // mandatory field (handle) - short name
	Data    string  `json: "data"`    // Longer description or bulk string data
	Prefix  string  `json:"prefix"`   // Collection: Hub, Node, Fragment?
	Weight  float64 `json:"weight"`   // importance rank
}

// ***************************************************************************

type Link struct {
	From     string `json:"_from"`     // mandatory field
	To       string `json:"_to"`       // mandatory field
        SId      string `json:"semantics"` // Matches Association key
	Negate     bool `json:"negation"`  // is this enable or block?
	Weight  float64 `json:"weight"`
	Key      string `json:"_key"`      // mandatory field (handle)
}

// ****************************************************************************

// Use these to store invariant relationship data as look up tables
// this prevents the DB data from being larger than necessary.

type Association struct {

	Key     string    `json:"_key"`

	STtype  int       `json:"STType"`
	Fwd     string    `json:"Fwd"`
	Bwd     string    `json:"Bwd"` 
	NFwd    string    `json:"NFwd"`
	NBwd    string    `json:"NBwd"`
}

//**************************************************************

var CONST_STtype = make(map[string]int)
var ASSOCIATIONS = make(map[string]Association)
var STTYPES []KeyValue

const GR_NEAR int      = 1  // approx like
const GR_FOLLOWS int   = 2  // i.e. influenced by
const GR_CONTAINS int  = 3 
const GR_EXPRESSES int = 4  // represents, etc

const NANO = 1000000000
const MILLI = 1000000

//**************************************************************

type VectorPair struct {

	From string
	To string
}

//**************************************************************

func InitializeSmartSpaceTime() {


	for i := 1; i < MAXCLUSTERS; i++ {

		STM_NGRAM_RANK[i] = make(map[string]float64)
		LTM_EVERY_NGRAM_OCCURRENCE[i] = make(map[string][]int)
	} 

	// first element needs to be there to store the lookup key
	// second element stored as int to save space

	ASSOCIATIONS["CONTAINS"] = Association{"CONTAINS",GR_CONTAINS,"contains","belongs to or is part of","does not contain","is not part of"}

	ASSOCIATIONS["GENERALIZES"] = Association{"GENERALIZES",GR_CONTAINS,"generalizes","is a special case of","is not a generalization of","is not a special case of"}

	// reversed case of containment semantics

	ASSOCIATIONS["PART_OF"] = Association{"PART_OF",-GR_CONTAINS,"incorporates","is part of","is not part of","doesn't contribute to"}

	// *

	ASSOCIATIONS["HAS_ROLE"] = Association{"HAS_ROLE",GR_EXPRESSES,"has the role of","is a role fulfilled by","has no role","is not a role fulfilled by"}

	ASSOCIATIONS["ORIGINATES_FROM"] = Association{"ORIGINATES_FROM",GR_FOLLOWS,"originates from","is the source/origin of","does not originate from","is not the source/origin of"}

	ASSOCIATIONS["EXPRESSES"] = Association{"EXPRESSES",GR_EXPRESSES,"expresses an attribute","is an attribute of","has no attribute","is not an attribute of"}

	ASSOCIATIONS["PROMISES"] = Association{"PROMISES",GR_EXPRESSES,"promises/intends","is intended/promised by","rejects/promises to not","is rejected by"}

	ASSOCIATIONS["HAS_NAME"] = Association{"HAS_NAME",GR_EXPRESSES,"has proper name","is the proper name of","is not named","isn't the proper name of"}

	// *

	ASSOCIATIONS["FOLLOWS_FROM"] = Association{"FOLLOWS_FROM",GR_FOLLOWS,"follows on from","is followed by","does not follow","does not precede"}

	ASSOCIATIONS["USES"] = Association{"USES",GR_FOLLOWS,"uses","is used by","does not use","is not used by"}

	ASSOCIATIONS["CAUSEDBY"] = Association{"CAUSEDBY",GR_FOLLOWS,"caused by","may cause","was not caused by","probably didn't cause"}

	ASSOCIATIONS["DERIVES_FROM"] = Association{"DERIVES_FROM",GR_FOLLOWS,"derives from","leads to","does not derive from","does not leadto"}

	ASSOCIATIONS["DEPENDS"] = Association{"DEPENDS",GR_FOLLOWS,"may depend on","may determine","doesn't depend on","doesn't determine"}

	// Neg

	ASSOCIATIONS["NEXT"] = Association{"NEXT",-GR_FOLLOWS,"comes before","comes after","is not before","is not after"}

	ASSOCIATIONS["THEN"] = Association{"THEN",-GR_FOLLOWS,"then","previously","but not","didn't follow"}

	ASSOCIATIONS["LEADS_TO"] = Association{"LEADS_TO",-GR_FOLLOWS,"leads to","doesn't imply","doen't reach","doesn't precede"}

	ASSOCIATIONS["PRECEDES"] = Association{"PRECEDES",-GR_FOLLOWS,"precedes","follows","doen't precede","doesn't precede"}

	// *

	ASSOCIATIONS["RELATED"] = Association{"RELATED",GR_NEAR,"may be related to","may be related to","likely unrelated to","likely unrelated to"}

	ASSOCIATIONS["ALIAS"] = Association{"ALIAS",GR_NEAR,"also known as","also known as","not known as","not known as"}

	ASSOCIATIONS["IS_LIKE"] = Association{"IS_LIKE",GR_NEAR,"is similar to","is similar to","is unlike","is unlike"}

	ASSOCIATIONS["CONNECTED"] = Association{"CONNECTED",GR_NEAR,"is connected to","is connected to","is not connected to","is not connected to"}

	ASSOCIATIONS["COACTIV"] = Association{"COACTIV",GR_NEAR,"occurred together with","occurred together with","never appears with","never appears with"}

	// *

	//SaveAssociations("ST_Associations",PC.S_db,ASSOCIATIONS)
	//newassociations := LoadAssociations(PC.S_db,"ST_Associations")
	//fmt.Println(newassociations)
}

// ****************************************************************************
//  Graph invariants
// ****************************************************************************

func CreateLink(g Analytics, c1 Node, rel string, c2 Node, weight float64) {

	var link Link

	//fmt.Println("CreateLink: c1",c1,"rel",rel,"c2",c2)

	link.From = c1.Prefix + strings.ReplaceAll(c1.Key," ","_")
	link.To = c2.Prefix + strings.ReplaceAll(c2.Key," ","_")
	link.SId = ASSOCIATIONS[rel].Key
	link.Weight = weight
	link.Negate = false

	if link.SId != rel {
		fmt.Println("Associations not set up -- missing InitializeSmartSpacecTime?")
		os.Exit(1)
	}

	AddLink(g,link)
}

// ****************************************************************************

func BlockLink(g Analytics, c1 Node, rel string, c2 Node, weight float64) {

	var link Link

	//fmt.Println("CreateLink: c1",c1,"rel",rel,"c2",c2)

	link.From = c1.Prefix + strings.ReplaceAll(c1.Key," ","_")
	link.To = c2.Prefix + strings.ReplaceAll(c2.Key," ","_")
	link.SId = ASSOCIATIONS[rel].Key
	link.Weight = weight
	link.Negate = true

	if link.SId != rel {
		fmt.Println("Associations not set up -- missing InitializeSmartSpacecTime?")
		os.Exit(1)
	}

	AddLink(g,link)
}

// ****************************************************************************

func IncrementLink(g Analytics, c1 Node, rel string, c2 Node) {

	var link Link

	//fmt.Println("IncremenLink: c1",c1,"rel",rel,"c2",c2)

	link.From = c1.Prefix + c1.Key
	link.To = c2.Prefix + c2.Key
	link.SId = ASSOCIATIONS[rel].Key

	IncrLink(g,link)
}

// ****************************************************************************

func CreateFragment(g Analytics, short_description,vardescription string, weight float64) Node {

	var concept Node

	// if no short description, use a hash of the data

	description := InvariantDescription(vardescription)

	concept.Data = description
	concept.Key = short_description             // _id
	concept.Prefix = "Fragments/"
	concept.Weight = weight

	AddFrag(g,concept)

	return concept
}

// ****************************************************************************

func CreateNode(g Analytics, short_description,vardescription string, weight float64) Node {

	var concept Node

	// if no short description, use a hash of the data

	description := InvariantDescription(vardescription)

	concept.Data = description
	concept.Key = short_description
	concept.Prefix = "Nodes/"
	concept.Weight = weight

	AddNode(g,concept)

	return concept
}

// ****************************************************************************

func CreateHub(g Analytics, short_description,vardescription string, weight float64) Node {

	var concept Node

	// if no short description, use a hash of the data

	description := InvariantDescription(vardescription)

	concept.Data = description
	concept.Key = short_description
	concept.Prefix = "Hubs/"
	concept.Weight = weight

	AddHub(g,concept)

	return concept
}

//**************************************************************

func InvariantDescription(s string) string {

	s1 := strings.ReplaceAll(s,"  "," ")
	return strings.Trim(s1,"\n ")
}

//**************************************************************

func KeyName(s string) string {

	strings.Trim(s,"\n ")
	
	if len(s) > 40 {
		s = s[:40]
	}

	s1 := strings.ReplaceAll(s,"—","_")
	s2 := strings.ReplaceAll(s1,"-","_")
	s3 := strings.ReplaceAll(s2,"`","")
	s4 := strings.ReplaceAll(s3,"'","")
	return strings.ReplaceAll(s4," ","_")
}

// ****************************************************************************
// Event History
// ****************************************************************************

func NextDataEvent(g *Analytics, shortkey, data string) Node {

	key  := CreateNode(*g, shortkey, data, 1.0)   // selection #n
	
	if g.previous_event_key.Key != "start" {
		
		CreateLink(*g, g.previous_event_key, "THEN", key, 1.0)
	}
	
	g.previous_event_key = key

	return key 
}

// ****************************************************************************

func NextParallelEvents(g *Analytics, shortkeys []string, data []string) []Node {

	// If each timestep has a number of parallel slit possibilities, we need to link them all

	var newset []Node = make([]Node,0)
	var key Node

	for to := range shortkeys {

		key = CreateNode(*g, shortkeys[to], data[to], 1.0)   // selection #n
		
		if g.previous_event_key.Key != "start" {

			// Link all the previous keys in slice

			for from := range g.previous_event_slice {

				CreateLink(*g, g.previous_event_slice[from], "THEN", key, 1.0)
			}
		}
		
		newset = append(newset,key)
	}

	g.previous_event_slice = newset
	g.previous_event_key = key

	return newset
}

// ****************************************************************************

func PreviousEvent(g *Analytics) Node {

	return g.previous_event_key
}

// ****************************************************************************

func GetNode(g Analytics, key string) string {

	var doc Node
	var prefix string
	var rawkey string
	var coll A.Collection

	prefix = path.Dir(key)
	rawkey = path.Base(key)

	//fmt.Println("Debug GetNode(key)",key," XXXX pref",prefix,"base",rawkey)

	switch prefix {

	case "Hubs": 
		coll = g.S_hubs
		break

	case "Fragments": 
		coll = g.S_frags
		break

	default:
		coll = g.S_nodes
		break
	}

	// if we use S_nodes reference then we don't need the Nodes/ prefix

	_, err := coll.ReadDocument(nil, rawkey, &doc)

	if err != nil {
		fmt.Println("No such concept",err,rawkey)
		os.Exit(1)
	}

	return doc.Data
}

//***********************************************************************
// Arango
//***********************************************************************

func OpenDatabase(name, url, user, pwd string) A.Database {

	var db A.Database
	var db_exists bool
	var err error
	var client A.Client

	ctx := context.Background()

	conn, err := http.NewConnection(http.ConnectionConfig{
		Endpoints: []string{ url },
	})

	if err != nil {
		fmt.Printf("Failed to create HTTP connection: %v", err)
	}

	client, err = A.NewClient(A.ClientConfig{
		Connection: conn,
		Authentication: A.BasicAuthentication(user, pwd),
	})

	db_exists, err = client.DatabaseExists(ctx,name)

	if db_exists {

		db, err = client.Database(ctx,name)

	} else {
		db, err = client.CreateDatabase(ctx,name, nil)
		
		if err != nil {
			fmt.Printf("Failed to create database: %v", err)
			os.Exit(1);
		}
	}

	return db
}

// ****************************************************************************

func fnvhash(b []byte) string { // Currently trusting this to have no collisions
        hash := fnv.New64a()
        hash.Write(b)
        h := hash.Sum64()
        return fmt.Sprintf("key_%d",h)
}

// **************************************************

func AddKV(g Analytics, collname string, kv KeyValue) {

	coll, err := g.S_db.Collection(nil, collname)

	if err != nil {
		coll, err = g.S_db.CreateCollection(nil, collname, nil)
		if err != nil {
			fmt.Println("AddKV No such collection",collname,kv)
			return
		}
	}

	exists,err := coll.DocumentExists(nil, kv.K)

	if !exists {

		_, err = coll.CreateDocument(nil, kv)
		
		if err != nil {
			fmt.Printf("Failed to create non existent node in AddKV: %s %v",kv.K,err)
			os.Exit(1);
		}
	} else {

		var checkkv KeyValue
		
		_,err = coll.ReadDocument(nil,kv.K,&checkkv)

		if checkkv.V != kv.V {

			_, err := coll.UpdateDocument(nil, kv.K, kv)
			if err != nil {
				fmt.Printf("Failed to update value: %s %v",kv.K,err)
				os.Exit(1);
			}
		}
	}
}

// **************************************************

func GetKV(g Analytics, collname, key string) KeyValue {

	var kv KeyValue

	coll, err := g.S_db.Collection(nil, collname)

	if err == nil {
		coll.ReadDocument(nil,key,&kv)
	}

	return kv
}

//****************************************************
// FLOAT KV
//****************************************************

func SavePromiseHistoryKVMap(g Analytics, collname string, kv []PromiseHistory) {

	// Create collection

	var err error
	var coll_exists bool
	var coll A.Collection

	coll_exists, err = g.S_db.CollectionExists(nil, collname)

	if coll_exists {
		fmt.Println("Collection " + collname +" exists already")

		coll, err = g.S_db.Collection(nil, collname)

		if err != nil {
			fmt.Printf("Existing collection: %v", err)
			os.Exit(1)
		}

	} else {

		coll, err = g.S_db.CreateCollection(nil, collname, nil)

		if err != nil {
			fmt.Printf("Failed to create collection: %v", err)
		}
	}

	for k := range kv {

		AddPromiseHistory(g, coll, collname, kv[k])
	}
}

// **************************************************

func PrintPromiseHistoryKV(g Analytics, coll_name string) {

	var err error
	var cursor A.Cursor

	querystring := "FOR doc IN " + coll_name +" LIMIT 1000 RETURN doc"

	cursor,err = g.S_db.Query(nil,querystring,nil)

	if err != nil {
		fmt.Printf("Query \""+ querystring +"\" failed: %v", err)
		return
	}

	defer cursor.Close()

	for {
		var kv PromiseHistory

		metadata,err := cursor.ReadDocument(nil,&kv)

		if A.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			fmt.Printf("KV returned: %v", err)
		} else {
			
			fmt.Print("debug (K,V): (",kv.PromiseId,",", kv.Q,")    ....    (",metadata,")\n")
		}
	}
}

// **************************************************

func AddPromiseHistory(g Analytics, coll A.Collection, coll_name string, e PromiseHistory) {

	exists,err := coll.DocumentExists(nil, e.PromiseId)

	if err != nil {
		fmt.Printf("Failed to check existent node in AddPromiseHistory: %s %v",e.PromiseId,err)
		os.Exit(1);
	}

	if exists {

		UpdatePromiseHistory(g, coll_name, e.PromiseId, e)

	} else {
		
		_, err := coll.CreateDocument(nil,e)
		
		if err != nil {
			fmt.Printf("Failed to create non existent node in AddPromiseHistory: %s %v (exists =%t)\n",e.PromiseId,err,exists)
			os.Exit(1);
		}
	}
}

// **************************************************

func GetPromiseHistory(g Analytics, collname, key string) (bool,PromiseHistory,A.Collection) {

	var coll A.Collection

	coll_exists, err := g.S_db.CollectionExists(nil, collname)

	if coll_exists {

		coll, err = g.S_db.Collection(nil, collname)

		if err != nil {
			fmt.Printf("Existing collection: %v", err)
			os.Exit(1)
		}

	} else {

		coll, err = g.S_db.CreateCollection(nil, collname, nil)

		if err != nil {
			fmt.Printf("Failed to create collection: %v", err)
			os.Exit(1)
		}
	}

	exists,err := coll.DocumentExists(nil, key)

	if exists {

		var checkkv PromiseHistory

		_,err = coll.ReadDocument(nil,key,&checkkv)
		
		return exists, checkkv, coll

	} else {
		var dud PromiseHistory
		dud.T = NOT_EXIST
		dud.Q = NOT_EXIST
		return exists, dud, coll		
	}
}

// **************************************************

func LearnUpdateKeyValue(g Analytics, coll_name, key string, q float64, units string) PromiseHistory {

	var e PromiseHistory

	e.PromiseId = key

	// Slide derivative window

	now := time.Now().UnixNano()

	// time is weird in go. Duration is basically int64 in nanoseconds

	exists, previous,coll := GetPromiseHistory(g,coll_name,key)
	
	if !exists {

		// Initial bootstrap defaults

		e.Q_av = 0.6 * float64(q)
		e.Q_var = 0

		e.T = now
		e.Dt_av = 0
		e.Dt_var = 0

		AddPromiseHistory(g, coll, coll_name, e)

	} else {
		e.Q2 = previous.Q1
		e.Q1 = previous.Q
		e.Q = q

		e.Units = units

		e.Q_av = 0.5 * previous.Q + 0.5 * float64(q)
		dv2 := (e.Q-e.Q_av) * (e.Q-e.Q_av)
		e.Q_var = 0.5 * e.Q_var + 0.5 * dv2
		
		e.T2 = previous.T1
		e.T1 = previous.T
		e.T = now

		dt := float64(now-previous.T) // time difference now-previous

		e.Dt_av = 0.5 * previous.Dt_av + 0.5 * dt
		e.Dt_var = 0.5 * e.Q_var + 0.5 * (e.Dt_av-dt) * (e.Dt_av-dt)

		UpdatePromiseHistory(g, coll_name, key, e)
	}

return e
}

// **************************************************

func UpdatePromiseHistory(g Analytics, coll_name, key string, e PromiseHistory) {

	querystring := fmt.Sprintf("LET doc = DOCUMENT(\"%s/%s\")\nUPDATE doc WITH { q: %f, q1: %f, q2: %f , q_av: %f, q_var: %f, lastT: %d, lastT1: %d,lastT22: %d, dT: %f, dT_var: %f } IN %s", coll_name,e.PromiseId,e.Q,e.Q1,e.Q2,e.Q_av,e.Q_var,e.T,e.T1,e.T2,e.Dt_av,e.Dt_var,coll_name)

	cursor,err := g.S_db.Query(nil,querystring,nil)

	if err != nil {
		fmt.Printf("Query \""+ querystring +"\" failed: %v", err)
	} else {
		cursor.Close()
	}
}

// **************************************************

func LoadPromiseHistoryKV2Map(g Analytics, coll_name string, extkv map[string]PromiseHistory) {

	var err error
	var cursor A.Cursor

	querystring := "FOR doc IN " + coll_name +" LIMIT 1000 RETURN doc"

	cursor,err = g.S_db.Query(nil,querystring,nil)

	if err != nil {
		fmt.Printf("Query failed: %v", err)
	}

	defer cursor.Close()

	for {
		var kv PromiseHistory

		_,err = cursor.ReadDocument(nil,&kv)

		if A.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			fmt.Printf("KV returned: %v", err)
		} else {
			extkv[kv.PromiseId] = kv
		}
	}
}

// **********************************************************************
// Promise Context
// **********************************************************************

func PromiseContext_Begin(g Analytics, name string) PromiseContext {

	// Set up memory for history, register callbacks

	var ctx PromiseContext
	ctx.Time = time.Now()
	ctx.Name = name
	return ctx
}

// **********************************************************************

func PromiseContext_End(g Analytics, ctx PromiseContext) PromiseHistory {

	promiseID := ctx.Name
	before := ctx.Time
	after := time.Now()

	const collname = "conn"

	// Semantic donut time key ..

	_, timeslot := DoughNowt()

	key := ctx.Name+":"+timeslot

	// make b = promise execution interval (latency) in this case

	b := float64(after.Sub(before)) // time difference now-previous

	// Direct db writes, these are separated from the time-based averaging

	previous_value := GetKV(g,collname,promiseID+"latency")
	previous_time := GetKV(g,collname,promiseID+"lasteen")

	var dt,db float64

	if previous_time.V == 0 {
		dt = 300 * NANO  // default bootstrap
	} else {
		dt = float64(after.UnixNano()) - previous_time.V
	}

	if previous_value.V == 0 {
		db = b/2         // default bootstrap
	} else {
		db = b - previous_value.V
	}

	dtau := dt/db * b

	e := LearnUpdateKeyValue(g,"observables",key,b,"ns")

	var lastlatency,lasttime KeyValue

	// Make the values latency

	lastlatency.K = promiseID+"latency"
	lastlatency.V = b

	lasttime.K = promiseID+"lastseen"
	lasttime.V = float64(after.UnixNano())

	AddKV(g,collname,lastlatency)
	AddKV(g,collname,lasttime)

	AddKV(g,promiseID+collname,lasttime)

	fmt.Println("------- INSTRUMENTATION --------------")
	fmt.Println("   Location:", promiseID+collname)
	fmt.Println("   Promise duration b (ms)", e.Q/MILLI,"=",b/MILLI)
	fmt.Println("   Running average 50/50", e.Q_av/NANO)

	fmt.Println("   Change in promise since last sample",db)
	fmt.Println("   Promise derivative b/s", db/dt)
	fmt.Println("")
	fmt.Println("   Time since last sample (s) phase",dt/NANO)
	fmt.Println("   Time signal uncertainty dtau (s) group",dtau/NANO)
	fmt.Println("   Running average sampling interval",e.Dt_av/NANO)
	fmt.Println("------- INSTRUMENTATION --------------")
	return e
}

// **********************************************************************

func AssessPromiseOutcome(g Analytics, e PromiseHistory, assessed_quality,promise_upper_bound,trust_interval float64) float64 {

	promised_ns := promise_upper_bound * NANO
	trust_ns := trust_interval * NANO

	key := e.PromiseId

	// This function decides the kinetic trust and adjusts the potential
	// V based on real time promise keeping. It doesn't consider the initial
	// determination of V -- i.e. whether we want to talk to the other agent
	// at all (as in security)

	var sig float64 = math.Sqrt(e.Q_var)

	// Here we've measured the timing and we've looked at the content
	// Now we need to determine the promise-kept assessment degree
	// and adjust the long term history for this promise

	// The trouble is that we don't usually know what was promised...

	promise_level := 1/(1+math.Exp(3*(e.Q-promised_ns)/promised_ns))

	fmt.Println("Promise level",promise_level,"+-",sig/promised_ns,"raw",e.Q/NANO,promise_upper_bound)

	if e.Dt_av == 0 {
		e.Dt_av = 1.0
	}

	fmt.Println("Assessing expected sampling interval",float64(e.T)/e.Dt_av)
	fmt.Println("Assessing desired sampling interval",float64(e.T)/trust_ns)

	// The assessed payload is the user defined arbitrary up or downvote
	// How well did we keep our promise payload?

	fmt.Println("Assessing expected Q level",float64(e.Q)/e.Q_av)
	fmt.Println("Assessing desired Q level",float64(e.Q)/promised_ns)
	fmt.Println("Assessing payload",assessed_quality)

	fmt.Println("Assessing level change",(e.Q-e.Q1)/promised_ns)

	// Get our previous estimate of reliability

	reliability := GetKV(g,"PromiseKeeping",key)

	if reliability.V == 0 {

		reliability.V = 0.5 // Start evens
	}

	// Q is always positive (latency here...)
 	// Some assessments of the event's general timeliness
	// A significant timescale for latency is 0.1 second?

	delta := promise_level * assessed_quality

	if math.Abs(e.Q_av) < sig {  // Down vote for noisy behaviour

		fmt.Println("1.PENALTY!")
		delta = delta / 1.5
	}

	// derivatives are possible signs of stress / coping (confidence)
	// if first first second derivatives are growing, this is not good for latency

	dqdt := FirstDerivative(e,promised_ns,trust_ns)
	d2qdt2 := SecondDerivative(e,promised_ns,trust_ns)

	const sensitivity = 0.01 // should this be the same for 1st and second?

	if dqdt < -sensitivity {
		fmt.Println("Gradient reducing (spot measure)")
		delta = delta + 0.1
	} else if dqdt > sensitivity {
		fmt.Println("Gradient increasing (spot measure)")
		delta = delta - 0.1
		fmt.Println("2.PENALTY!")
	}

	if d2qdt2 < -sensitivity {
		fmt.Println("Curvature decelerating (positive force)")
		delta = delta + 0.1
	} else if d2qdt2 > sensitivity {
		fmt.Println("Curvature accelerating (negative force)")
		delta = delta - 0.1
		fmt.Println("3.PENALTY!")
	}

	//if math.Fabs(SecondDeriv(e)) > SCALE {
	//	delta = delta - 0.5
	//}

	// Adjust reliability according to timing AND quality

	fmt.Println("Old ML running reliability(delta)",reliability.V)

	if delta < 0 {

		delta = 0
	}


	reliability.K = key
	reliability.V = reliability.V * 0.4 + delta * 0.6

	fmt.Println("New ML running reliability(delta)",reliability.V,delta)

	AddKV(g,"PromiseKeeping",reliability)

	return reliability.V
}

// **********************************************************************
// VARIOUS
// **********************************************************************

func RandomAccept(probability float64) bool {

	if rand.Float64() < probability {
		return true
	}

	return false	
}

//***********************************************************************
// Graph functions
//***********************************************************************

func OpenAnalytics(dbname, service_url, user, pwd string) Analytics {

	var g Analytics
	var db A.Database

	InitializeSmartSpaceTime()

	db = OpenDatabase(dbname, service_url, user, pwd)

	// Book-keeping: wiring up edgeCollection to store the edges

	var F_edges A.EdgeDefinition
	var C_edges A.EdgeDefinition
	var N_edges A.EdgeDefinition
	var E_edges A.EdgeDefinition

	F_edges.Collection = "Follows"
	F_edges.From = []string{"Nodes","Hubs","Fragments"}  // source set
	F_edges.To = []string{"Nodes","Hubs","Fragments"}    // sink set

	C_edges.Collection = "Contains"
	C_edges.From = []string{"Nodes","Hubs"}              // source set
	C_edges.To = []string{"Nodes","Hubs","Fragments"}    // sink set

	N_edges.Collection = "Near"
	N_edges.From = []string{"Nodes","Hubs","Fragments"}  // source set
	N_edges.To = []string{"Nodes","Hubs","Fragments"}    // sink set

	E_edges.Collection = "Expresses"
	E_edges.From = []string{"Nodes","Hubs"}  // source set
	E_edges.To = []string{"Nodes","Hubs"}    // sink set

	var options A.CreateGraphOptions
	options.OrphanVertexCollections = []string{"Disconnected"}
	options.EdgeDefinitions = []A.EdgeDefinition{F_edges,C_edges,N_edges,E_edges}

	// Begin - feed options into a graph 

	var graph A.Graph
	var err error
	var gname string = "concept_spacetime"
	var g_exists bool

	g_exists, err = db.GraphExists(nil, gname)

	if g_exists {
		graph, err = db.Graph(nil,gname)

		if err != nil {
			fmt.Printf("Open graph: %v", err)
			os.Exit(1)
		}

	} else {
		graph, err = db.CreateGraph(nil, gname, &options)

		if err != nil {
			fmt.Printf("Create graph: %v", err)
			os.Exit(1)
		}
	}

	// *** Nodes

	var frag_vertices A.Collection
	var node_vertices A.Collection
	var hub_vertices A.Collection

	frag_vertices, err = graph.VertexCollection(nil, "Fragments")

	if err != nil {
		fmt.Printf("Vertex collection Fragments: %v", err)
	}

	node_vertices, err = graph.VertexCollection(nil, "Nodes")

	if err != nil {
		fmt.Printf("Vertex collection Nodes: %v", err)
	}

	hub_vertices, err = graph.VertexCollection(nil, "Hubs")

	if err != nil {
		fmt.Printf("Vertex collection Hubs: %v", err)
	}

	// *** Links

	var F_edgeset A.Collection
	var C_edgeset A.Collection
	var E_edgeset A.Collection
	var N_edgeset A.Collection

	F_edgeset, _, err = graph.EdgeCollection(nil, "Follows")

	if err != nil {
		fmt.Printf("Egdes follows: %v", err)
	}

	C_edgeset, _, err = graph.EdgeCollection(nil, "Contains")

	if err != nil {
		fmt.Printf("Edges contains: %v", err)
	}

	E_edgeset, _, err = graph.EdgeCollection(nil, "Expresses")

	if err != nil {
		fmt.Printf("Edges expresses: %v", err)
	}

	N_edgeset, _, err = graph.EdgeCollection(nil, "Near")

	if err != nil {
		fmt.Printf("Edges near: %v", err)
	}

	g.S_db = db	
	g.S_graph = graph
	g.S_nodes = node_vertices
	g.S_hubs = hub_vertices
	g.S_frags = frag_vertices

	g.S_Follows = F_edgeset
	g.S_Contains = C_edgeset
	g.S_Expresses = E_edgeset	
	g.S_Near = N_edgeset

	g.previous_event_key = Node{ Key: "start" }

	return g
}

// **************************************************

func AddLinkCollection(g Analytics, name string, nodecoll string) A.Collection {

	var edgeset A.Collection
	var c A.VertexConstraints

	// Remember we have to define allowed source/sink constraints for edges

	c.From = []string{nodecoll}  // source set
	c.To = []string{nodecoll}    // sink set

	exists, err := g.S_graph.EdgeCollectionExists(nil, name)

	if !exists {
		edgeset, err = g.S_graph.CreateEdgeCollection(nil, name, c)
		
		if err != nil {
			fmt.Printf("Edge collection failed: %v\n", err)
		}
	}

return edgeset
}

// **************************************************

func AddNodeCollection(g Analytics, name string) A.Collection {

	var nodeset A.Collection

	exists, err := g.S_graph.VertexCollectionExists(nil, name)

	if !exists {
		nodeset, err = g.S_graph.CreateVertexCollection(nil, name)
		
		if err != nil {
			fmt.Printf("Node collection failed: %v\n", err)
		}
	}

return nodeset
}

// **************************************************

func AddNode(g Analytics, node Node) {

	var coll A.Collection = g.S_nodes
	InsertNodeIntoCollection(g,node,coll)
}

// **************************************************

func AddFrag(g Analytics, node Node) {

	var coll A.Collection = g.S_frags
	InsertNodeIntoCollection(g,node,coll)
}

// **************************************************

func AddHub(g Analytics, node Node) {

	var coll A.Collection = g.S_hubs
	InsertNodeIntoCollection(g,node,coll)
}

// **************************************************

func InsertNodeIntoCollection(g Analytics, node Node, coll A.Collection) {
	
	exists,err := coll.DocumentExists(nil, node.Key)
	
	if !exists {
		_, err = coll.CreateDocument(nil, node)
		
		if err != nil {
			fmt.Println("Failed to create non existent node in InsertNodeIntoCollection: ",node,err)
			return
		}

	} else {

		// Don't need to check correct value, as each tuplet is unique, but check the data

		if node.Data == "" && node.Weight == 0 {
			// Leave the values alone if we don't mean to update them
			return
		}
		
		var checknode Node

		_,err := coll.ReadDocument(nil,node.Key,&checknode)

		if err != nil {
			fmt.Printf("Failed to read value: %s %v",node.Key,err)
			return	
		}

		if checknode != node {

			//fmt.Println("Correcting link values",checknode,"to",node)

			_, err := coll.UpdateDocument(nil, node.Key, node)

			if err != nil {
				fmt.Printf("Failed to update value: %s %v",node,err)
				return

			}
		}
	}
}

// **************************************************

func AddLink(g Analytics, link Link) {

	// Don't add multiple edges that are identical! But allow types
	// We have to make our own key to prevent multiple additions
        // - careful of possible collisions, but this should be overkill

        description := link.From + link.SId + link.To
	key := fnvhash([]byte(description))

	ass := ASSOCIATIONS[link.SId].Key

	if ass == "" {
		fmt.Println("Unknown association from link",link,"Sid",link.SId)
		os.Exit(1)
	}

	edge := Link{
 	 	From: link.From, 
		SId: ass,
		Negate: link.Negate,
		To: link.To, 
		Key: key,
		Weight: link.Weight,
	}

	var links A.Collection
	var coltype int

	// clumsy abs()

	if ASSOCIATIONS[link.SId].STtype < 0 {

		coltype = -ASSOCIATIONS[link.SId].STtype

	} else {

		coltype = ASSOCIATIONS[link.SId].STtype

	}

	switch coltype {

	case GR_FOLLOWS:   links = g.S_Follows
	case GR_CONTAINS:  links = g.S_Contains
	case GR_EXPRESSES: links = g.S_Expresses
	case GR_NEAR:      links = g.S_Near

	}

	exists,_ := links.DocumentExists(nil, key)

	if !exists {
		_, err := links.CreateDocument(nil, edge)
		
		if err != nil {
			fmt.Println("Failed to add new link", err, link, edge)
			os.Exit(1);
		}
	} else {

		if edge.Weight < 0 {

			// Don't update if the weight is negative
			return
		}

		// Don't need to check correct value, as each tuplet is unique, but check the weight
		
		var checkedge Link

		_,err := links.ReadDocument(nil,key,&checkedge)

		if err != nil {
			fmt.Printf("Failed to read value: %s %v",key,err)
			os.Exit(1);	
		}

		if checkedge != edge {

			fmt.Println("Correcting link weight",checkedge,"to",edge)

			_, err := links.UpdateDocument(nil, key, edge)

			if err != nil {
				fmt.Printf("Failed to update value: %s %v",edge,err)
				os.Exit(1);

			}
		}
	}
}

// **************************************************

func IncrLink(g Analytics, link Link) {

	// Don't add multiple edges that are identical! But allow types
	// We have to make our own key to prevent multiple additions
        // - careful of possible collisions, but this should be overkill

        description := link.From + link.SId + link.To
	key := fnvhash([]byte(description))

	ass := ASSOCIATIONS[link.SId].Key

	if ass == "" {
		fmt.Println("Unknown association from link",link,"Sid",link.SId)
		os.Exit(1)
	}

	edge := Link{
 	 	From: link.From, 
		SId: ass,
		To: link.To, 
		Key: key,
		Weight: 0,
	}

	var links A.Collection
	var coltype int

	// clumsy abs()

	if ASSOCIATIONS[link.SId].STtype < 0 {

		coltype = -ASSOCIATIONS[link.SId].STtype

	} else {

		coltype = ASSOCIATIONS[link.SId].STtype

	}

	switch coltype {

	case GR_FOLLOWS:   links = g.S_Follows
	case GR_CONTAINS:  links = g.S_Contains
	case GR_EXPRESSES: links = g.S_Expresses
	case GR_NEAR:      links = g.S_Near

	}

	exists,_ := links.DocumentExists(nil, key)

	if !exists {
		_, err := links.CreateDocument(nil, edge)
		
		if err != nil {
			fmt.Println("Failed to add new link", err, link, edge)
			os.Exit(1);
		}
	} else {

		// Don't need to check correct value, as each tuplet is unique, but check the weight
		
		var checkedge Link

		_,err := links.ReadDocument(nil,key,&checkedge)

		if err != nil {
			fmt.Printf("Failed to read value: %s %v",key,err)
			os.Exit(1);	
		}

		edge.Weight = checkedge.Weight + 1.0
		
		_, err = links.UpdateDocument(nil, key, edge)
		
		if err != nil {
			fmt.Printf("Failed to update value: %s %v",edge,err)
			os.Exit(1);
			
		}
	}
}

// **************************************************

func PrintNodes(g Analytics, collection string) {

	var err error
	var cursor A.Cursor

	querystring := "FOR doc IN " + collection + " RETURN doc"

	cursor,err = g.S_db.Query(nil,querystring,nil)

	if err != nil {
		fmt.Printf("Query failed: %v", err)
	}

	defer cursor.Close()

	for {
		var doc Node

		_,err = cursor.ReadDocument(nil,&doc)

		if A.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			fmt.Printf("Doc returned: %v", err)
		} else {
			fmt.Print(collection,doc,"\n")
		}
	}
}

// **************************************************

func GetSuccessorsOf(g Analytics, node string, sttype int) SemanticLinkSet {

	return GetNeighboursOf(g,node,sttype,"+")
}

// **************************************************

func GetPredecessorsOf(g Analytics, node string, sttype int) SemanticLinkSet {

	return GetNeighboursOf(g,node,sttype,"-")
}

// **************************************************

func GetNeighboursOf(g Analytics, node string, sttype int, direction string) SemanticLinkSet {

	var err error
	var cursor A.Cursor
	var coll string

	if !strings.Contains(node,"/") {
		fmt.Println("GetNeighboursOf(node) without collection prefix",node)
		os.Exit(1)
	}

	switch sttype {

	case -GR_FOLLOWS, GR_FOLLOWS:   
		coll = "Follows"

	case -GR_CONTAINS, GR_CONTAINS:  
		coll = "Contains"

	case -GR_EXPRESSES, GR_EXPRESSES: 
		coll = "Expresses"

	case -GR_NEAR, GR_NEAR:      
		coll = "Near"
	default:
		fmt.Println("Unknown STtype in GetNeighboursOf",sttype)
		os.Exit(1)
	}

	var querystring string

	switch direction {

	case "+": 
		querystring = "FOR my IN " + coll + " FILTER my._from == \"" + node + "\" RETURN my"
		break
	case "-":
		querystring = "FOR my IN " + coll + " FILTER my._to == \"" + node + "\"  RETURN my"
		break
	default:
		fmt.Println("NeighbourOf direction can only be + or -")
		os.Exit(1)
	}

	//fmt.Println("query:",querystring)

	cursor,err = g.S_db.Query(nil,querystring,nil)

	if err != nil {
		fmt.Printf("Neighbour query \"%s\"failed: %v", querystring,err)
	}

	defer cursor.Close()

	var result SemanticLinkSet = make(SemanticLinkSet)

	for {
		var doc Link
		var nodekey string
		var linktype ConnectionSemantics

		_,err = cursor.ReadDocument(nil,&doc)

		if A.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			fmt.Printf("Doc returned: %v", err)
		} else {
			switch direction {

			case "-": 
				nodekey = doc.From
				linktype.From = doc.To
				linktype.LinkType = ASSOCIATIONS[doc.SId].Bwd
				break
			case "+":
				nodekey = doc.To
				linktype.From = doc.From
				linktype.LinkType = ASSOCIATIONS[doc.SId].Fwd
				break
			}

			result[nodekey] = append(result[nodekey],linktype)
		}
	}

	return result
}

// ********************************************************************

func GetAdjacencyMatrixByKey(g Analytics, assoc_type string, symmetrize bool) map[VectorPair]float64 {

	var adjacency_matrix = make(map[VectorPair]float64)

	var err error
	var cursor A.Cursor
	var coll string

	sttype := ASSOCIATIONS[assoc_type].STtype

	switch sttype {

	case -GR_FOLLOWS, GR_FOLLOWS:   
		coll = "Follows"

	case -GR_CONTAINS, GR_CONTAINS:  
		coll = "Contains"

	case -GR_EXPRESSES, GR_EXPRESSES: 
		coll = "Expresses"

	case -GR_NEAR, GR_NEAR:      
		coll = "Near"

	default:
		fmt.Println("Unknown STtype in GetNeighboursOf",assoc_type)
		os.Exit(1)
	}

	var querystring string

	querystring = "FOR my IN " + coll + " FILTER my.semantics == \"" + assoc_type + "\" RETURN my"

	cursor,err = g.S_db.Query(nil,querystring,nil)

	if err != nil {
		fmt.Printf("Neighbour query \"%s\"failed: %v", querystring,err)
	}

	defer cursor.Close()

	for {
		var doc Link

		_,err = cursor.ReadDocument(nil,&doc)

		if A.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			fmt.Printf("Doc returned: %v", err)
		} else {
			if sttype == GR_NEAR || symmetrize {
				adjacency_matrix[VectorPair{From: doc.From, To: doc.To }] = 1.0
				adjacency_matrix[VectorPair{From: doc.To, To: doc.From }] = 1.0
			} else {
				adjacency_matrix[VectorPair{From: doc.From, To: doc.To }] = 1.0
			}
		}
	}

return adjacency_matrix
}

// ********************************************************************

func GetAdjacencyMatrixByInt(g Analytics, assoc_type string, symmetrize bool) ([][]float64,int,map[int]string) {

	var key_matrix = make(map[VectorPair]float64)

	var err error
	var cursor A.Cursor
	var coll string

	sttype := ASSOCIATIONS[assoc_type].STtype

	switch sttype {

	case -GR_FOLLOWS, GR_FOLLOWS:   
		coll = "Follows"

	case -GR_CONTAINS, GR_CONTAINS:  
		coll = "Contains"

	case -GR_EXPRESSES, GR_EXPRESSES: 
		coll = "Expresses"

	case -GR_NEAR, GR_NEAR:      
		coll = "Near"

	default:
		fmt.Println("Unknown STtype in GetNeighboursOf",assoc_type)
		os.Exit(1)
	}

	var querystring string

	querystring = "FOR my IN " + coll + " FILTER my.semantics == \"" + assoc_type + "\" RETURN my"

	cursor,err = g.S_db.Query(nil,querystring,nil)

	if err != nil {
		fmt.Printf("Neighbour query \"%s\"failed: %v", querystring,err)
	}

	defer cursor.Close()

	var sets = make(Set)

	for {
		var doc Link

		_,err = cursor.ReadDocument(nil,&doc)

		if A.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			fmt.Printf("Doc returned: %v", err)
		} else {

			// Merge an idempotent list of nodes to find int address

			TogetherWith(sets,"adj",doc.To)
			TogetherWith(sets,"adj",doc.From)

			if sttype == GR_NEAR || symmetrize {
				key_matrix[VectorPair{From: doc.From, To: doc.To }] = 1.0
				key_matrix[VectorPair{From: doc.To, To: doc.From }] = 1.0
			} else {
				key_matrix[VectorPair{From: doc.From, To: doc.To }] = 1.0
			}
		}
	}

	//fmt.Println(sets)

	dimension := len(sets["adj"])
	var adjacency_matrix = make([][]float64,dimension)
	var keys = make(map[int]string)
	var i int = 0
	var j int = 0

	for ri := range sets["adj"] {

		adjacency_matrix[i] = make([]float64,dimension)
		keys[i] = sets["adj"][ri]

		for rj := range sets["adj"] {

			if key_matrix[VectorPair{From: sets["adj"][ri], To: sets["adj"][rj]}] > 0 {
				adjacency_matrix[i][j] = 1.0
			}
			j++
		}
		i++
	}

	return adjacency_matrix, dimension, keys
}

//*************************************************************

func GetFullAdjacencyMatrix(g Analytics, symmetrize bool) ([][]float64,int,map[int]string) {

	var key_matrix = make(map[VectorPair]float64)
	var sets = make(Set)

	var err error
	var cursor A.Cursor

	var STtypes []string = []string{ "Follows", "Contains", "Expresses", "Near" }

	for coll := range STtypes {

		var querystring string

		querystring = "FOR my IN " + STtypes[coll] + " RETURN my"
		
		cursor,err = g.S_db.Query(nil,querystring,nil)
		
		if err != nil {
			fmt.Printf("Full adjacency query \"%s\"failed: %v", querystring,err)
		}
		
		defer cursor.Close()
		
		for {
			var doc Link
			
			_,err = cursor.ReadDocument(nil,&doc)
			
			if A.IsNoMoreDocuments(err) {
				break
			} else if err != nil {
				fmt.Printf("Doc returned: %v", err)
			} else {

				// Merge an idempotent list of nodes to find int address
				
				TogetherWith(sets,"adj",doc.To)
				TogetherWith(sets,"adj",doc.From)
				
				if symmetrize {
					key_matrix[VectorPair{From: doc.From, To: doc.To }] = 1.0
					key_matrix[VectorPair{From: doc.To, To: doc.From }] = 1.0
				} else {
					key_matrix[VectorPair{From: doc.From, To: doc.To }] = 1.0
				}
			}
		}
	}

	//fmt.Println(sets)

	dimension := len(sets["adj"])
	var adjacency_matrix = make([][]float64,dimension)
	var keys = make(map[int]string)
	var i int = 0
	var j int = 0

	for ri := range sets["adj"] {

		adjacency_matrix[i] = make([]float64,dimension)
		keys[i] = sets["adj"][ri]

		for rj := range sets["adj"] {

			if key_matrix[VectorPair{From: sets["adj"][ri], To: sets["adj"][rj]}] > 0 {
				adjacency_matrix[i][j] = 1.0
			}
			j++
		}
		i++
	}

	return adjacency_matrix, dimension, keys
}

//**************************************************************

func PrintMatrix(adjacency_matrix [][]float64,dim int,keys map[int]string) {

	for i := 1; i < dim; i++ {

		fmt.Printf("%12.12s: ",keys[i])

		for j := 1; j < dim; j++ {
			fmt.Printf("%3.3f ",adjacency_matrix[i][j])
		}
		fmt.Println("")
	}
}

//**************************************************************

func PrintVector (vec []float64,dim int,keys map[int]string) {

	for i := 1; i < dim; i++ {
		
		fmt.Printf("%12.12s: ",keys[i])
		fmt.Printf("%3.3f \n",vec[i])
	}
}

//**************************************************************

func GetPrincipalEigenvector(adjacency_matrix [][]float64, dim int) []float64 {

	var ev = make([]float64,dim)
	var sum float64 = 0

	// start with a uniform positive value

	for i := 1; i < dim; i++ {
		ev[i] = 1.0
	}

	// Three iterations is probably enough .. could improve on this

	ev = MatrixMultiplyVector(adjacency_matrix,ev,dim)
	ev = MatrixMultiplyVector(adjacency_matrix,ev,dim)
	ev = MatrixMultiplyVector(adjacency_matrix,ev,dim)

	for i := 1; i < dim; i++ {
		sum += ev[i]
	}

	// Normalize vector

	if sum == 0 {
		sum = 1.0
	}

	for i := 1; i < dim; i++ {
		ev[i] = ev[i] / sum
	}

	return ev
}

//**************************************************************

func MatrixMultiplyVector(adj [][]float64,v []float64,dim int) []float64 {

	var result = make([]float64,dim)

	// start with a uniform positive value

	for i := 1; i < dim; i++ {

		result[i] = 0

		for j := 1; j < dim; j++ {

			result[i] = result[i] + adj[i][j] * v[j]
		}
	}

return result
}

//**************************************************************

func GetPossibilityCone(g Analytics, start_key string, sttype int, visited map[string]bool) (Cone,int) {

	// A cone is a sequence of spacelike slices orthogonal to the proper time defined by sttype
	// Each slice is formed from patches that spread from nodes in the current slice
	
	// width first

	var layer int = 0
	var counter int = 0
	var total int = 0
	var cone = make(Cone)

	var start string = start_key

	cone[layer] = InitializeSemanticLinkSet(start)

	for {		
		var fanout SemanticLinkSet

		cone[layer+1] = make(SemanticLinkSet)

		for nodekey := range cone[layer] {

			if visited[nodekey] {
				continue
			} else {
				visited[nodekey] = true
			}

			fanout = GetSuccessorsOf(g, nodekey, sttype)
			
			if len(fanout) == 0 {
				return cone,total
			}

			//fmt.Println(counter, "Successor", nodekey,"result", fanout)
						
			for nextkey := range fanout {

				for wire := range fanout[nextkey] {
					
					fanout[nextkey][wire].FwdSrc = nextkey

					if !AlreadyLinkType(cone[layer+1][nextkey],fanout[nextkey][wire]) {

						cone[layer+1][nextkey] = append(cone[layer+1][nextkey],fanout[nextkey][wire])
					}
				}

				//fmt.Println("Debug",counter,nextkey,fanout[nextkey])				
				counter = counter + 1
			}
		}
		
		layer = layer + 1
		total = total + counter
		counter = 0
	}
}

// **************************************************

func AlreadyLinkType(existing []ConnectionSemantics, newlnk ConnectionSemantics) bool {

	for e := range existing {

		if newlnk.LinkType == existing[e].LinkType {
			return true
		}
	}

return false
}

// **************************************************

func GetConePaths(g Analytics, start_key string, sttype int, visited map[string]bool) []string {

	// A cone is a sequence of spacelike slices orthogonal to the proper time defined by sttype
	// Each slice is formed from patches that spread from nodes in the current slice
	
	// width first

	var layer int = 0

	paths := GetPathsFrom(g, layer, start_key, sttype, visited)
	return paths
}

// **************************************************

func GetPathsFrom(g Analytics, layer int, startkey string, sttype int, visited map[string]bool) []string {

	// return a path starting from startkey

	var paths []string

	var fanout SemanticLinkSet

	// opendir()

	fanout = GetSuccessorsOf(g, startkey, sttype)
	
	if len(fanout) == 0 {
		return nil
	}
	
	// (readdir())
	for nextkey := range fanout {

		// Get the previous mixed link state
		
		var mixed_link string = ":("
	
		// join multiple linknames pointing to nextkey

		for linktype := range fanout[nextkey] {
			
			if len(mixed_link) > 2 {
				mixed_link = mixed_link + " or "
			}
			
			mixed_link = mixed_link + fanout[nextkey][linktype].LinkType
		}
		
		mixed_link = mixed_link + "):"

		prefix:= startkey + mixed_link

		// Then look for postfix children - depth first
		// which returns a string starting from nextkey
	
		subdir := GetPathsFrom(g,layer+1,nextkey,sttype,visited)
		
		for subpath := 0; subpath < len(subdir); subpath++ {

			paths = append(paths,prefix + subdir[subpath])
		}

		if len(subdir) == 0 {
			
			paths = append(paths,prefix + nextkey + ":(end)")
		}
	}

	return paths
}

// **************************************************

func InitializeSemanticLinkSet(start string) SemanticLinkSet {
	
	var startlink SemanticLinkSet = make(SemanticLinkSet)
	startlink[start] = []ConnectionSemantics{ ConnectionSemantics{From: "nothing"}}
	return startlink
}

// **************************************************

func SaveAssociations(collname string, db A.Database, kv map[string]Association) {

	// Create collection

	var err error
	var coll_exists bool
	var coll A.Collection

	coll_exists, err = db.CollectionExists(nil, collname)

	if coll_exists {
		fmt.Println("Collection " + collname +" exists already")

		coll, err = db.Collection(nil, collname)

		if err != nil {
			fmt.Printf("Existing collection: %v", err)
			os.Exit(1)
		}

	} else {

		coll, err = db.CreateCollection(nil, collname, nil)

		if err != nil {
			fmt.Printf("Failed to create collection: %v", err)
		}
	}

	for k := range kv {

		AddAssocKV(coll, k, kv[k])
	}
}

// **************************************************

func LoadAssociations(db A.Database, coll_name string) map[string]Association {

	assocs := make(map[string]Association)

	var err error
	var cursor A.Cursor

	querystring := "FOR doc IN " + coll_name +" LIMIT 1000 RETURN doc"

	cursor,err = db.Query(nil,querystring,nil)

	if err != nil {
		fmt.Printf("Query failed: %v", err)
	}

	defer cursor.Close()

	for {
		var assoc Association

		_,err = cursor.ReadDocument(nil,&assoc)

		if A.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			fmt.Printf("Assoc returned: %v", err)
		} else {
			assocs[assoc.Key] = assoc
		}
	}

	return assocs
}

// **************************************************

func AddAssocKV(coll A.Collection, key string, assoc Association) {

	// Add data with convergent semantics, CFEngine style

	exists,err := coll.DocumentExists(nil, key)

	if !exists {

		_, err = coll.CreateDocument(nil, assoc)
		
		if err != nil {
			fmt.Printf("Failed to create non existent node: %s %v",key,err)
			os.Exit(1);
		}
	} else {

		var checkassoc Association
		
		_,err = coll.ReadDocument(nil,key,&checkassoc)

		if checkassoc != assoc {

			_, err := coll.UpdateDocument(nil, key, assoc)
			if err != nil {
				fmt.Printf("Failed to update value: %s %v",assoc,err)
				os.Exit(1);

			}
		}
	}
}

// ****************************************************************************

func FirstDerivative(e PromiseHistory, qscale,tscale float64) float64 {

	dq := (e.Q - e.Q1)/qscale
	dt := float64(e.T-e.T1)/tscale

	if dt == 0 {
		return 0
	}

	dqdt := dq/dt

	fmt.Println("Deriv dq/dt (latency)",dqdt)

	return dqdt
}

// ****************************************************************************

func SecondDerivative(e PromiseHistory, qscale,tscale float64) float64 {

	dv := ((e.Q - e.Q1)/float64(e.T-e.T1) - (e.Q1 - e.Q2)/float64(e.T1-e.T2))/qscale*tscale

	dt := (e.Q1 *float64(e.T-e.T1)/tscale)

	d2qdt2 := dv/dt

	if dt == 0 {
		return 0
	}

	fmt.Println("Deriv d2q/dt2 (latency)",d2qdt2)

	return d2qdt2
}

// ****************************************************************************
// Generate graphs
// ****************************************************************************

func AppendStringToFile(name string, s string) {

	f, err := os.OpenFile(name,os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	if err != nil {
		fmt.Println("Couldn't open for write/append to",name,err)
		return
	}

	_, err = f.WriteString(s)

	if err != nil {
		fmt.Println("Couldn't write/append to",name,err)
	}

	f.Close()
}

// ****************************************************************************

func AppendFileValue(name string, value float64) {

	s := fmt.Sprintf("%f\n",value)
	AppendStringToFile(name,s)
}

// ****************************************************************************
// Set/Collection Aggregation - two versions using hashing/lists, which faster?
// ****************************************************************************

type Set map[string]map[string]string
type LinSet map[string][]string

// ****************************************************************************

func BelongsToSet(sets Set,member string) (bool,string,string) {

	// Generate the formatted superset of all nodes that contains "member" within it
	
	for s := range sets {

		if sets[s][member] == member {
			var list string
			for l := range sets[s] {
				list = list + sets[s][l] + ","
			}
			return true,"super-"+s,list
		}
	}
	
	return false,"",""
}

// ****************************************************************************

func TogetherWith(sets Set, a1,a2 string) {

	// Place a1 and s2 into the same set, growing the sets if necessary
	// i.e. gradual accretion of sets by similarity of a1 and a2, we use
	// maps (hashes) so no linear searching as lists get big

	var s1,s2 string

	var got1 bool = false
	var got2 bool = false

	for s := range sets {

		if sets[s][a1] == a1 {
			s1 = s
			got1 = true
		}
			
		if sets[s][a2] == a2 {
			s2 = s
			got2 = true
		}

		if got1 && got2 {
			break
		}
	}

	if got1 && got2 {

		if s1 == s2 {
			
			return        // already ok
			
		} else {
			// merge two sets - this might be a mistake when data are big
			// would like to just move a tag somehow, but still the search time
			// has to grow as the clusters cover more data

			// Since this is time consuming, move the smaller set

			l1 := len(sets[s1])
			l2 := len(sets[s2])

			if (l1 <= l2) {
				for m := range sets[s1] {
					sets[s2][m] = sets[s1][m]
				}
				delete(sets,s1)
			} else {
				for m := range sets[s2] {
					sets[s1][m] = sets[s2][m]
				}
				delete(sets,s2)
			}

			return
		}
	} 

	if got1 { // s1 is the home
		sets[s1][a2] = a2
		return
	}

	if got2 { // s2 is the home
		sets[s2][a1] = a1
		return
	}

	// new pair, pick a key

	sets[a1] = make(map[string]string)
	sets[a2] = make(map[string]string)

	sets[a1][a1] = a1
	sets[a1][a2] = a2

}

// ****************************************************************************
// Linearized version
// ****************************************************************************

func LinTogetherWith(sets LinSet, a1,a2 string) {

	var s1,s2 string

	var got1 bool = false
	var got2 bool = false

	for s := range sets {

		for m:= range sets[s] {
			if sets[s][m] == a1 {
				s1 = s
				got1 = true
			}
			
			if sets[s][m] == a2 {
				s2 = s
			got2 = true
			}
		}
		
	}

	if got1 && got2 {

		if s1 == s2 {
			
			return        // already ok
			
		} else {
			// merge two sets

			l1 := len(sets[s1])
			l2 := len(sets[s2])

			if (l1 <= l2) {
				for m := range sets[s1] {
					sets[s2] = append(sets[s2],sets[s1][m])
				}
				delete(sets,s1)
			} else {
				for m := range sets[s1] {
					sets[s1] = append(sets[s1],sets[s2][m])
				}
				delete(sets,s2)
			}

			return
		}
	} 

	if got1 { // s1 is the home
		sets[s1] = append(sets[s1],a2)
		return
	}

	if got2 { // s2 is the home
		sets[s2] = append(sets[s2],a1)
		return
	}

	// new pair, pick a key

	sets[a1] = append(sets[a1],a1)
	sets[a1] = append(sets[a1],a2)

}

// ****************************************************************************

func BelongsToLinSet(sets LinSet,member string) (bool,string,string) {

	for s := range sets {
		for m := range sets[s] {
			if member == sets[s][m] {
				var list string
				for l := range sets[s] {
					list = list + sets[s][l] + ","
				}
				return true,"super-"+s,list
			}
		}
	}

	return false,"",""
}

// ****************************************************************************

var GR_DAY_TEXT = []string{
        "Monday",
        "Tuesday",
        "Wednesday",
        "Thursday",
        "Friday",
        "Saturday",
        "Sunday",
    }
        
var GR_MONTH_TEXT = []string{
        "January",
        "February",
        "March",
        "April",
        "May",
        "June",
        "July",
        "August",
        "September",
        "October",
        "November",
        "December",
}
        
var GR_SHIFT_TEXT = []string{
        "Night",
        "Morning",
        "Afternoon",
        "Evening",
    }

// ****************************************************************************
// Semantic timeslots
// ****************************************************************************

func DoughNowt() (string,string) {

	// Time on the torus

	then := time.Now()

	year := fmt.Sprintf("Yr%d",then.Year())
	month := GR_MONTH_TEXT[int(then.Month())-1]
	day := then.Day()
	hour := fmt.Sprintf("Hr%02d",then.Hour())
	mins := fmt.Sprintf("Min%02d",then.Minute())
	quarter := fmt.Sprintf("Q%d",then.Minute()/15 + 1)
	shift :=  fmt.Sprintf("%s",GR_SHIFT_TEXT[then.Hour()/6])

	//secs := then.Second()
	//nano := then.Nanosecond()

	dayname := then.Weekday()
	dow := fmt.Sprintf("%.3s",dayname)
	daynum := fmt.Sprintf("Day%d",day)

        interval_start := (then.Minute() / 5) * 5
        interval_end := (interval_start + 5) % 60
        minD := fmt.Sprintf("Min%02d_%02d",interval_start,interval_end)

	var when string = fmt.Sprintf("%s,%s,%s,%s,%s at %s %s %s %s",shift,dayname,daynum,month,year,hour,mins,quarter,minD)

	var key string = fmt.Sprintf("%s:%s:%s",dow,hour,minD)

	return when, key
}

// ****************************************************************************

func Here(depth int) string {

        // Interal usage
	p,fname,line, ok := runtime.Caller(depth)
	
	var location string

	if ok {
		var funcname = runtime.FuncForPC(p).Name()
		fn := "function "+funcname
		file := "file "+fname
		lnr := fmt.Sprintf("line %d",line)
		location = fmt.Sprintf(" in %s of %s at %s",fn,file,lnr)
		
	} else {
		location = "unknown origin"
	}
	
	return location
}

// ****************************************************************************
// FRACTIONATION
// ****************************************************************************

func ReadAndCleanFile(filename string) string {

	content, _ := ioutil.ReadFile(filename)

	// Start by stripping HTML / XML tags before para-split
	// if they haven't been removed already

	m1 := regexp.MustCompile("<[^>]*>") 
	stripped1 := m1.ReplaceAllString(string(content),"") 

	//Strip and \begin \latex type commands

	m2 := regexp.MustCompile("\\\\[^ –\n]+") 
	stripped2 := m2.ReplaceAllString(stripped1," ") 

	// Non-English alphabet (tricky), but leave ?!:;

	m3 := regexp.MustCompile("[–{&}““”#%^+_#”=$’~‘/<>\"&]*") 
	stripped3 := m3.ReplaceAllString(stripped2,"") 

	m4 := regexp.MustCompile("[:;]+")
	stripped4 := m4.ReplaceAllString(stripped3,".")

	m5 := regexp.MustCompile("([^.,: ][\n])+")
	stripped5 := m5.ReplaceAllString(stripped4,"$0:")

	m6 := regexp.MustCompile("[^- a-zA-ZåøæÅØÆ.:,()!?\n]*")
	stripped6 := m6.ReplaceAllString(stripped5,"")

	m7 := regexp.MustCompile("[?!.]+")
	mark := m7.ReplaceAllString(stripped6,"$0#")

	m8 := regexp.MustCompile("[ \n]+")
	cleaned := m8.ReplaceAllString(mark," ")

	return cleaned
}

//**************************************************************

func FractionateSentences(text string) []Narrative {

	var sentences []string
	var selected_sentences []Narrative

	// Coordinatize the non-trivial sentences in terms of their ngrams

	if len(text) == 0 {
		return selected_sentences
	}

	sentences = SplitIntoSentences(text)

	var meaning = make([]float64,len(sentences))

	for s_idx := range sentences {

		meaning[s_idx] = FractionateThenRankSentence(s_idx,sentences[s_idx],len(sentences))
	}

	// Some things to note: importance tends to be clustered around the start and the end of
	// a story. The start is automatically weakner in this method, due to lack of data. We can
	// compensate by weighting the start and the end by sentence number.

	midway := len(sentences) / 2

	for s_idx := range sentences {

		scale_factor := 1.0 + float64((midway - s_idx) * (midway - s_idx)) / float64(midway*midway)

		n := NarrationMarker(sentences[s_idx], meaning[s_idx] * scale_factor, s_idx)
			
		selected_sentences = append(selected_sentences,n)
		
		ALL_SENTENCE_INDEX++
	}

return selected_sentences
}

//**************************************************************

func SplitIntoSentences(text string) []string {

	// Note this regex split has the effect of removing .?!

	re := regexp.MustCompile("#")
	sentences := re.Split(text, -1)

	var cleaned  = make([]string,1)

	for sen := range sentences{

		content := strings.Trim(sentences[sen]," ")

		if len(content) > 0 {			
			cleaned = append(cleaned,content)
		}
	}

	return cleaned
}

//**************************************************************

func FractionateThenRankSentence(s_idx int, sentence string, total_sentences int) float64 {

	var rrbuffer [MAXCLUSTERS][]string
	var sentence_meaning_rank float64 = 0
	var rank float64

	// split on any punctuation here, because punctuation cannot be in the middle
	// of an n-gram by definition of punctuation's promises
	// THIS IS A PT +/- constraint
	
	re := regexp.MustCompile("[,.;:!?]")
	sentence_frags := re.Split(sentence, -1)
	
	for f := range sentence_frags {
		
		// For one sentence, break it up into codons and sum their importances
		
		clean_sentence := strings.Split(string(sentence_frags[f])," ")
		
		for word := range clean_sentence {
			
			// This will be too strong in general - ligatures and foreign languages etc
			
			m := regexp.MustCompile("[/()?!]*") 
			cleanjunk := m.ReplaceAllString(clean_sentence[word],"") 
			cleanword := strings.Trim(strings.ToLower(cleanjunk)," ")
			
			if len(cleanword) == 0 {
				continue
			}

			// Shift all the rolling longitudinal Ngram rr-buffers by one word
			
			rank, rrbuffer = NextWordAndUpdateLTMNgrams(s_idx,cleanword, rrbuffer,total_sentences)
			sentence_meaning_rank += rank
		}
	}

	return sentence_meaning_rank
}

//**************************************************************

func RankByIntent(selected_sentences []Narrative) map[string]float64 {

	var topics = make(map[string]float64)

	sentences := len(selected_sentences)

	fmt.Println("--------- Sumarize ngram Intentionality threshold selection ---------------------------")

	for n := 1; n < MAXCLUSTERS; n++ {

		var last,delta int

		// Search through all sentence ngrams and measure distance between repeated
		// try to indentify any scales that emerge

		for ngram := range LTM_EVERY_NGRAM_OCCURRENCE[n] {

			occurrences := len(LTM_EVERY_NGRAM_OCCURRENCE[n][ngram])

			intent := Intentionality(n,ngram,sentences)

			if intent < 0.3  {
				continue
			}

			fmt.Println(n,ngram,occurrences,STM_NGRAM_RANK[n][ngram],"---------",intent)

			// if ngram of occurrences exceeds an expectation threshold in terms of length

			last = 0

			var min_delta int = 9999
			var max_delta int = 0
			var sum_delta int = 0

			for location := 0; location < occurrences; location++ {

				// Foreach occurrence, check proximity to others
				// This is about seeing if an ngram is a recurring input in the stream.
				// Does the subject recur several times over some scale? The scale may be
				// logarithmic like n / log (o1-o2) for occurrence separation
				// Radius = 100 sentences, how many occurrences of this ngram close together?
				
				// Does meaning have an intrinsic radius? It doesn't make sense that it
				// depends on the length of the document. How could we measure this?	
				
				// two one relative to first occurrence (absolute range), one to last occurrence??
				// only the last is invariant on the scale of a story
				
				delta = LTM_EVERY_NGRAM_OCCURRENCE[n][ngram][location] - last			
				last = LTM_EVERY_NGRAM_OCCURRENCE[n][ngram][location]

				sum_delta += delta

				if min_delta > delta {
					min_delta = delta
				}

				if max_delta < delta {
					max_delta = delta
				}
			}
			// which ngrams occur in bursty clusters. If completely even, then significance
			// is low or the theme of the whole piece. If cluster span/total span
			// max interdistance >> min interdistance then bursty
			
			if min_delta == 0 {
				continue
			}

			av_delta := float64(sum_delta)/float64(occurrences)

			if (av_delta > 3) && (av_delta < float64(LEG_WINDOW) * 4) {

				topics[ngram] = intent
			}
		}
	}

return topics
}

// *****************************************************************

func LongitudinalPersistentConcepts(topics map[string]float64) {
	
	fmt.Println("----- Emergent Longitudinally Stable Concept Fragments ---------")
	
	var sortable []Score
	
	for ngram := range topics {
		
		var item Score
		item.Key = ngram
		item.Score = topics[ngram]
		sortable = append(sortable,item)
	}
	
	sort.Slice(sortable, func(i, j int) bool {
		return sortable[i].Score < sortable[j].Score
	})
	
	// The score is the average interval between repetitions
	// If this is very long, the focus is spurious, so we look at the
	// shortest sample
	
	for i := 0; i < len(sortable); i++ {
		
		fmt.Printf("Particular theme/topic \"%s\" (= %f)\n", sortable[i].Key, sortable[i].Score)
	}
}

// *****************************************************************
	
func ReviewAndSelectEvents(filename string, selected_sentences []Narrative) {

	// The importances have now all been measured in realtime, but we review them now...posthoc
	// Now go through the history map chronologically, by sentence only reset the narrative  
        // `leg' counter when it fills up to measure story progress. 
	// This determines the sampling density of "important sentences" - pick a few from each leg

	fmt.Println(">>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>")
	fmt.Println("> Select inferred intentional content summary ...about",filename)
	fmt.Println(">>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>")

	var steps,leg int

	// Sentences to summarize per leg of the story journey

	steps = 0

	// We rank a leg by summing its sentence ranks

	var rank_sum float64 = 0
	var av_rank_for_leg []float64
	
	// First, coarse grain the narrative into `legs', 
        // i.e. standardized "narrative regions" by meter not syntax

	for s := range selected_sentences {

		// Make list of summed importance ranks for each leg

		rank_sum += selected_sentences[s].rank

		// Once we've summed all the importances and reached the end of the leg
		// define the leg_rank_average as the average over the interval and add this
		// to a list/array indexed by legs sequentially (append)

		if steps > LEG_WINDOW {

			steps = 0
			leg_rank_average := rank_sum / float64(LEG_WINDOW)
			av_rank_for_leg = append(av_rank_for_leg,leg_rank_average)
			rank_sum = 0
		}

		steps++	
	}

	// Don't forget any final "short" leg if there is one (residuals from the loop < LEG_WINDOW)

	leg_rank_average := rank_sum / float64(steps)
	av_rank_for_leg = append(av_rank_for_leg,leg_rank_average)

	// Find the leg of maximum importance

	var max_all_legs float64 = 0

	for l := range av_rank_for_leg {

		if max_all_legs < av_rank_for_leg[l] {

			max_all_legs = av_rank_for_leg[l]
		}
	}

	// Select a sampling rate that's lazy (> 1 sentence per leg) or busy ( <a few)
	// for important legs

	steps = 0
	leg = 0
	var this_leg_av_rank float64 = av_rank_for_leg[0]

	var sentence_id_by_rank = make(map[int]map[float64]int)
	sentence_id_by_rank[0] = make(map[float64]int)

	// Go through all the sentences that haven't been excluded and pick a sampling density that's
	// approximately evenly distributed-- split into LEG_WINDOW intervals

	for s := range selected_sentences {

		sentence_id_by_rank[leg][selected_sentences[s].rank] = s

		if steps > LEG_WINDOW {

			this_leg_av_rank = av_rank_for_leg[leg]

			// At the start of a long doc, there's insufficient weight to make an impact, so
			// we need to compensate by some arbitrary amount, this needs to be replaced by a ratio?
			// Based on word density...

			AnnotateLeg(filename, selected_sentences, leg, sentence_id_by_rank[leg], this_leg_av_rank, max_all_legs)

			steps = 0
			leg++

			sentence_id_by_rank[leg] = make(map[float64]int)
		}

		steps++
	}

	// Don't forget the final remainder (catch leg++)

	this_leg_av_rank = av_rank_for_leg[leg]
	
	AnnotateLeg(filename, selected_sentences, leg, sentence_id_by_rank[leg], this_leg_av_rank, max_all_legs)

	// Summarize	

	fmt.Println("------------------------------------------")
	fmt.Println("Notable events = ",KEPT,"of total ",ALL_SENTENCE_INDEX,"efficiency = ",100*float64(ALL_SENTENCE_INDEX)/float64(KEPT),"%")
	fmt.Println("------------------------------------------\n")
}

//**************************************************************

func Intentionality(n int, s string, sentence_count int) float64 {

	occurrences := STM_NGRAM_RANK[n][s]
	work := float64(len(s))
	legs := float64(sentence_count) / float64(LEG_WINDOW)

	if occurrences < 3 {
		return 0
	}

	if work < 5 {
		return 0
	}

	// lambda should have a cutoff for insignificant words, like "a" , "of", etc that occur most often

	lambda := occurrences / float64(LEG_WINDOW)

	// This constant is tuned to give words a growing importance up to a limit
	// or peak occurrences, then downgrade

	// Things that are repeated too often are not important
	// but length indicates purposeful intent

	meaning := lambda * work / (1.0 + math.Exp(lambda-legs))

return meaning
}

//**************************************************************

func AnnotateLeg(filename string, selected_sentences []Narrative, leg int, sentence_id_by_rank map[float64]int, this_leg_av_rank, max float64) {

	var sentence_ranks []float64
	var ranks_in_order []int

	key := make(map[float64]int)

	for fl := range sentence_id_by_rank {

		sentence_ranks = append(sentence_ranks,fl)
	}

	var samples_per_leg = len(sentence_ranks)

	if samples_per_leg < 1 {
		return
	}

	// Rank by importance and rescale all as dimensionless between [0,1]

	sort.Float64s(sentence_ranks)
	scale_free_trust := this_leg_av_rank / max

	// We now have an array of sentences whose indices are ascending ordered rankings, max = last
	// and an array of rankings min to max
	// Set up a key = sentence with rank = r as key[r]

	for i := range sentence_ranks {
		key[sentence_ranks[i]] = sentence_id_by_rank[sentence_ranks[i]]
	}

	// Select only the most important remaining in order for the hub
	// Hubs will overlap with each other, so some will be "near" others i.e. "approx" them
	// We want the degree of overlap between hubs TT.CompareContexts()

	fmt.Println("\n >> (Rank leg interest potential (anomalous intent)",leg,"=",scale_free_trust,")")

	// How do we quantitatively adjust output rate/velocity based on above threshold deviation

	var detail_per_leg_policy int

	if scale_free_trust > 0 { // Always true (legacy)

		var start int

		// Scale processing velocity like sqrt of probable mistrust event rate per leg

		detail_per_leg_policy = int(0.5 + math.Sqrt(float64(LEG_WINDOW) * scale_free_trust))

		fmt.Println(" >> (Dynamic kinetic event selection velocity", detail_per_leg_policy,"(events per leg)",LEG_WINDOW,")")

		// top intra_leg_sampling_density = count backwards from the end

		if samples_per_leg > detail_per_leg_policy {

			start = len(sentence_ranks) - detail_per_leg_policy

		} else {
			start = 0
		}

		for i :=  start; i < len(sentence_ranks); i++ {

			r := key[sentence_ranks[i]]
			ranks_in_order = append(ranks_in_order,r)
		}

		// Put the ranked selections back in sentence order

		sort.Ints(ranks_in_order)

	}

	// Now highest importance within the lef, in order of occurrence

	for r := range ranks_in_order {

		fmt.Printf("\nEVENT[Leg %d selects %d]: %s\n",leg,ranks_in_order[r],selected_sentences[ranks_in_order[r]].text)
		KEPT++
	}
}

//**************************************************************

func NextWordAndUpdateLTMNgrams(s_idx int, word string, rrbuffer [MAXCLUSTERS][]string,total_sentences int) (float64,[MAXCLUSTERS][]string) {

	// Word by word, we form a superposition of scores from n-grams of different lengths
	// as a simple sum. This means lower lengths will dominate as there are more of them
	// so we define intentionality proportional to the length also as compensation

	var rank float64 = 0

	for n := 2; n < MAXCLUSTERS; n++ {
		
		// Pop from round-robin

		if (len(rrbuffer[n]) > n-1) {
			rrbuffer[n] = rrbuffer[n][1:n]
		}
		
		// Push new to maintain length

		rrbuffer[n] = append(rrbuffer[n],word)

		// Assemble the key, only if complete cluster
		
		if (len(rrbuffer[n]) > n-1) {
			
			var key string
			
			for j := 0; j < n; j++ {
				key = key + rrbuffer[n][j]
				if j < n-1 {
					key = key + " "
				}
			}

			if ExcludedByBindings(rrbuffer[n][0],rrbuffer[n][n-1]) {

				continue
			}

			STM_NGRAM_RANK[n][key]++
			rank += Intentionality(n,key,total_sentences)

			LTM_EVERY_NGRAM_OCCURRENCE[n][key] = append(LTM_EVERY_NGRAM_OCCURRENCE[n][key],s_idx)

		}
	}

	STM_NGRAM_RANK[1][word]++
	rank += Intentionality(1,word,total_sentences)

	LTM_EVERY_NGRAM_OCCURRENCE[1][word] = append(LTM_EVERY_NGRAM_OCCURRENCE[1][word],s_idx)

	return rank, rrbuffer
}

//**************************************************************
// MISC
//**************************************************************

func NarrationMarker(text string, rank float64, index int) Narrative {

	var n Narrative

	n.text = text
	n.rank = rank
	n.index = index

return n
}

//**************************************************************

func ExcludedByBindings(firstword,lastword string) bool {

	// A standalone fragment can't start/end with these words, because they
	// Promise to bind to something else...
	// Rather than looking for semantics, look at spacetime promises only - words that bind strongly
	// to a prior or posterior word.

	if (len(firstword) == 1) || len(lastword) == 1 {
		return true
	}

	for s := range FORBIDDEN_ENDING {
		if lastword == FORBIDDEN_ENDING[s] {
			return true
		}
	}

	for s := range FORBIDDEN_STARTER {
		if firstword == FORBIDDEN_STARTER[s] {
			return true
		}
	}

return false 
}

