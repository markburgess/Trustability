
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
}

// ***********************************************************

func main() {

	url := "https://en.wikipedia.org/w/index.php?title=Jan_Bergstra&action=history"

	//url := "https://en.wikipedia.org/w/index.php?title=Mark_Burgess_(computer_scientist)&action=history&offset=&limit=500"
	MainPage(url)

}

// ***********************************************************

func MainPage(url string) {

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
	var plaintext string = ""

	// By entry state

	var entry WikiNote
	var changelog []WikiNote

	// Start parsing

	tokenizer := html.NewTokenizer(response.Body)

	for {
		tokenType := tokenizer.Next()
		token := tokenizer.Token()

		if tokenizer.Err() == io.EOF {
			return
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
			return
			
		case html.TextToken:

			if s == "prev" {

				attend = true
				var empty WikiNote
				entry = empty
				after_edits = false
				message = ""

				plaintext = plaintext + "----------------\n"
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

			if attend && strings.HasPrefix(s,"undo") {
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
				message += s + " "
			}

			last = s

		case html.StartTagToken:

		case html.EndTagToken:

			if token.Data == "body" {
				
				sort.Slice(changelog, func(i, j int) bool {
					return changelog[i].Date.Before(changelog[j].Date)
				})

				for i := range changelog {
					fmt.Println(changelog[i])
				}
			}
		}
	}
}

// *******************************************************************************

func Assessment(entry WikiNote) string {

	return ""
}

