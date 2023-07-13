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
//

package main

import (
	"fmt"
	"golang.org/x/net/html"
	"net/http"
	"sort"
	"strings"
	"time"
	"io"
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

func main() {

	//url := "https://en.wikipedia.org/w/index.php?title=Jan_Bergstra&action=history"

	url := "https://en.wikipedia.org/w/index.php?title=Michael_Jackson&action=history&offset=&limit=1000"

	//url := "https://en.wikipedia.org/w/index.php?title=Mark_Burgess_(computer_scientist)&action=history&offset=&limit=500"
	changelog := MainPage(url)

	Assessment(changelog)
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
	var last string = ""
	var history int = 0

	// By entry state

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

		s := strings.TrimSpace(html.UnescapeString(token.String()))

		for i := range token.Attr {

			if token.Attr[i].Val == "mw-contributions-list" {				
				history++
			}

			if token.Attr[i].Val == "mw-changeslist-date" {				
				date = true
			}
		}

		if len(s) == 0 {
			continue
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
			}

			if attend && (strings.HasPrefix(s,"undo") ||strings.HasPrefix(s,"cur")||strings.HasPrefix(s,"<img")) {
				attend = false
				entry.Message = message
				changelog = append(changelog,entry)
			}

			if attend && s == "talk" {
				entry.User = last
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
					}
				}
			}

			if attend && after_edits {

				if strings.Contains(s,"Revert") {
					entry.Revert++
				}

				message += s + " "
			}

			last = s

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

	var users_idemp = make(map[string]int)
	var users_revert = make(map[string]int)
	var users []string

	sort.Slice(changelog, func(i, j int) bool {
		return changelog[i].Date.Before(changelog[j].Date)
	})
	
	for i := range changelog {

		users_idemp[changelog[i].User]++

		if changelog[i].Revert > 0 {
			users_revert[changelog[i].User] += changelog[i].Revert
		}
	}

	fmt.Println("Users", len(users_idemp))

	for s := range users_idemp {
		users = append(users,s)
	}

	sort.Slice(users, func(i, j int) bool {
		return users_idemp[users[i]] > users_idemp[users[j]]
	})

	fmt.Println("Ranked user changes (histo)")

	for s := range users {
		fmt.Println(" >",users[s],users_idemp[users[s]])
	}

	fmt.Println("Reversions (histo)")

	users = nil

	for s := range users_revert {
		users = append(users,s)
	}

	sort.Slice(users, func(i, j int) bool {
		return users_revert[users[i]] > users_revert[users[j]]
	})

	for s := range users {
		fmt.Println(" R",users[s],users_revert[users[s]])
	}

	// Time intervals between user changes (specific user and any user) -- Apply ML

}

