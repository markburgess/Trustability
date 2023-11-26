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

var AVERAGE_INTENT float64
var ALIGNMENT = make(map[int]int)

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

	subjects := ReadSubjects("wiki_samples_total.in")
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

	SaveAlignmentSpectrum("../data/ML/alignment_spectrum")
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

func SaveAlignmentSpectrum(filename string) {

	var keys []int

	for class := range ALIGNMENT {
		keys = append(keys,class)
	}

	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})

	for a := range keys {
		data := fmt.Sprintf("%d %d\n",a,ALIGNMENT[keys[a]])
		TT.AppendStringToFile(filename,data)
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
	ARTICLE_ISSUES = 0

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
	TT.Println("!!!!!!!!!!",topics)	
//	invariants := TT.LongitudinalPersistentConcepts(topics)
//	SaveProcessInvariants(invariants)
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
	var episode_len int = 0
	var burststart int64
	var edit_balance float64 = 0
	var sum_burst_bytes float64 = 0
	var context = make(map[string]int)
	var lasttime float64
	var cumulative_message string 
	var add,rm string
	const punctuation_scale = 5.0
	const min_episode_duration = 2*DAY

	//var last_overlap int = 0

	burststart = changelog[0].Date.UnixNano()

	ctx := TT.StampedPromiseContext_Begin(G, TT.KeyName(subject,0), changelog[0].Date)

	// Parse past timeline as Stamped History

	for i := 0; i < len(changelog); i++ {

		// Measure the signal bandwidth
		episode_len++
		sum_burst_bytes += math.Abs(float64(changelog[i].EditDelta))
		edit_balance += float64(changelog[i].EditDelta)

		// Measure the signal rate
		delta_t = float64(changelog[i].Date.UnixNano()) - lasttime
		all_users_averagetime = 0.4 * all_users_averagetime + 0.6 * delta_t
		lasttime = float64(changelog[i].Date.UnixNano())
		burst_duration := float64(changelog[i].Date.UnixNano() - burststart)

		/* Ignore the content and try the WORK amounts only - trust the platform balance report in EdDelta

                _,nadd,nrm := GetDiffFractionsForEpisode(changelog[i].DiffUrl)

		add += nadd
		rm += nrm */

		// Signal content
		cumulative_message += changelog[i].Message + " "

		TT.LearnWeeklyKV(G, "interactions", changelog[i].Date.Unix(), 1.0)

		if changelog[i].Revert > 0 && i > 1 {

			ARTICLE_ISSUES++

			Extend(context,"explicit_undo")
			Extend(context,"state_of_contention")
			Extend(context,"state_of_uncertainty_about_article")
			TT.LearnWeeklyKV(G, "contention", changelog[i].Date.Unix(), 1.0)
		}

		if math.Abs(float64(changelog[i].EditDelta + last_delta)) < float64(last_delta)/10.0  {

			ARTICLE_ISSUES++

			Extend(context,"effective_undo")
			Extend(context,"state_of_contention")
			Extend(context,"state_of_uncertainty_about_article")
			TT.LearnWeeklyKV(G, "contention", changelog[i].Date.Unix(), 1.0)
		}

		// *****

		if (i == len(changelog)-1) || (burst_duration > min_episode_duration) && (delta_t > all_users_averagetime) && (delta_t > min_episode_duration * punctuation_scale) {

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

			if i < len(changelog)-1 {
				burststart = changelog[i+1].Date.UnixNano()
			}

			// Checks

			if sum_burst_bytes != 0 {
				trust_level := AssessChanges(context,add,rm,cumulative_message,episode_len,edit_balance/sum_burst_bytes,burst_duration)
				Extend(context,trust_level)
			}	

			if TT.VERBOSE {
				fmt.Println("CONTEXT tick",episode,"/",i,context)
			}

			context = make(map[string]int)
			TT.StampedPromiseContext_End(G, ctx,changelog[i].Date)

			if i+1 < len(changelog) {
				ctx = TT.StampedPromiseContext_Begin(G, TT.KeyName(subject,0), changelog[i+1].Date)
			}

			sum_burst_bytes = 0
			edit_balance = 0
			cumulative_message = ""
			episode_len = 0
			add = ""
			rm = ""
			ARTICLE_ISSUES = 0
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

func AssessChanges(context map[string]int,add,rm,message string, eplen int, align,duration float64) string {

	add_len := len(add)
	rm_len := len(rm)

	//  Learn these and compare probability, we need a fixed vector
	
	// TT.ASSESS_EXCELLENT_S = "trust_high" 1.0
	// TT.ASSESS_PAR_S = "trust_ok"   0.5
	// TT.ASSESS_WEAK_S = "trust_low" 0.25
	// TT.ASSESS_SUBPART_S = "untrusted" 0


// The number of contributions to the episode-....

//	edits x rounds = flow rate

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

	// Pick some arbitrary signals

	var bad_signals = []string{"fuck","cunt","bastard","unhelpful","unfair","too","deceiv","deceptive","terrorism","justice","unsourced"}
	var bad_flag = false
	var sign string

	for s := range bad_signals {
		if strings.Contains(strings.ToLower(message),bad_signals[s]) {
			bad_flag = true
			sign = bad_signals[s]
		}
	}

	if bad_flag {
		fmt.Printf("\n Intentional heuristic in messaging -- (%s)\n",sign)
		Extend(context,"counter_policy_message")
	}

	intent := TT.StaticIntent(G,message)

	// How do we understand this threshold? We need to compare it to something similar, meta not subject

	AVERAGE_INTENT = 0.5 * AVERAGE_INTENT + 0.5 * intent

	sig := math.Sqrt((intent-AVERAGE_INTENT)*(intent-AVERAGE_INTENT)/(AVERAGE_INTENT*AVERAGE_INTENT))

	//fmt.Println("*",eplen,"ticks, Issues=", ARTICLE_ISSUES," WORK=(",balance,"/",total,")= align(",balance/total,")","intent variance=",intent,anomaly)

	// Divide the alignment -1 < a < +1 of user intent into finite classes

	align_class := int(align/0.05)
	ALIGNMENT[align_class]++
	days := duration / DAY

	// Save data

	const filename = "../data/ML/alignment_intent_delta_t"
	data := fmt.Sprintf("%f %f %f\n",align,sig,days)
	TT.AppendStringToFile(filename,data)

	if sig > 1.5 {
		fmt.Printf("\n Anomalous message intent -- (%s) = %f\n\n",message,sig)
		Extend(context,"anomalous_message")
	}

	if assess < TT.ASSESS_PAR {

		if len(rm) > 100 {
			fmt.Printf("\n --> RM %.100s ...(len=%d)\n\n",rm,len(rm))
		}

		if strings.Contains(add,"http") {
			Extend(context,"url_warning")
		} else if len(add) > 100 {
			fmt.Printf("\n <-- ADD %.100s ...(len=%d)\n\n",add,len(add))
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



