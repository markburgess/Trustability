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
// To run, edit the list of pages if necessary and simply run 
// (takes a long time to complete and generates a lot of output)
// It generates/appends to a file /tmp/trust.dat (should not exist in advance)
//
//      go run wikipedia_history_ml.go
// ***********************************************************

package main

import (
	"fmt"
	"golang.org/x/net/html"
	"net/http"
	"sort"
	"strings"
	"regexp"
	"time"
	"math"
	"io"
	"TT"
	"os"
	"bufio"
	"flag"
)

// ***********************************************************

type WikiProcess struct {       // A list of all edit events
	
	Date      time.Time
	User      string
	EditSize  int
	EditDelta int
	Message   string
	Revert    int
	DiffUrl  string
}

// ***********************************************************

const DAY = float64(3600 * 24 * TT.NANO)
const MINUTE = float64(60 * TT.NANO)

var G TT.Analytics
var ARTICLE_ISSUES int = 0

var EPISODE_CLUSTER_FREQ = make(map[int]int)
var NOBOTS bool = false

type UserProfile struct {       // A list of all edit events
	
	Imposing int
	Imposed  int
	Edits    int
	Topics   map[string]int
}

var CONTENTION_USER_IMPOSING = make(map[string]int)
var CONTENTION_USER_IMPOSED = make(map[string]int)
var CONTENTION_USER_EDITS = make(map[string]int)
var CONTENTION_USER_TOPICS = make(map[string]map[string]int)

// ***********************************************************

func main() {

        flag.Usage = usage
        flag.Parse()
        args := flag.Args()

	if len(args) == 1 && (args[0] != "verbose" || args[0] != "-v") {
		TT.VERBOSE = true
	} else if len(args) > 0 {
                usage()
                os.Exit(1);
        }

	// ***********************************************************

	// Example pages, some familiar some notorious

	subjects := ReadSubjects("wiki_samples_short_test.in")

	//subjects := []string{ "Laser" }

	// ***********************************************************
	
	TT.InitializeSmartSpaceTime()

	var dbname string = "SST-ML"
	var dburl string = "http://localhost:8529"
	var user string = "root"
	var pwd string = "mark"

	G = TT.OpenAnalytics(dbname,dburl,user,pwd)

	// ***********************************************************

	for n := range subjects {

		fmt.Println(n,subjects[n],"...")

		ngram_ctx := AnalyzeTopicContext(subjects[n])
		
		AnalyzeTopicProcess(subjects[n],ngram_ctx)
	}

	//fmt.Println("\nContentious",CONTENTION_USER_IMPOSING)
	//fmt.Println("\nNo confidence",CONTENTION_USER_IMPOSED)

	ov,tot := FindOverlap(CONTENTION_USER_IMPOSING,CONTENTION_USER_IMPOSED)
	Freq(CONTENTION_USER_IMPOSING,"imposing/attack")
	Freq(CONTENTION_USER_IMPOSED,"imposed/no confidence")
	fmt.Println("\nOverlap of contention",len(ov),"/",tot)
}

//**************************************************************

func usage() {

        fmt.Fprintf(os.Stderr, "usage: go run wikipedia_history.go [verbose]\n")
        flag.PrintDefaults()
        os.Exit(2)
}

// ***********************************************************

func Freq(data map[string]int,title string){

	var freq = make(map[int]float64)
	var keys []int

	for val := range data {
		freq[data[val]]++
	}

	fmt.Println(title)

	for class := range freq {
		keys = append(keys,class)
	}

	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})

	for i := range keys {
		fmt.Println(" -- freq[",i,"] = ",freq[keys[i]])
	}
}

// ***********************************************************

func AnalyzeTopicContext(subject string) [TT.MAXCLUSTERS]map[string]float64 {

	page_url := "https://en.wikipedia.org/wiki/" + subject

	TT.LEG_WINDOW = 100           // Standard for narrative text
	TT.LEG_SELECTIONS = make([]string,0)

	mainpage := MainPage(page_url)

	selected,ltm := TT.FractionateSentences(mainpage)

	TT.ReviewAndSelectEvents(subject,selected)		

	pagetopics := TT.RankByIntent(selected,ltm)

	return TT.LongitudinalPersistentConcepts(pagetopics)
}
// ***********************************************************

func AnalyzeTopicProcess(subject string, ngram_ctx [TT.MAXCLUSTERS]map[string]float64) {

	log_url := "https://en.wikipedia.org/w/index.php?title="+subject+"&action=history&offset=&limit=1000"

	// Go straight to discussion (user behaviour)

	TT.LEG_WINDOW = 10
	TT.LEG_SELECTIONS = make([]string,0)

	changelog := HistoryPage(log_url)

	sort.Slice(changelog, func(i, j int) bool {
		return changelog[i].Date.Before(changelog[j].Date)
	})

	// Look at signals from text analysis

	HistoryAssessment(subject,changelog,ngram_ctx)

	historypage := TotalText(changelog)

	remarks,ltm := TT.FractionateSentences(historypage)
	
	TT.ReviewAndSelectEvents(subject + " edit history",remarks)		
	
	topics := TT.RankByIntent(remarks,ltm)
	
//	inv := 
TT.LongitudinalPersistentConcepts(topics)
}

// ***********************************************************

func FindOverlap(a,b map[string]int) (map[string]bool,int) {

	var overlap = make(map[string]bool)
	var tot int = 0

	for inv1 := range a {
		
		tot += len(b)
		
		for inv2 := range b {
			
			if inv1 == inv2 {
				overlap[inv1] = true
				delete(b,inv2)
			}
		}
	}

	return overlap, tot
}


// ***********************************************************

func FindNgramOverlap(a,b [TT.MAXCLUSTERS]map[string]float64) (map[string]bool,int) {

	var overlap = make(map[string]bool)
	var tot int = 0

	for n := 1; n < TT.MAXCLUSTERS; n++ {

		for inv1 := range a[n] {

			tot += len(b[n])

			for inv2 := range b[n] {
				
				if inv1 == inv2 {
					overlap[inv1] = true
					delete(b[n],inv2)
				}
			}
		}
	}

	return overlap, tot
}

// ***********************************************************

func ReadSubjects(filename string) []string {

	var list []string
	
	file, err := os.Open(filename)
	
	if err != nil {
		fmt.Println("Error opening",filename,":",err)
		return list
	}
	
	defer file.Close()
	
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {

		s := scanner.Text()
		s = strings.ReplaceAll(s,"–","-")
		list = append(list,s)
	}

	fmt.Println("Reading",len(list),"topics in",filename)
	
	return list
}

// ***********************************************************

func MainPage(url string) string {

	var capture bool = false

	// If at first you don't succeed, the network timed out

	response, err := http.Get(url)

	if err != nil {
		for {
			
			response, err = http.Get(url)
			
			if err == nil {
				break
			}
			
			fmt.Println("Retrying (",err,")")
			time.Sleep(1)
		}
	}

	defer response.Body.Close()
	
	// Parse HTML
	
	var plaintext string = ""
	
	tokenizer := html.NewTokenizer(response.Body)
	
	for {
		tokenType := tokenizer.Next()
		token := tokenizer.Token()
		
		if tokenizer.Err() == io.EOF {
			fmt.Println("EOT",err)
			return plaintext
		}

		// Strip out unwanted characters and mark end of sentence with a # symbol
		

		s := strings.TrimSpace(html.UnescapeString(token.String()))
		s = TT.HashcodeSentenceSplit(s)
		s = strings.ReplaceAll(s,"→","")
		s = strings.ReplaceAll(s,"←","")
		s = strings.ReplaceAll(s,"'","")
		s = strings.ReplaceAll(s,"{{","")
		s = strings.ReplaceAll(s,"}}","")
		s = strings.ReplaceAll(s,"(","")
		s = strings.ReplaceAll(s,")","")
		s = strings.ReplaceAll(s,"|"," ")
		s = strings.ReplaceAll(s,"\n"," ")
		s = strings.TrimSpace(html.UnescapeString(s))
		s = strings.TrimSpace(s)
		
		for i := range token.Attr {
			
			if token.Attr[i].Val == "mw-content-text" {				
				capture = true
			}
		}
		if len(s) == 0 {
			continue
		}
		
		switch tokenType {
			
		case html.ErrorToken:
			
			fmt.Printf("Error: %v\n", tokenizer.Err())
			return plaintext
			
		case html.TextToken:
			
			if capture {
				
				if Bracketed(s) {
					continue
				}
				
				if IsCode(s) {
					continue
				}
				
				//fmt.Printf("Token-body: %v\n", s)
				
				if IsLegal(s) {
					continue
				}

				if strings.Contains(s,"remove this template message") {
					// Ignore administrivia
					continue
				}
				
				if s == "citation needed" {
					ARTICLE_ISSUES++
					continue
				}

				if strings.Contains(s,"vandalism") {
					ARTICLE_ISSUES++
				}
				
				if s == "edit" {
					continue
				}
				
				
				if s == "References" {
					return plaintext
				}
				
				// Strip commas etc for the n-grams

				plaintext += strings.TrimSpace(s) + " "

			}
			
		case html.StartTagToken:

		case html.EndTagToken:
		
			switch token.Data {

			case "h1":
			case "h2":
			case "td":
			case "p":
				plaintext += "#"
			}

			if token.Data == "body" {
				capture = false
				return plaintext
			}
		}
	}
	
	return plaintext
}

// ***********************************************************

func HistoryPage(url string) []WikiProcess {

	response, err := http.Get(url)

	if err != nil {
		fmt.Println(err)
	}

	defer response.Body.Close()

	// Parse HTML

	var attend bool = false
	var after_editsize = false
	var after_edits = false
	var message string
	var date = false
	var user = false
	var difftext bool = false
	var history int = 0
	var entry WikiProcess
	var changelog []WikiProcess

	// Start parsing

	tokenizer := html.NewTokenizer(response.Body)

	for {
		tokenType := tokenizer.Next()
		token := tokenizer.Token()

		if tokenizer.Err() == io.EOF {
			return changelog
		}

		// Strip out junk characters

		r := regexp.MustCompile("[?!.][ \n]")
		s := strings.TrimSpace(html.UnescapeString(token.String()))
		s = r.ReplaceAllString(s,"$0#")
		s = strings.ReplaceAll(s,"→","")
		s = strings.ReplaceAll(s,"←","")
		s = strings.ReplaceAll(s,"'","")
		s = strings.ReplaceAll(s,"{{","")
		s = strings.ReplaceAll(s,"}}","")
		s = strings.ReplaceAll(s,"(","")
		s = strings.ReplaceAll(s,")","")
		s = strings.ReplaceAll(s,"|"," ")
		s = strings.ReplaceAll(s,"No edit summary","")
		s = strings.ReplaceAll(s,"External links:","")
		s = strings.TrimSpace(html.UnescapeString(s))
		s = strings.TrimSpace(s)

		for i := range token.Attr {

			if token.Attr[i].Val == "mw-contributions-list" {				
				history++
			}

			if token.Attr[i].Val == "mw-userlink" {
				user = true
			}

			if token.Attr[i].Val == "new mw-userlink" {
				user = true
			}

			if token.Attr[i].Val == "mw-userlink mw-anonuserlink" {
				user = true
			}

			if token.Attr[i].Val == "mw-changeslist-date" {				
				date = true
			}

			if difftext {

				// Use this anchor as the reset trigger
				if token.Attr[i].Key == "href" && strings.Contains(s,"diff=prev") {
					attend = true
					var empty WikiProcess
					entry = empty
					after_edits = false
					after_editsize = true
					user = false
					difftext = false
					message = ""
					entry.DiffUrl = token.Attr[i].Val
				}
			}
		}
		
		switch tokenType {
			
		case html.ErrorToken:
			
			fmt.Printf("Error: %v", tokenizer.Err())
			return changelog
			
		case html.TextToken:

			// End of the list is an image box to compare versions

			if attend && date {

				date = false

				t, err := time.Parse("15:04, 2 January 2006", s)

				if err != nil{
					fmt.Println("ERR",err)
				}

				entry.Date = t
				continue
			}

			if attend && user {
				s = strings.ReplaceAll(s,"#","") // remove formatting from IP addresses
				entry.User = s
				user = false
				continue
			}

			if attend && strings.HasSuffix(s,"bytes") {

				var bytes int = 0
				s = strings.ReplaceAll(s,",","") // remove formatting
				fmt.Sscanf(s,"%d",&bytes)

				if bytes > 0 {
					entry.EditSize = bytes
					after_editsize = true
				}

			} else if after_editsize && attend {

				// Past the editsize, must be delta

				if entry.EditSize > 0 && entry.EditDelta == 0 {

					var delta int = 0
					s = strings.ReplaceAll(s,"−","-") // remove unicode
					fmt.Sscanf(s,"%d",&delta)

					if delta != 0 {
						entry.EditDelta = delta
						after_editsize = false
						after_edits = true
						continue
					}
				}
			}

			if attend && after_edits {

				if strings.Contains(s,"Revert") {
					entry.Revert++
				}

				if attend && (strings.HasPrefix(s,"Tag") || strings.HasPrefix(s,"bot") || strings.HasPrefix(s,"New page:") ||strings.HasPrefix(s,"Category:") || strings.HasPrefix(s,"undo") || strings.HasPrefix(s,"cur")|| strings.HasPrefix(s,"<img")) {
				} else if len(s) > 0 {

					message += strings.TrimSpace(s) + " "
				}
			}

			if attend && (strings.HasPrefix(s,"undo") || strings.HasPrefix(s,"cur") || strings.HasPrefix(s,"<img")) {

				attend = false
				entry.Message = strings.TrimSpace(message) + ". "

				if NOBOTS && IsBot(entry.User) {
					fmt.Println("Skipping",entry.User)
				} else {
					changelog = append(changelog,entry)
				}
			}

		case html.StartTagToken:
				
			if token.Data == "a" {
				difftext = true
			}

		case html.EndTagToken:
		
			switch token.Data {

			case "td":
			case "p":
				entry.Message += "#"
			}

			if token.Data == "body" {

				return changelog				
			}
		}
	}
}

// *******************************************************************************

func HistoryAssessment(subject string, changelog []WikiProcess, ngram_ctx [TT.MAXCLUSTERS]map[string]float64) {

	var users_changecount = make(map[string]int)
	var users_revert = make(map[string]int)
	var users_lasttime = make(map[string]int64)
	var users_averagetime = make(map[string]float64)	
	var users_revert_dt = make(map[string]float64)
	var users []string
	var last_delta int = 0
	var last_user string
	var lasttime float64 = 0
	var user_delta_t float64
	var all_users_averagetime float64 = float64(MINUTE)
	var delta_t float64 = float64(MINUTE)
	var burst_size int = 0
	var sum_burst_size int = 0
	var sum_burst_bytes float64 = 0
	var episode int = 1
	var event int = 1
	var burststart,burstend int64

	var allusers = make(map[string][]int64)
	var allepisodes = make(map[int]map[string]int)
	var episode_duration = make(map[int]int64)
	var episode_bytes = make(map[int]float64)
	var episode_users = make(map[string]int)

	var episode_user_start = make(map[string]int)
	var episode_user_last = make(map[string]int)

	var all_invariants [TT.MAXCLUSTERS]map[string]float64

	var context = make(map[string]int)

	var last_overlap int = 0

	for i := 1; i < TT.MAXCLUSTERS; i++ {
		all_invariants[i] = make(map[string]float64)
	} 

	allepisodes[episode] = make(map[string]int)

	burststart = changelog[0].Date.UnixNano()

	for i := 0; i < len(changelog); i++ {

		episode_users[changelog[i].User]++
		allusers[changelog[i].User] = append(allusers[changelog[i].User],changelog[i].Date.UnixNano())
		allepisodes[episode][changelog[i].User]++

		sum_burst_bytes += math.Abs(float64(changelog[i].EditDelta))

		// Bootstrap difference

		if users_lasttime[changelog[i].User] == 0 {
			user_delta_t = float64(MINUTE)
			users_averagetime[changelog[i].User] = user_delta_t
		}

		// For each user independently

		if users_lasttime[changelog[i].User] > 0 {

			user_delta_t = float64(changelog[i].Date.UnixNano() - users_lasttime[changelog[i].User])

			if user_delta_t < 0 {
				user_delta_t = 0
			}

			// Update running average delta t per user
			users_averagetime[changelog[i].User] = 0.4 * users_averagetime[changelog[i].User] + 0.6 * user_delta_t
		}

		// For all users collectively

		delta_t = float64(changelog[i].Date.UnixNano()) - lasttime

		if delta_t < 0 {
			delta_t = 0
		}

		// Last time is running value to enable computing time interval since last (delta)

		users_lasttime[changelog[i].User] = changelog[i].Date.UnixNano()
		lasttime = float64(changelog[i].Date.UnixNano())

		// Keep track of how many edits in this burst, before reset below
		burst_size++

		// End of burst

		const punctuation_scale = 10.0
		const min_episode_duration = int64(DAY)

		if episode_user_start[changelog[i].User] == 0 {
			episode_user_start[changelog[i].User] = event
		}

		episode_user_last[changelog[i].User] = event

		// Episodes are a longer timescale variability of state

		// Demarcate episode boundary *********************************************
		// We need a minimum size for a burst to protect against average being zero

		burstend = changelog[i].Date.UnixNano()
		last_duration := burstend - burststart

		if (i == len(changelog)-1) || (delta_t > float64(min_episode_duration)) && (last_duration > min_episode_duration) && (delta_t > all_users_averagetime * punctuation_scale) {

			sum_burst_size += burst_size
			episode_duration[episode] = last_duration
			episode_bytes[episode] = sum_burst_bytes

			Extend(context,subject)

			// Check for adabatic focal change (symbol interferometry)
			// This is overlap with subject material for context alignment

			ngram_prc := LinkDiffFractionsToEpisode(changelog[i].DiffUrl)
			ov,_ := FindNgramOverlap(ngram_ctx,ngram_prc)
			overlap := len(ov)
			interfer_rate := overlap - last_overlap
			last_overlap = len(ov)

			const threshold = 10

			if interfer_rate * interfer_rate > threshold {
				Extend(context,"observed_attention_change")
			}

			sum_burst_bytes = 0
			burst_size = 0

			if i < len(changelog)-1 {
				burststart = changelog[i+1].Date.UnixNano()
			}

			//fmt.Println("CONTEXT",context,"\n")
			//fmt.Println("USERS:",episode_users)

			context = make(map[string]int)
			episode++
			allepisodes[episode] = make(map[string]int)
			EPISODE_CLUSTER_FREQ[len(episode_users)]++
			episode_users = make(map[string]int)

			event = 1
			episode_user_start = make(map[string]int)
			episode_user_last = make(map[string]int)
		}

		// Demarcate episode boundary *********************************************

		// Update running average for all users

		all_users_averagetime = 0.4 * all_users_averagetime + 0.6 * delta_t

		// Changes with reversions

		users_changecount[changelog[i].User]++

		if CONTENTION_USER_TOPICS[changelog[i].User] == nil {
			CONTENTION_USER_TOPICS[changelog[i].User] = make(map[string]int)
		}

		if changelog[i].Revert > 0 && i > 1 {
			
			users_revert[changelog[i].User] += changelog[i].Revert

			CONTENTION_USER_EDITS[changelog[i].User]++

			if last_user != changelog[i].User {

				TT.Println("    (STATE .. Explicit undo of",last_user,"by",changelog[i].User,")")
				Extend(context,"state_of_contention")
				Extend(context,"explicit_undo")
				Extend(context,last_user)
				Extend(context,changelog[i].User)

				// article trustworthiness, update Node context - if balanced high level of activity
				/// if no activity, maybe untrustworthy..
				// User trustworthiness
				// How do we allocate trust?

				CONTENTION_USER_IMPOSING[changelog[i].User]++
				CONTENTION_USER_IMPOSED[last_user]++
				CONTENTION_USER_TOPICS[changelog[i].User][subject]++
			}

			dt := float64(changelog[i].Date.UnixNano() - changelog[i-1].Date.UnixNano())
			users_revert_dt[changelog[i].User] = 0.6 * dt + 0.4 * users_revert_dt[changelog[i].User]
		}

		// This is a real undo if the next change cancels 90% of the previous

		if math.Abs(float64(changelog[i].EditDelta + last_delta)) < float64(last_delta)/10.0  {

			ARTICLE_ISSUES++

			TT.Println("  (.. Effective undo of",last_user,"by",changelog[i].User,")")

			users_revert[changelog[i].User]++

			CONTENTION_USER_IMPOSING[changelog[i].User]++
			CONTENTION_USER_IMPOSED[last_user]++
			CONTENTION_USER_TOPICS[changelog[i].User][subject]++

			Extend(context,"effective_undo")
			Extend(context,"state_of_contention")
			Extend(context,"state_of_uncertainty_about_article")
			Extend(context,last_user)
			Extend(context,TT.CanonifyName(changelog[i].User))
		}

		last_delta = changelog[i].EditDelta
		last_user = changelog[i].User
	}

	TT.Println("\n----------- USER BEHAVIOUR ANALYSIS --------------------")

	// Get an idempotent list of all users for this topic's history

	for username := range users_changecount {
		users = append(users,username)
	}

	// Sort by number of changes

	sort.Slice(users, func(i, j int) bool {
		return users_changecount[users[i]] > users_changecount[users[j]]
	})

	TT.Println("\nRanked number of user changes: number and average time interval")

	var bots,humans float64 = 0,0

	for s := range users {

		if IsBot(users[s]) {
			bots++
		} else {
			humans++
		}

		if users_changecount[users[s]] > 1 {

			TT.Printf("  > %20s  (%2d)   av_delta %-3.2f (days)\n",
				users[s],
				users_changecount[users[s]],
				users_averagetime[users[s]]/float64(DAY))
		} else {

			TT.Print(users[s],", ")

		}
	}

	TT.Println("\n\nReversions (agents exhibiting contentious behaviour)")

	users = nil

	for s := range users_revert {
		users = append(users,s)
	}

	sort.Slice(users, func(i, j int) bool {
		return users_revert[users[i]] > users_revert[users[j]]
	})

	for s := range users {
		TT.Printf(" R  %20s (%d) of %d after average of %3.2f mins\n",
			users[s],
			users_revert[users[s]],
			users_changecount[users[s]],
			users_revert_dt[users[s]]/MINUTE)
	}

	for s := range users {

		// If all the changes were reversions without additions, policing user, 
		// if more than say a third then just contentious

		if users_revert[users[s]] == users_changecount[users[s]] {

			TT.Printf(" POLICING  %20s (%d) of %d after average of %3.2f mins\n",
				users[s],users_revert[users[s]],
				users_changecount[users[s]],
				users_revert_dt[users[s]]/MINUTE)

		} else if users_revert[users[s]] > 1 && float64(users_revert[users[s]]) / float64(users_changecount[users[s]]) > 0.3 {

			TT.Printf(" CONTENTIOUS  %20s (%d) of %d after average of %3.2f mins\n",
				users[s],
				users_revert[users[s]],
				users_changecount[users[s]],
				users_revert_dt[users[s]]/MINUTE)
		}
	}
}

// *******************************************************************************

func Merge(delta,total [TT.MAXCLUSTERS]map[string]float64) {

	for n := 1; n < TT.MAXCLUSTERS; n++ {

		for ngram := range delta[n] {
			total[n][ngram] = delta[n][ngram]
		}
	} 
}

// *******************************************************************************

func Extend(context map[string]int,s string) {

	context[s]++
}

// *******************************************************************************

func TotalText(changelog []WikiProcess) string {

	var text string = ""

	for i := range changelog {

		text += changelog[i].Message

	}

	return text
}

// **************************************************************************
// Some helpers
// **************************************************************************

func Bracketed(s string) bool {

	if strings.HasPrefix(s,"[") || strings.HasSuffix(s,"]") {

		return true
	}

return false
}

// **************************************************************************

func IsCode(s string) bool {

	if strings.Contains(s,"{") || strings.HasSuffix(s,"}")  || strings.HasSuffix(s,"^") {

		return true
	}

	if strings.Contains(s,"=") {
		return true
	}

	if strings.Contains(s,"http") || strings.Contains(s,"/") || strings.HasSuffix(s,":") {

		return true
	}

	if strings.Contains(s,"Categories :") || strings.Contains(s,"Hidden categories :") {
		return true
	}
return false
}

// **************************************************************************

func IsLegal(s string) bool {

	if strings.Contains(s,"terms may appl") {

		return true
	}

	if strings.Contains(s,"Creative Commons") {

		return true
	}

	if strings.Contains(s,"last edited") {

		return true
	}

	if strings.Contains(s,"Terms of Use") {

		return true
	}

	if strings.Contains(s,"Privacy Policy") {

		return true
	}



return false
}

// **************************************************************************

func IsBot(s string) bool {

	if strings.Contains(s,"Bot") || strings.HasSuffix(s,"bot") {

		return true
	}

return false
}

// ***********************************************************

func LinkDiffFractionsToEpisode(url string) [TT.MAXCLUSTERS]map[string]float64 {

	var ngrams [TT.MAXCLUSTERS]map[string]float64

	for i := 1; i < TT.MAXCLUSTERS; i++ {
		ngrams[i] = make(map[string]float64)
	}

	difftext_3 := DiffPage(url)
	difftext_2 := strings.ReplaceAll(difftext_3,"\n","")
	difftext_1 := strings.ReplaceAll(difftext_2,"[[","")
	difftext_0 := strings.ReplaceAll(difftext_1,"]]","")
	search := "\\[[0-9]+"
	r := regexp.MustCompile(search)
	tmp := r.ReplaceAllString(difftext_0,"")
	difftext := strings.TrimSpace(tmp)

	edits,ltm := TT.FractionateSentences(difftext)
	concepts := TT.RankByIntent(edits,ltm)

	for t := range concepts {

		n := strings.Count(t," ") + 1

		if n < 3 && n > 5 {
			continue
		}

		ngrams[n][t] = concepts[t]

		//fmt.Println("SUPPORT:: edit_content_invariant...\"",n,t,"\"")
	}

	return ngrams
}

// ***********************************************************

func DiffPage(url string) string {

	//test_url := "https://en.wikipedia.org/w/index.php?title=Promise_theory&diff=prev&oldid=1139754849"

	prefix := "https://en.wikipedia.org"

	response, err := http.Get(prefix+url)

	if err != nil {
		fmt.Println(err)
	}

	defer response.Body.Close()

	// Parse HTML

	var attend bool = false
	var total string 

	// Start parsing

	tokenizer := html.NewTokenizer(response.Body)

	for {
		tokenType := tokenizer.Next()
		token := tokenizer.Token()

		if tokenizer.Err() == io.EOF {
			return ""
		}

		// Strip out junk characters

		r := regexp.MustCompile("<.+>")
		s := strings.TrimSpace(html.UnescapeString(token.String()))
		s = r.ReplaceAllString(s," ")
		s = strings.ReplaceAll(s,"→","")
		s = strings.ReplaceAll(s,"←","")
		s = strings.ReplaceAll(s,"'","")
		s = strings.ReplaceAll(s,"{{","")
		s = strings.ReplaceAll(s,"}}","")
		s = strings.ReplaceAll(s,"("," ")
		s = strings.ReplaceAll(s,")"," ")
		s = strings.ReplaceAll(s,"|"," ")
		s = strings.ReplaceAll(s,"No edit summary","")
		s = strings.ReplaceAll(s,"External links:","")
		s = strings.TrimSpace(html.UnescapeString(s))
		s = strings.TrimSpace(s)

		for i := range token.Attr {

			if token.Attr[i].Val == "diff-context diff-side-deleted" {
				attend = false
			}

			if token.Attr[i].Val == "diff-context diff-side-added" {
				attend = true
			}
		}
		
		switch tokenType {
			
		case html.ErrorToken:
			
			fmt.Printf("Error: %v", tokenizer.Err())
			return ""
			
		case html.TextToken:

			if attend && len(s) > 0 && !IsCode(s) {				
				total += s + " "
			}

		case html.StartTagToken:
				
		case html.EndTagToken:

			switch token.Data {

			case "td":
			case "p":
				total += "#"
			}

			if token.Data == "body" {

				return total
			}
		}
	}
}



