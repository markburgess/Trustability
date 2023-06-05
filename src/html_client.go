
package main

import (
	"fmt"
	"golang.org/x/net/html"
	"io"
	"net/http"
	"strings"
)

// ***********************************************************

var CAPTURE bool = false

// ***********************************************************

func main() {

	response, err := http.Get("https://en.wikipedia.org/wiki/Mark_Burgess_(computer_scientist)")

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
			return
		}

		s := strings.TrimSpace(html.UnescapeString(token.String()))

		//fmt.Printf("x %T", tokenType)

		if len(s) == 0 {
			continue
		}
		
		switch tokenType {
			
		case html.ErrorToken:
			
			fmt.Printf("Error: %v", tokenizer.Err())
			return
			
		case html.TextToken:

			if CAPTURE {

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

				if s == "citation needed" {
					fmt.Println("MISSING CITATION trustworthiness negative??")
					continue
				}

				if s == "edit" {
					continue
				}


				// Strip commas etc for the n-grams

				plaintext = plaintext + " " + s + "\n"
			}

		case html.StartTagToken:

			//fmt.Printf("Tag: %v\n", s)

			isAnchor := token.Data == "a"

			if isAnchor {
				//fmt.Println("anchor!",s)
			}

			if token.Data == "script" {
				fmt.Println("IMPOSITION!",s)
			}

			if token.Data == "h1" {
				CAPTURE = true
			}

		case html.EndTagToken:

			if token.Data == "body" {
				CAPTURE = false
				fmt.Println("SUMMARY:",plaintext)
			}

			//fmt.Println("END OF",s,token.Data)
		}
		
	}
}

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
