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

type UserProfile struct {       // A list of all edit events
	
	Imposing int
	Imposed  int
	Edits    int
	Topics   map[string]int
}

var CONTENTION_USER_IMPOSING = make(map[string]int)
var CONTENTION_USER_IMPOSED = make(map[string]int)
var CONTENTION_LAST_USER_EDIT = make(map[string]int64)
var CONTENTION_USER_TOPICS = make(map[string]int)

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

	subjects := ReadSubjects("wiki_samples.in")
	//subjects := ReadSubjects("wiki_samples_short_test.in")

	//subjects := []string{ "Laser" }

	// ***********************************************************
	
	TT.InitializeSmartSpaceTime()

	var dbname string = "SST-ML"
	var dburl string = "http://localhost:8529"
	var user string = "root"
	var pwd string = "mark"

	G = TT.OpenAnalytics(dbname,dburl,user,pwd)

	// Load any pretraining

	for n := 1; n < TT.MAXCLUSTERS; n++ {
		TT.LoadNgram(G,n)
	}

	// ***********************************************************

	for n := range subjects {

		fmt.Println(n,subjects[n],"...")

		// Don't look at the whole text now
		// ngram_ctx := AnalyzeTopicContext(subjects[n])

		var ngram_ctx [TT.MAXCLUSTERS]map[string]float64
 		
		AnalyzeTopicProcess(subjects[n],ngram_ctx)
	}
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

	for class := range freq {
		keys = append(keys,class)
	}

	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})

	name := "../data/UserData/" + title + ".output"

	f, err := os.OpenFile(name,os.O_CREATE|os.O_WRONLY, 0644)

	if err != nil {
		fmt.Println("Couldn't open for write/append to",name,err)
		return
	}

	var s string

	for i := range keys {
		s = fmt.Sprintln(i,freq[keys[i]])
		_, err = f.WriteString(s)

		if err != nil {
			fmt.Println("Couldn't write/append to",name,err)
		}
	}

	f.Close()

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

	if changelog == nil {
		return
	}

	sort.Slice(changelog, func(i, j int) bool {
		return changelog[i].Date.Before(changelog[j].Date)
	})

	// Look at signals from text analysis

	HistoryAssessment(subject,changelog,ngram_ctx)

	historypage := TotalText(changelog)

	// Keep learning

	remarks,ltm := TT.FractionateSentences(historypage)
	
	TT.ReviewAndSelectEvents(subject + " edit history",remarks)		
	
	topics := TT.RankByIntent(remarks,ltm)
	
	invariants := TT.LongitudinalPersistentConcepts(topics)

	SaveProcessInvariants(invariants)
}

// ***********************************************************

func SaveProcessInvariants(invariants [TT.MAXCLUSTERS]map[string]float64) {

	TT.SaveNgrams(G,invariants)
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
				plaintext += "# "
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
					TT.Println("Time parsing error",err)
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
				
			if token.Data == "a" {
				difftext = true
			}

		case html.EndTagToken:
		
			switch token.Data {

			case "td":
			case "p":
				entry.Message += "# "
			}

			if token.Data == "body" {

				return changelog				
			}
		}
	}
}

// *******************************************************************************

func HistoryAssessment(subject string, changelog []WikiProcess, ngram_ctx [TT.MAXCLUSTERS]map[string]float64) {

	var last_delta int = 0
	var all_users_averagetime float64 = 0
	var delta_t float64 = 0
	var episode int = 1
	var burststart int64
	var sum_burst_bytes float64 = 0
	var context = make(map[string]int)
	var lasttime float64
	var cumulative_message string 
	var add,rm string
	const punctuation_scale = 5.0
	const min_episode_duration = DAY

	//var last_overlap int = 0

	burststart = changelog[0].Date.UnixNano()

	//name := subject // this is really a context label. We could also derive  MTWTFSS from date
	//ctx := TT.StampedPromiseContext_Begin(G, TT.KeyName(name,0), changelog[0].Date)

	// Parse past timeline as Stamped History

	for i := 0; i < len(changelog); i++ {

		sum_burst_bytes += math.Abs(float64(changelog[i].EditDelta))

		delta_t = float64(changelog[i].Date.UnixNano()) - lasttime
		all_users_averagetime = 0.4 * all_users_averagetime + 0.6 * delta_t
		lasttime = float64(changelog[i].Date.UnixNano())
		burst_duration := changelog[i].Date.UnixNano() - burststart

		_,nadd,nrm := GetDiffFractionsForEpisode(changelog[i].DiffUrl)

		add += nadd
		rm += nrm
		cumulative_message += changelog[i].Message + " "

		if math.Abs(float64(changelog[i].EditDelta + last_delta)) < float64(last_delta)/10.0  {

			ARTICLE_ISSUES++

			Extend(context,"effective_undo")
			Extend(context,"state_of_contention")
			Extend(context,"state_of_uncertainty_about_article")
		}

		// *****

		if (i == len(changelog)-1) || (burst_duration > int64(min_episode_duration)) && (delta_t > all_users_averagetime) && (delta_t > min_episode_duration) {

			// The appropriate unit is the episode

			Extend(context,subject)

			/* Let's assume this expensive check is not good investment...
			ov,_ := FindNgramOverlap(ngram_ctx,ngram_prc)
			overlap := len(ov)
			interfer_rate := overlap - last_overlap
			last_overlap = len(ov)

			const threshold = 10

			if interfer_rate * interfer_rate > threshold {
				Extend(context,"observed_attention_change")
			}*/

			sum_burst_bytes = 0

			if i < len(changelog)-1 {
				burststart = changelog[i+1].Date.UnixNano()
			}

			// Checks

			trust_level := AssessChanges(context,add,rm,cumulative_message)
			Extend(context,trust_level)

			fmt.Println("CONTEXT tick",context)

			context = make(map[string]int)

			cumulative_message = ""
			add = ""
			rm = ""

			episode++
		}
	}
}

// *******************************************************************************

func IsAnonymous(user string) bool {

	// Is the userID probably an IP address?

	if strings.Count(user,".") == 3 || strings.Count(user,":") > 2 {
		return true
	}

	return false
}

// *******************************************************************************

func AssessChanges(context map[string]int,add,rm,message string) string {

	add_len := len(add)
	rm_len := len(rm)

	//  Learn these and compare probability, we need a fixed vector
	
	// TT.ASSESS_EXCELLENT_S = "trust_high" 1.0
	// TT.ASSESS_PAR_S = "trust_ok"   0.5
	// TT.ASSESS_WEAK_S = "trust_low" 0.25
	// TT.ASSESS_SUBPART_S = "untrusted" 0

	if add_len == 0 && rm_len == 0 {
		return TT.ASSESS_PAR_S
	}

	// The following are all heuristics

	const change_limit_bytes = 1000

	var assess_s string = TT.ASSESS_PAR_S
	var assess float64 = TT.ASSESS_PAR

	delta := add_len - rm_len

	if delta > change_limit_bytes || delta < -change_limit_bytes {
		assess_s = TT.ASSESS_WEAK_S
		assess = TT.ASSESS_WEAK
		Extend(context,"large_edit")
	}

	if rm_len > add_len * 10 {
		assess_s = TT.ASSESS_WEAK_S
		assess = TT.ASSESS_WEAK
		Extend(context,"large_deletion")
	}

	if rm_len > change_limit_bytes {
		assess_s = TT.ASSESS_SUBPAR_S
		assess = TT.ASSESS_SUBPAR
		Extend(context,"large_deletion")
	}


	var bad_signals = []string{"fuck","cunt","bastard","unhelpful","unfair","too","deceiv","deceptive","terrorism","justice","unsourced"}
	// Pick some arbitrary signals

	var bad_flag = false

	for s := range bad_signals {
		if strings.Contains(strings.ToLower(message),bad_signals[s]) {
			bad_flag = true
		}
	}

	intent := TT.StaticIntent(G,message)
	
	// How do we understand this threshold? We need to compare it to something similar, meta not subject

	if intent > 20000 || bad_flag{
		fmt.Printf("\n Anomalous message intent -- (%s) = %f\n\n",message,intent)
		Extend(context,"anomalous_message")
	}

	if assess < TT.ASSESS_PAR {

		if len(rm) > 100 {
			fmt.Printf(" --> RM %.20s ...(len=%d)\n",rm,len(rm))
		}

		if strings.Contains(add,"http") {
			Extend(context,"url_warning")
		} else if len(add) > 100 {
			fmt.Printf(" <-- ADD %.20s ...(len=%d)\n",add,len(add))
		}
	}

	return assess_s
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

func GetDiffFractionsForEpisode(url string) ([TT.MAXCLUSTERS]map[string]float64,string,string) {

	text,additions,removals := DiffPage(url)

	ngrams := TT.FractionateText2Ngrams(text)

	return ngrams, additions, removals
}

// ***********************************************************

func DiffPage(url string) (string,string,string) {

	//test_url := "https://en.wikipedia.org/w/index.php?title=Promise_theory&diff=prev&oldid=1139754849"

	prefix := "https://en.wikipedia.org"

	response, err := http.Get(prefix+url)

	if err != nil {
		fmt.Println(err)
	}

	defer response.Body.Close()

	// Parse HTML

	var attend bool = false
	var insert bool = false
	var remove bool = false
	var total,del,add string 

	// Start parsing

	tokenizer := html.NewTokenizer(response.Body)

	for {
		tokenType := tokenizer.Next()
		token := tokenizer.Token()

		if tokenizer.Err() == io.EOF {
			return "","",""
		}

		// Strip out junk characters

		r := regexp.MustCompile("(<.+>)|([0-9]+px)|([a-zA-Z0-9]+=[A-Z0-9a-z]+)|.+\\.png|.+\\.PNG|[0-9abcdef][0-9abcdef][0-9abcdef][0-9abcdef][0-9abcdef][0-9abcdef]|[^a-zA-Z0-9 ]+")
		s := strings.TrimSpace(html.UnescapeString(token.String()))
		s = r.ReplaceAllString(s," ")
		s = strings.ReplaceAll(s,"→"," ")
		s = strings.ReplaceAll(s,"←"," ")
		s = strings.ReplaceAll(s,"'"," ")
		s = strings.ReplaceAll(s,"{"," ")
		s = strings.ReplaceAll(s,"}","")
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
			return "","",""
			
		case html.TextToken:

			if attend && len(s) > 0 && !IsCode(s) {				
				total += s + " "
			}

			if insert && len(s) > 0 {
				// This is the sum of deletions/insertions
				add += s + " "
				insert = false
			}

			if remove && len(s) > 0 {
				del += s + " "
				remove = false
			}

		case html.StartTagToken:

			switch token.Data {

			case "ins":
				insert = true
			case "del":
				remove = true
			}
				
		case html.EndTagToken:

			switch token.Data {

			case "td":
			case "p":
				total += "# "
			}

			if token.Data == "body" {

				return total,add,del
			}
		}
	}
}



