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

var CONTENTION_USER_PLUS = make(map[string]int)
var CONTENTION_USER_MINUS = make(map[string]int)
var CONTENTION_SUBJ = make(map[string]int)

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

	// subjects := ReadSubjects("wiki_samples_total.in")

	subjects := []string{ "Laser" }

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

		AnalyzeTopic(subjects[n])
	}
}

//**************************************************************

func usage() {

        fmt.Fprintf(os.Stderr, "usage: go run wikipedia_history.go [verbose]\n")
        flag.PrintDefaults()
        os.Exit(2)
}

// ***********************************************************

func AnalyzeTopic(subject string) {

	//page_url := "https://en.wikipedia.org/wiki/" + subject
	log_url := "https://en.wikipedia.org/w/index.php?title="+subject+"&action=history&offset=&limit=1000"

	// Go straight to discussion (user behaviour)

	TT.LEG_WINDOW = 10
	TT.LEG_SELECTIONS = make([]string,0)

	changelog := HistoryPage(log_url)

	sort.Slice(changelog, func(i, j int) bool {
		return changelog[i].Date.Before(changelog[j].Date)
	})

	// Look at signals from text analysis

	HistoryAssessment(subject,changelog)

	historypage := TotalText(changelog)

	remarks,ltm := TT.FractionateSentences(historypage)
	
	TT.ReviewAndSelectEvents(subject + " edit history",remarks)		
	
	topics := TT.RankByIntent(remarks,ltm)
	
	TT.LongitudinalPersistentConcepts(topics)
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

func HistoryAssessment(subject string, changelog []WikiProcess) {

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

	TT.Println("\n==============================================\n")
	TT.Println("HISTORY OF CHANGE ANALYSIS: Starting assessment of history for",subject)
	TT.Println("\n==============================================\n")

	allepisodes[episode] = make(map[string]int)

	burststart = changelog[0].Date.UnixNano()

	for i := 0; i < len(changelog); i++ {

		//fmt.Printf(">> %15s (%v)(%d), %s --> %s\n", changelog[i].User,changelog[i].Date,changelog[i].EditDelta,changelog[i].Message,changelog[i].DiffUrl)

		// Setup lists of edits for each user

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

			// Generate Adjacency Matrix for Group (range episode_users) and principal eigenvector

			// Reset for next episode

			episode_key := fmt.Sprintf("%s_ep_%d",subject,episode)

			fmt.Println("\nlink users:",episode_users)
			fmt.Println("\n..to average state context on scale of episode (contenious or cooperative)")
			fmt.Println("\nlink message ngrams",changelog[i].Message)

			ep := TT.NextDataEvent(&G,subject,"episode",episode_key,changelog[i].Message,int64(delta_t),burststart,burstend)

			fmt.Println("\nsubject name and ngrams are part of state context",subject)

			fmt.Println("\nLEARN grov context association to GOOD/BAD/UGLY")

			//raather than good bad ugly: do, undo, (degree of text overlap)


			if episode == 1 {
				LinkEpisodeChainAndSpectrumToTopic(ep,subject,burststart,burstend)
			}

			LinkUsersToEpisode(episode_users,ep)

			fmt.Println("\nCOMPUTE DIFF FRACTION and ngram content - check against ngram signal strength")
			LinkDiffFractionsToEpisode(ep,changelog[i].DiffUrl)

			sum_burst_bytes = 0
			burst_size = 0

			if i < len(changelog)-1 {
				burststart = changelog[i+1].Date.UnixNano()
			}

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

		if changelog[i].Revert > 0 && i > 1 {
			
			users_revert[changelog[i].User] += changelog[i].Revert

			if last_user != changelog[i].User {
				fmt.Println("STATE .. Explicit undo of",last_user,"by",changelog[i].User)
				fmt.Println("LEARN DIRECTED RELN",last_user,"by",changelog[i].User)
				ARTICLE_ISSUES++
				CONTENTION_USER_PLUS[changelog[i].User]++
				CONTENTION_USER_MINUS[last_user]++
				CONTENTION_SUBJ[subject]++
			}

			dt := float64(changelog[i].Date.UnixNano() - changelog[i-1].Date.UnixNano())
			users_revert_dt[changelog[i].User] = 0.6 * dt + 0.4 * users_revert_dt[changelog[i].User]
		}

		// This is a real undo if the next change cancels 90% of the previous

		if math.Abs(float64(changelog[i].EditDelta + last_delta)) < float64(last_delta)/10.0  {

			ARTICLE_ISSUES++
			fmt.Println(" .. Effective undo of",last_user,"by",changelog[i].User)
			fmt.Println("LEARN DIRECTED RELN",last_user,"by",changelog[i].User)
			users_revert[changelog[i].User]++
			CONTENTION_USER_PLUS[changelog[i].User]++
			CONTENTION_USER_MINUS[last_user]++
			CONTENTION_SUBJ[subject]++
		}

		fmt.Println("HIGH INTENTIONALITY NGRAMS ... associated with an attention level")
		fmt.Println("DEFINE: a working set of context signals that we can return from a query")
		fmt.Println("DEFINE:   - some instantaneous characters")
		fmt.Println("DEFINE:   - some graph inferences based on agent ID and trigger words")

		fmt.Println("UPDATE: weights/importance on nodes and links")
		fmt.Println("------------------------------------------------")

		last_delta = changelog[i].EditDelta
		last_user = changelog[i].User
	}

	TT.Println("\n----------- EDITS ANALYSIS --------------------")

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

	// If a users changes are ALL reversions, they are police

	TT.Println("\n**************************")
	TT.Println("> Infer user promise/intent")
	TT.Println("> 100% changes are reversions, then they are police")
	TT.Println("> 30% of changes are reversions contentious")
	TT.Println("**************************\n")

	for s := range users {

		// If all the changes were reversions without additions, policing user, 
		// if more than say a third then just contentious

		if users_revert[users[s]] == users_changecount[users[s]] {

			TT.Printf(" POLICING  %20s (%d) of %d after average of %3.2f mins\n",
				users[s],users_revert[users[s]],
				users_changecount[users[s]],
				users_revert_dt[users[s]]/MINUTE)

			LinkSignalToUser(users[s],"correctional")

		} else if users_revert[users[s]] > 1 && float64(users_revert[users[s]]) / float64(users_changecount[users[s]]) > 0.3 {

			TT.Printf(" CONTENTIOUS  %20s (%d) of %d after average of %3.2f mins\n",
				users[s],
				users_revert[users[s]],
				users_changecount[users[s]],
				users_revert_dt[users[s]]/MINUTE)

			LinkSignalToUser(users[s],"contentious")

		}
	}

	//av_burst_size := float64(sum_burst_size + burst_size) / float64(episode)

	//active_users := len(users_changecount)


}

// *******************************************************************************

func PlotUserBursts(histogram map[int]int, filename string) {
	
	// sum the groups intoa histogram

	f, err := os.OpenFile(filename,os.O_CREATE|os.O_WRONLY, 0644)

	if err != nil {
		fmt.Println("Couldn't open for write/append to",filename,err)
		return
	}

	var n_tot float64 = 0

	for n := range histogram {
		n_tot += float64(histogram[n])
	}

	for n := range histogram {

		// Workgroup events from TT.Set: size of an aggregate associative cluster (potentially growing)
		// size of cluster, frequency/how many, log size, log frequency

		h := float64(histogram[n])

		s := fmt.Sprintf("%f %f %f %f\n",float64(n),float64(h/n_tot),math.Log(float64(n)),math.Log(float64(h/n_tot)))

		_, err = f.WriteString(s)

		if err != nil {
			fmt.Println("Couldn't write/append to",filename,err)
		}
	}

	f.Close()
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

// **************************************************************************

func LinkPersistentToSubject(subject string, concepts map[string]float64) {

	var count int = 0

	n_from := TT.CreateNode(G,"topic",subject,subject,0.0,0,0,0)

	// First add the story samples

	var last TT.Node = n_from

	fmt.Println(" - adding story selections x",len(TT.LEG_SELECTIONS))

	for event := range TT.LEG_SELECTIONS {

		count++

		key := TT.KeyName(subject+"_story",count)
		
		if len(key) < TT.MIN_LEGAL_KEYNAME {
			continue
		}

		this := TT.CreateNode(G,"event",key,TT.LEG_SELECTIONS[event],0,0,0,0)

		TT.CreateLink(G, last,"LEADS_TO", this, 0)
		//Connect the concept to the episode it occurred in

		LinkAllNgramsFromTo(TT.LEG_SELECTIONS[event],this)

		last = this
	}

	// Link the longitudinal persistent concepts

	count = 0

	fmt.Println(" - adding story concepts x",len(concepts))

	// These are the surviving ngrams that we need to further fractionate

	for frag := range concepts {

		words := strings.Count(frag," ") + 1
		collection := fmt.Sprintf("ngram%d",words)

		// Put the concept fragment in its node collection

		key := TT.KeyName(frag,0)

		if len(key) < TT.MIN_LEGAL_KEYNAME {
			continue
		}

		frag_node := TT.CreateNode(G,collection,key,frag,0,0,0,0)

		// Make sure all concepts also take us to the topic subject

		TT.CreateLink(G,n_from,"TALKSABOUT", frag_node, 0)

		LinkAllNgramsFromTo(frag,frag_node)
	}
}

// **************************************************************************

func LinkAllNgramsFromTo(concept string,org_node TT.Node) {
	
	words := strings.Split(concept," ")

	var rrbuffer [TT.MAXCLUSTERS][]string

	for word := range words {
		
		// This will be too strong in general - ligatures and foreign languages etc
		
		if len(words[word]) == 0 {
			continue
		}
		
		// Shift all the rolling longitudinal Ngram rr-buffers by one word
		
		rrbuffer = NextWordAndUpdateNgrams(concept,org_node,words[word],rrbuffer)
	}
}

// **************************************************************************

func NextWordAndUpdateNgrams(original string,org_node TT.Node,word string, rrbuffer [TT.MAXCLUSTERS][]string) [TT.MAXCLUSTERS][]string {

	// Word by word, we form a superposition of scores from n-grams of different lengths
	// as a simple sum. This means lower lengths will dominate as there are more of them
	// so we define intentionality proportional to the length also as compensation

	for n := 2; n < TT.MAXCLUSTERS; n++ {
		
		// Pop from round-robin

		if (len(rrbuffer[n]) > n-1) {
			rrbuffer[n] = rrbuffer[n][1:n]
		}
		
		// Push new to maintain length

		rrbuffer[n] = append(rrbuffer[n],word)

		// Assemble the key, only if complete cluster
		
		if (len(rrbuffer[n]) > n-1) {
			
			var partial string
			
			for j := 0; j < n; j++ {
				partial = partial + rrbuffer[n][j]
				if j < n-1 {
					partial = partial + " "
				}
			}

			if TT.ExcludedByBindings(rrbuffer[n][0],rrbuffer[n][n-1]) {

				continue
			}

			if partial != original {
				LinkFragToFrag(n,partial,org_node)
			}
		}
	}

	if word != original {
		LinkFragToFrag(1,word,org_node)

	}

	return rrbuffer
}


// **************************************************************************

func LinkFragToFrag(n int, part string,org_node TT.Node) {

	part_key := TT.KeyName(part,0)

	if len(part_key) < TT.MIN_LEGAL_KEYNAME {
		return
	}

	coll := fmt.Sprintf("ngram%d",n)

	part_node := TT.CreateNode(G,coll,part_key,part,TT.STM_NGRAM_RANK[n][part],0,0,0)

	TT.CreateLink(G, org_node, "CONTAINS", part_node, 0)
}

// **************************************************************************

func LinkUsersToEpisode(usernames map[string]int,ep TT.Node) {

	for user := range usernames {

		name := TT.CanonifyName(user)

		if len(name) < TT.MIN_LEGAL_KEYNAME {
			continue
		}

		n_from := TT.CreateNode(G,"user",name,name,0.0,0,0,0)
		TT.CreateLink(G, n_from, "INFL", ep,0)
	}
}

// **************************************************************************

func LinkEpisodeChainAndSpectrumToTopic(ep TT.Node, subject string, begin,end int64) {

	n_from := TT.CreateNode(G,"topic",subject,"",0.0,0,begin,end)
	TT.CreateLink(G, n_from, "THEN", ep,0)
}

// **************************************************************************

func LinkSignalToUser(username,signal string) {

	name := TT.KeyName(username,0)

	if len(name) < TT.MIN_LEGAL_KEYNAME {
		return
	}

	n_from := TT.CreateNode(G,"user",name,username,0.0,0,0,0)
	n_to := TT.CreateNode(G,"signal",signal,signal,0.0,0,0,0)

	TT.CreateLink(G, n_from, "EXPRESSES", n_to,0)
}

// ***********************************************************

func LinkDiffFractionsToEpisode(n_from TT.Node, url string) {

	difftext_2 := DiffPage(url)

	difftext_1 := strings.ReplaceAll(difftext_2,"[[","")
	difftext := strings.ReplaceAll(difftext_1,"]]","")

	// Strip [[..]]

	edits,ltm := TT.FractionateSentences(difftext)
	concepts := TT.RankByIntent(edits,ltm)

	fmt.Println("\nHERE ARE THE MAIN TOPICAL NGRAMS",concepts)
	var count int = 0

	for t := range concepts {

		fmt.Println("DIFF...analysis")

		if strings.Count(t," ") < 2 {
			continue
		}

		count++
		key := TT.KeyName(t,count)

		if len(key) < TT.MIN_LEGAL_KEYNAME {
			continue
		}

		n := strings.Count(t," ") + 1

		LinkFragToFrag(n,key,n_from)

		fmt.Println("NGRAMS",key,"-->",n_from)
	}
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

			if token.Data == "body" {

				return total
			}
		}
	}
}



