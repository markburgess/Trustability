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
//     go run ngrams.go /home/mark/Laptop/Work/SST/data_samples/obama.dat

package main

import (
	"strings"
	"os"
	"flag"
	"fmt"
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
		fmt.Println("The trust threshold should be between 20 and 100 percent")
		os.Exit(1);
	}
		
	threshold := float64(level)/100
	
	if threshold > 1 || threshold < 0.2 {

		fmt.Println("The scanning threshold should be between 20 and 100 percent")
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
	var url string = "http://localhost:8529"
	var user string = "root"
	var pwd string = "mark"

	// ***********************************************************

	G = TT.OpenAnalytics(dbname,url,user,pwd)

	filename := args[0]

	if strings.HasSuffix(filename,".dat") {

		proto_text := TT.ReadAndCleanFile(filename)

		selected_sentences := TT.FractionateSentences(proto_text)

		TT.ReviewAndSelectEvents(filename,selected_sentences)		

		topics := TT.RankByIntent(selected_sentences)

		TT.LongitudinalPersistentConcepts(topics)
	}
}

//**************************************************************

func usage() {
	
	fmt.Fprintf(os.Stderr, "usage: go run scan_text.go [file].dat [1-100]\n")
	flag.PrintDefaults()
	os.Exit(2)
}
