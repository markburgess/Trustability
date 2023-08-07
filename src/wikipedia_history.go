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
//      go run wikipedia_history.go
//
// To generate graphs, install GNUplot and run script in gnuplot.in
//
// This is tuned specifically to Wikipedia scanning. using the general methods
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
}

// ***********************************************************

const DAY = float64(3600 * 24 * 1000000000)
const MINUTE = float64(60 * 1000000000)
const OUTPUT_FILE = "/tmp/trust.dat"

var G TT.Analytics
var ARTICLE_ISSUES int = 0

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
	


	// Example pages, some familiar some notorious

	subjects := ReadSubjects("wiki_samples.in")

	for n := range subjects {

		result := AnalyzeTopic(subjects[n])
		TT.AppendStringToFile(OUTPUT_FILE,result)
	}

	fmt.Println("\nWrote",len(subjects),"lines to graph table:\n",OUTPUT_FILE)
}

//**************************************************************

func usage() {

        fmt.Fprintf(os.Stderr, "usage: go run wikipedia_history.go [verbose]\n")
        flag.PrintDefaults()
        os.Exit(2)
}

// ***********************************************************

func AnalyzeTopic(subject string) string {
	
	page_url := "https://en.wikipedia.org/wiki/" + subject
	log_url := "https://en.wikipedia.org/w/index.php?title="+subject+"&action=history&offset=&limit=1000"

	// ***********************************************************
	
	TT.InitializeSmartSpaceTime()

	var dbname string = "SemanticSpacetime"
	var dburl string = "http://localhost:8529"
	var user string = "root"
	var pwd string = "mark"

	G = TT.OpenAnalytics(dbname,dburl,user,pwd)

	// ***********************************************************

	TT.LEG_WINDOW = 100           // Standard for narrative text

	mainpage := MainPage(page_url)
	
	textlength := len(mainpage)

	selected := TT.FractionateSentences(mainpage)

	TT.Println("*********************************************")
	TT.Println("* Mainpage for",subject,"-- length",textlength,"chars")
	TT.Println("* Sentences",len(selected))
	TT.Println("* Legs",float64(len(selected))/float64(TT.LEG_WINDOW))
	TT.Println("*********************************************")
	
	TT.ReviewAndSelectEvents(subject,selected)		
	
	pagetopics := TT.RankByIntent(selected)
	
	TT.LongitudinalPersistentConcepts(pagetopics)

	// ***********************************************************

	// Examine the change log

	TT.LEG_WINDOW = 10 // Need a smaller window than normal for fragmented text

	changelog := HistoryPage(log_url)

	sort.Slice(changelog, func(i, j int) bool {
		return changelog[i].Date.Before(changelog[j].Date)
	})

	// Look at signals

	history_users, episodes, avt, avep, useredits, usergroups, episode_duration, episode_bytes := HistoryAssessment(subject,changelog)

	historypage := TotalText(changelog)

	talklength := len(historypage)

	remarks := TT.FractionateSentences(historypage)

	TT.Println("*********************************************")
	TT.Println("* Historypage length",subject,talklength)
	TT.Println("* Sentences",len(remarks))
	TT.Println("* Legs",float64(len(remarks))/float64(TT.LEG_WINDOW))
	TT.Println("* Total users involved in shared process", history_users)
	TT.Println("* Change episodes with discernable punctuation", episodes)
	TT.Println("* The average time between changes is",avt/float64(MINUTE),"mins",avt/float64(DAY),"days")
	TT.Println("*********************************************")
	
	TT.ReviewAndSelectEvents(subject + " edit history",remarks)		
	
	topics := TT.RankByIntent(remarks)
	
	TT.LongitudinalPersistentConcepts(topics)

	TT.Println("\n****** User/agent persistence on this page ....\n")

	for u := range useredits {
		lifetime := float64(useredits[u][len(useredits[u])-1]-useredits[u][0])/float64(DAY)
		if lifetime > 1 {
			TT.Println(" Lifetime ",u,lifetime,"days")
		}
	}

	TT.Println("\n******* EDITING EPISODIC BURSTS....\n")

	var average_cluster float64 = 0

	for g := 1; g <= len(usergroups); g++ {
		duration := float64(episode_duration[g])/float64(DAY)
		users := float64(len(usergroups[g]))
		average_cluster += users/float64(len(usergroups))
		TT.Println("\n",g," Episode with",users,"users\n        ",usergroups[g],"\n        duration (days)",duration,"\n        dur/user",duration/users,"\n        Byte changes",episode_bytes[g],"\n        Changes/user",episode_bytes[g]/users)
	}


	I := float64(ARTICLE_ISSUES)            // counted altercations
	IL := math.Log(I)

	N := average_cluster                // av users per episode
	NL := math.Log(N)

	N2 := (N*N-N)
	N2L := math.Log(N2)

	L := float64(textlength)            // article length in sentences
	LL := math.Log(L)

	H := float64(talklength)            // change process in sentence/entries
	//HL := math.Log(H)

	s := float64(len(remarks))          // changes subsampled on trust
	S := float64(len(selected))         // article subsampled on trust

	e := s/H  // trusted process watch list fraction
	E := S/L  // trusted article watch list fraction

	// work and sampled(untrusted)

	w := H/L
	wL := math.Log(w)
	u := s/S
	uL := math.Log(u)

	TT.Println("\n*********************************************")
	TT.Println("* SUMMARY")
	TT.Println("* Mainpage for",subject,"-- length",textlength,"chars")
	TT.Printf("* Total contentious article assessments for %s = %d\n",subject,ARTICLE_ISSUES)
	TT.Printf("* I/L = Contention x1000 per unit length = %.2f\n",1000*float64(ARTICLE_ISSUES)/float64(textlength))
	TT.Printf("* Contention x1000 per user = %.2f\n",1000*float64(ARTICLE_ISSUES)/float64(history_users))
	TT.Printf("* Contention x1000 per user^2 = %.2f\n",1000*float64(ARTICLE_ISSUES)/float64(history_users*history_users))
	TT.Println("* Process history length =",len(remarks))
	TT.Println("* Process history length / article length =",float64(talklength)/float64(textlength))
	TT.Println("* Process selections / article selections =",float64(len(remarks))/float64(len(selected)))
	TT.Println("* Efficiency History/Article  =",e/E)
	TT.Println("* Total users involved in shared process", history_users)
	TT.Println("* Average user cluster size per episode", average_cluster)
	TT.Println("* Change episodes with discernable punctuation", episodes)
	TT.Println("* Average episode size (notes/remarks)", avep)
	TT.Println("* The average time between changes is",avt/float64(MINUTE),"mins",avt/float64(DAY),"days")
	TT.Println("*********************************************\n")

	output := fmt.Sprintln(
		L,   // 1 text
		LL,  // 2
		N,   // 3 users
		NL,  // 4
		N2,  // 5 users-cluster
		N2L, // 6
		I,   // 7
		IL,  // 8
 		w,  // 9
		wL, // 10
		u,  // 11
		uL, // 12
	)
	// Look for the dynamics of the change process as fn of L and N
	// I(N)  -> (3,7), exp(3,8), pow(4,8)
	// I(N2) -> (5,7), exp(5,8), pow(6,8) 
	// I(L)  -> (1,7), exp(1,8), pow(2,8)
	
	// Relaxation of history activity per length of article H/L and its trusted fraction s/S
	// R(L) as time goes on, L converges and issues come to a halt, L is a proxy for tau
	// Rw(L)  -> (1,9), exp(1,10), pow(2,10)
	// Ru(L)  -> (1,11), exp(1,12), pow(2,12)

	return output
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

	response, err := http.Get(url)

	if err != nil {
		fmt.Println(err)
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
		
		r := regexp.MustCompile("[?!.]+")
		s := strings.TrimSpace(html.UnescapeString(token.String()))
		s = r.ReplaceAllString(s,"$0#")
		s = strings.ReplaceAll(s,"→","")
		s = strings.ReplaceAll(s,"←","")
		s = strings.ReplaceAll(s,"'","")
		s = strings.ReplaceAll(s,"{{","")
		s = strings.ReplaceAll(s,"}}","")
		s = strings.ReplaceAll(s,"(","")
		s = strings.ReplaceAll(s,")","")
		s = strings.ReplaceAll(s,"|","")
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

				plaintext += strings.TrimSpace(s) + ". "

			}
			
		case html.StartTagToken:
			
		case html.EndTagToken:
			
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

		r := regexp.MustCompile("[?!.]+")
		s := strings.TrimSpace(html.UnescapeString(token.String()))
		s = r.ReplaceAllString(s,"$0#")
		s = strings.ReplaceAll(s,"→","")
		s = strings.ReplaceAll(s,"←","")
		s = strings.ReplaceAll(s,"'","")
		s = strings.ReplaceAll(s,"{{","")
		s = strings.ReplaceAll(s,"}}","")
		s = strings.ReplaceAll(s,"(","")
		s = strings.ReplaceAll(s,")","")
		s = strings.ReplaceAll(s,"|","")
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
		}
		
		switch tokenType {
			
		case html.ErrorToken:
			
			fmt.Printf("Error: %v", tokenizer.Err())
			return changelog
			
		case html.TextToken:

			if s == "prev" {

				attend = true
				var empty WikiProcess
				entry = empty
				after_edits = false
				after_editsize = true
				user = false
				message = ""
			}			

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
				changelog = append(changelog,entry)
			}

		case html.StartTagToken:

		case html.EndTagToken:

			if token.Data == "body" {

				return changelog				
			}
		}
	}
}

// *******************************************************************************

func HistoryAssessment(subject string, changelog []WikiProcess) (int,int,float64,float64,map[string][]int64,map[int]map[string]int,map[int]int64,map[int]float64) {

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
	var episodes int = 1
	var burststart int64

	var allusers = make(map[string][]int64)
	var allepisodes = make(map[int]map[string]int)
	var episode_duration = make(map[int]int64)
	var episode_bytes = make(map[int]float64)

	TT.Println("\n==============================================\n")
	TT.Println("CHANGE ANALYSIS: Starting assessment of history for",subject)
	TT.Println("\n==============================================\n")

	TT.Println("\n----------- EDITS --------------------")

	allepisodes[episodes] = make(map[string]int)

	burststart = changelog[0].Date.UnixNano()

	for i := range changelog {

		//fmt.Printf(">> %15s (%v)(%d), %s\n", changelog[i].User,changelog[i].Date,changelog[i].EditDelta,changelog[i].Message)

		// Setup lists of edits for each user

		allusers[changelog[i].User] = append(allusers[changelog[i].User],changelog[i].Date.UnixNano())
		allepisodes[episodes][changelog[i].User]++

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

		const punctuation_scale = 10.0

		// End of burst

		if delta_t > all_users_averagetime * punctuation_scale {

			sum_burst_size += burst_size
			episode_duration[episodes] = changelog[i].Date.UnixNano() - burststart
			episode_bytes[episodes] = sum_burst_bytes
			sum_burst_bytes = 0
			burst_size = 0
			burststart = changelog[i].Date.UnixNano()
			episodes++
			allepisodes[episodes] = make(map[string]int)
		}

		// Update running average for all users
		all_users_averagetime = 0.4 * all_users_averagetime + 0.6 * delta_t

		// Changes with reversions

		users_changecount[changelog[i].User]++

		if changelog[i].Revert > 0 && i > 1 {
			
			users_revert[changelog[i].User] += changelog[i].Revert

			if last_user != changelog[i].User {
				TT.Println(" .. Explicit undo of",last_user,"by",changelog[i].User)
				ARTICLE_ISSUES++
			}

			dt := float64(changelog[i].Date.UnixNano() - changelog[i-1].Date.UnixNano())
			users_revert_dt[changelog[i].User] = 0.6 * dt + 0.4 * users_revert_dt[changelog[i].User]
		}

		// This is a real undo if the next change cancels 90% of the previous

		if math.Abs(float64(changelog[i].EditDelta + last_delta)) < float64(last_delta)/10.0  {

			ARTICLE_ISSUES++
			TT.Println(" .. Effective undo of",last_user,"by",changelog[i].User)
			users_revert[changelog[i].User]++
		}

		last_delta = changelog[i].EditDelta
		last_user = changelog[i].User
	}

	TT.Println("\n----------- EDITS ANALYSIS --------------------")

	for s := range users_changecount {
		users = append(users,s)
	}

	sort.Slice(users, func(i, j int) bool {
		return users_changecount[users[i]] > users_changecount[users[j]]
	})

	TT.Println("\nRanked number of user changes: number and average time interval")

	for s := range users {
		if users_changecount[users[s]] > 1 {
			TT.Printf("  > %20s  (%2d)   av_delta %-3.2f (days)\n",users[s],users_changecount[users[s]],users_averagetime[users[s]]/float64(DAY))
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
		TT.Printf(" R  %20s (%d) of %d after average of %3.2f mins\n",users[s],users_revert[users[s]],users_changecount[users[s]],users_revert_dt[users[s]]/MINUTE)
	}

	// If a users changes are ALL reversions, they are police

	TT.Println("\n**************************")
	TT.Println("> Infer user promise/intent")
	TT.Println("> 100% changes are reversions, then they are police")
	TT.Println("> 30% of changes are reversions contentious")
	TT.Println("**************************\n")

	for s := range users {

		// If all the changes were reversions without additions, policing user, if more than say a third then just contentious

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

	av_burst_size := float64(sum_burst_size + burst_size) / float64(episodes)

	return len(users_changecount), episodes, all_users_averagetime, av_burst_size, allusers, allepisodes, episode_duration, episode_bytes
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
