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
// e.g. set url variable and
//      go run wikipedia_history.go
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
	"time"
	"io"
	"TT"
)

// ***********************************************************

type WikiNote struct {
	
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

var G TT.Analytics

// ***********************************************************

func main() {

	// Example pages, some familiar some notorious

	subject := "Mark Burgess"
	url := "https://en.wikipedia.org/w/index.php?title=Mark_Burgess_(computer_scientist)&action=history&offset=&limit=1000"

	//subject := "Jan Bergstra"
	//url := "https://en.wikipedia.org/w/index.php?title=Jan_Bergstra&action=history&offset=&limit=1000"

	//subject := "Michael Jackson"
	//url := "https://en.wikipedia.org/w/index.php?title=Michael_Jackson&action=history&offset=&limit=1000"

	// subject := "George W. Bush"
	//url := "https://en.wikipedia.org/w/index.php?title=George_W._Bush&action=history&offset=&limit=1000"

	// ***********************************************************

	TT.InitializeSmartSpaceTime()

	var dbname string = "SemanticSpacetime"
	var dburl string = "http://localhost:8529"
	var user string = "root"
	var pwd string = "mark"

	// ***********************************************************

	TT.LEG_WINDOW = 10 // Need a smaller window than normal for fragmented text

	G = TT.OpenAnalytics(dbname,dburl,user,pwd)

	TT.SetTrustThreshold(0.1)

	changelog := MainPage(url)

	sort.Slice(changelog, func(i, j int) bool {
		return changelog[i].Date.Before(changelog[j].Date)
	})

	Assessment(changelog)

	fulltext := TotalText(changelog)

	selected_sentences := TT.FractionateSentences(fulltext)
	
	TT.ReviewAndSelectEvents(subject,selected_sentences)		
	
	topics := TT.RankByIntent(selected_sentences)
	
	TT.LongitudinalPersistentConcepts(topics)

}

// ***********************************************************

func MainPage(url string) []WikiNote {

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

	var entry WikiNote
	var changelog []WikiNote

	// Start parsing

	tokenizer := html.NewTokenizer(response.Body)

	for {
		tokenType := tokenizer.Next()
		token := tokenizer.Token()

		if tokenizer.Err() == io.EOF {
			return changelog
		}

		// Strip out junk characters

		s := strings.TrimSpace(html.UnescapeString(token.String()))
		s = strings.ReplaceAll(s,"→","")
		s = strings.ReplaceAll(s,"←","")
		s = strings.ReplaceAll(s,"'","")
		s = strings.ReplaceAll(s,".","")
		s = strings.ReplaceAll(s,"{{","")
		s = strings.ReplaceAll(s,"}}","")
		s = strings.ReplaceAll(s,"(","")
		s = strings.ReplaceAll(s,")","")
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
				var empty WikiNote
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

				if attend && (strings.HasPrefix(s,"Tag") || strings.HasPrefix(s,"bot") || strings.HasPrefix(s,"New page:") ||strings.HasPrefix(s,"Category:") || strings.HasPrefix(s,"undo") || strings.HasPrefix(s,"cur")|| strings.HasPrefix(s,"<img") || strings.Contains(s,"|")) {
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

func Assessment(changelog []WikiNote) {

	var users_changecount = make(map[string]int)
	var users_revert = make(map[string]int)
	var users_lasttime = make(map[string]int64)
	var users_averagetime = make(map[string]float64)	
	var users_revert_dt = make(map[string]float64)
	var users []string
	
	for i := range changelog {

		//fmt.Printf(">> %15s (%v)(%d), %s\n", changelog[i].User,changelog[i].Date,changelog[i].EditDelta,changelog[i].Message)

		if users_lasttime[changelog[i].User] > 0 {
			delta := float64(changelog[i].Date.UnixNano() - users_lasttime[changelog[i].User])
			users_averagetime[changelog[i].User] = 0.4 * users_averagetime[changelog[i].User] + 0.6 * delta
		}

		users_lasttime[changelog[i].User] = changelog[i].Date.UnixNano()

		users_changecount[changelog[i].User]++

		if changelog[i].Revert > 0 && i > 1 {
			
			users_revert[changelog[i].User] += changelog[i].Revert
			dt := float64(changelog[i].Date.UnixNano() - changelog[i-1].Date.UnixNano())
			users_revert_dt[changelog[i].User] = 0.6 * dt + 0.4 * users_revert_dt[changelog[i].User]
		}
	}

	fmt.Println("Total users involved in shared process", len(users_changecount))

	for s := range users_changecount {
		users = append(users,s)
	}

	sort.Slice(users, func(i, j int) bool {
		return users_changecount[users[i]] > users_changecount[users[j]]
	})

	fmt.Println("Ranked user changes: number and average time interval")

	for s := range users {
		if users_changecount[users[s]] > 1 {
			fmt.Printf("  > %20s  (%2d)   av_delta %-3.2f (days)\n",users[s],users_changecount[users[s]],users_averagetime[users[s]]/DAY)
		} else {
			fmt.Print(users[s],", ")
		}
	}

	fmt.Println("\n\nReversions (agents exhibiting contentious behaviour)")

	users = nil

	for s := range users_revert {
		users = append(users,s)
	}

	sort.Slice(users, func(i, j int) bool {
		return users_revert[users[i]] > users_revert[users[j]]
	})

	for s := range users {
		fmt.Printf(" R  %20s (%d) of %d after average of %3.2f mins\n",users[s],users_revert[users[s]],users_changecount[users[s]],users_revert_dt[users[s]]/MINUTE)
	}
}

// *******************************************************************************

func TotalText(changelog []WikiNote) string {

	var text string = ""

	for i := range changelog {

		text += changelog[i].Message

	}

	return text
}