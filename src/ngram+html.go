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
//     go run ngram+html.go "https://jimruttshow.blubrry.net/the-jim-rutt-show-transcripts/transcript-of-ep-190-peter-turchin-on-cliodynamics-and-end-times/" 40 | more

package main

import (
	"strings"
	"os"
	"flag"
	"fmt"
	"golang.org/x/net/html"
	"io"
	"regexp"
	"net/http"
	"strconv"
	"TT"
)

// ****************************************************************************

// This version of the ngram.go moves reusable parts to the TT library package

// ****************************************************************************

var G TT.Analytics

// ****************************************************************************
// SCAN themed stories as text to understand their components
//
//   go run scan_stream.go ~/LapTop/SST/test3.dat 
//
// We want to input streams of narrative and extract phrase fragments to see
// which become statistically significant - maybe forming a hierarchy of significance.
// Try to measure some metrics/disrtibutions as a function of "amount", where
// amount is measured in characters, words, sentences, paragraphs, since these
// have different semantics.
// ****************************************************************************

func main() {

	flag.Usage = usage
	flag.Parse()
	args := flag.Args()
	
	if len(args) < 2 {
		usage()
		os.Exit(1);
	}
	
	level, err := strconv.Atoi(args[1])
	
	if err != nil {
		fmt.Println("The trust threshold should be between 0 and 100 percent")
		os.Exit(1);
	}
		
	threshold := float64(level)/100
	
	if threshold > 1 || threshold < 0 {

		fmt.Println("The scanning threshold should be between 0 and 100 percent")
		os.Exit(1);

	} else {
		TT.SetTrustThreshold(threshold)

		fmt.Println("******************************************************************")
		fmt.Println("** SEMANTIC TEXT SAMPLER, SST basis model")
		fmt.Println("** Sampling trust threshold = ",threshold*100,"/ 100")
		fmt.Println("******************************************************************")
	}

	// ***********************************************************

	TT.InitializeSmartSpaceTime()

	var dbname string = "SemanticSpacetime"
	var dburl string = "http://localhost:8529"
	var user string = "root"
	var pwd string = "mark"

	// ***********************************************************

	G = TT.OpenAnalytics(dbname,dburl,user,pwd)

	url := args[0]

	raw_text, err := http.Get(url)

	defer raw_text.Body.Close()

	proto_text := CleanHTML(raw_text)

	selected_sentences := TT.FractionateSentences(proto_text)

	TT.ReviewAndSelectEvents("url",selected_sentences)		
	
	topics := TT.RankByIntent(selected_sentences)
	
	TT.LongitudinalPersistentConcepts(topics)

}

//**************************************************************

func usage() {
	
	fmt.Fprintf(os.Stderr, "usage: go run scan_text.go [file].dat [1-100]\n")
	flag.PrintDefaults()
	os.Exit(2)
}

//**************************************************************

func CleanHTML(stream *http.Response) string {

	var plaintext string = ""

	var capture bool = false

	tokenizer := html.NewTokenizer(stream.Body)

	for {
		tokenType := tokenizer.Next()
		token := tokenizer.Token()

		if tokenizer.Err() == io.EOF {
			return ""
		}

		s := strings.TrimSpace(html.UnescapeString(token.String()))

		//fmt.Printf("x %T", tokenType)

		if len(s) == 0 {
			continue
		}
		
		switch tokenType {
			
		case html.ErrorToken:
			
			fmt.Printf("Error: %v", tokenizer.Err())
			return ""
			
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

				if s == "citation needed" {
					//fmt.Println("MISSING CITATION trustworthiness negative??")
					continue
				}

				if s == "edit" {
					continue
				}

				// Strip commas etc for the n-grams

				m := regexp.MustCompile("[?!.]+")
				marked := m.ReplaceAllString(s,"$0#")

				plaintext = plaintext + " " + marked
			}

		case html.StartTagToken:

			//fmt.Printf("Tag: %v\n", s)

			isAnchor := token.Data == "a"

			if isAnchor {
				//fmt.Println("anchor!",s)
			}

			if token.Data == "script" {
				//fmt.Println("IMPOSITION!",s)
			}

			if token.Data == "p" {
				capture = true
			}

		case html.EndTagToken:

			if token.Data == "body" {
				capture = false
				//fmt.Println("SUMMARY:",plaintext)
				return plaintext
			}
		}
		
	}

	return plaintext
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

