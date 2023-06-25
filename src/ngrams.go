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
	"io/ioutil"
	"flag"
	"fmt"
	"regexp"
	"sort"
	"math"
	"TT"
)

// ****************************************************************************

// Invariants - looks for interferometry of fragments -- persistent sequences
// over consecutive legs. This helps to stabilize conceptual fragments - more
// certain if they are repeated.

// In this expt we reduced the threshold for meaning so collecting more events
// higher density

// ****************************************************************************

// Short term memory class

type Narrative struct {

	rank float64
	text string
	index int
}

type Score struct {

	Key   string
	Score float64
}

// ***************************************************************************

var WORDCOUNT int = 0
var LEGCOUNT int = 0
var KEPT int = 0
var SKIPPED int = 0
var ALL_SENTENCE_INDEX int = 0

var SELECTED_SENTENCES []Narrative

var THRESH_ACCEPT float64 = 0
var TOTAL_THRESH float64 = 0

// ************** SOME INTRINSIC SPACETIME SCALES ****************************

const MAXCLUSTERS = 7
const LEG_WINDOW = 100

var ATTENTION_LEVEL float64 = 1.0
var SENTENCE_THRESH float64 = 100 // chars

const REPEATED_HERE_AND_NOW  = 1.0 // initial primer
const INITIAL_VALUE = 0.5

const MEANING_THRESH = 20      // reduce this if too few samples
const FORGET_FRACTION = 0.001  // this amount per sentence ~ forget over 1000 words

// ****************************************************************************

var LTM_EVERY_NGRAM_OCCURRENCE [MAXCLUSTERS]map[string][]int
var INTENTIONALITY [MAXCLUSTERS]map[string]float64
var SCALE_PERSISTENTS [MAXCLUSTERS]map[string]float64

var STM_NGRAM_RANK [MAXCLUSTERS]map[string]float64

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
	
	if len(args) < 1 {
		fmt.Println("file list expected")
		os.Exit(1);
	}

	for i := 1; i < MAXCLUSTERS; i++ {

		STM_NGRAM_RANK[i] = make(map[string]float64)
		INTENTIONALITY[i] = make(map[string]float64)
		SCALE_PERSISTENTS[i] = make(map[string]float64)
		LTM_EVERY_NGRAM_OCCURRENCE[i] = make(map[string][]int)
	} 
	
	// ***********************************************************

	TT.InitializeSmartSpaceTime()

	var dbname string = "SemanticSpacetime"
	var url string = "http://localhost:8529"
	var user string = "root"
	var pwd string = "mark"

	// ***********************************************************

	G = TT.OpenAnalytics(dbname,url,user,pwd)

	for i := range args {

		if strings.HasSuffix(args[i],".dat") {

			ReadSentenceStream(args[i])  // Once for whole thing, reset and compare to realtime

			ReviewAndSelectEvents(args[i])


			//SummarizeHistograms(G)

		}
	}

	fmt.Println("\nKept = ",KEPT,"of total ",ALL_SENTENCE_INDEX,"efficiency = ",100*float64(ALL_SENTENCE_INDEX)/float64(KEPT),"%")

}

//**************************************************************
// Scan text input
//**************************************************************

func ReadSentenceStream(filename string) {

	// The take the filename as a marker for the semantic map
	// as an arbitrary starting concept marker

	ReadAndCleanRawStream(filename)
}

//**************************************************************

func ReadAndCleanRawStream(filename string) {

	// Here we can provide different readers for different formats

	proto_text := CleanFile(string(filename))
	
	FractionateSentences(proto_text)
}

//**************************************************************

func CleanFile(filename string) string {

	content, _ := ioutil.ReadFile(filename)

	// Start by stripping HTML / XML tags before para-split
	// if they haven't been removed already

	m1 := regexp.MustCompile("<[^>]*>") 
	stripped1 := m1.ReplaceAllString(string(content),"") 

	//Strip and \begin \latex type commands

	m2 := regexp.MustCompile("\\\\[^ \n]+") 
	stripped2 := m2.ReplaceAllString(stripped1," ") 

	// Non-English alphabet (tricky), but leave ?!:;

	m3 := regexp.MustCompile("[–{&}“#%^+_#”=$’~‘/()<>\"&]*") 
	stripped3 := m3.ReplaceAllString(stripped2,"") 

	// Strip digits, this is probably wrong in general
	m4 := regexp.MustCompile("[:;]+")
	stripped4 := m4.ReplaceAllString(stripped3,".")

	m5 := regexp.MustCompile("[^ a-zA-Z.,!?\n]*")
	stripped5 := m5.ReplaceAllString(stripped4,"")

	m6 := regexp.MustCompile("[?!.]+")
	mark := m6.ReplaceAllString(stripped5,"$0#")

	m7 := regexp.MustCompile("[ \n]+")
	cleaned := m7.ReplaceAllString(mark," ")

	return cleaned
}

//**************************************************************

func FractionateSentences(text string) {

	var sentences []string

	// Coordinatize the non-trivial sentences in terms of their ngrams

	if len(text) == 0 {
		return
	}

	sentences = SplitIntoSentences(text)

	var meaning = make([]float64,len(sentences))

	for s_idx := range sentences {

		meaning[s_idx] = FractionateThenRankSentence(ALL_SENTENCE_INDEX,sentences[s_idx])

	}

	meaning = SearchInvariantsAndUpdateImportance(meaning)

	for s_idx := range sentences {

		n := NarrationMarker(sentences[s_idx], meaning[s_idx], ALL_SENTENCE_INDEX)
			
		SELECTED_SENTENCES = append(SELECTED_SENTENCES,n)
		
		ALL_SENTENCE_INDEX++
	}
}

//**************************************************************

func SplitIntoSentences(text string) []string {

	// Note this regex split has the effect of removing .?!

	re := regexp.MustCompile("#")
	sentences := re.Split(text, -1)

	var cleaned  = make([]string,1)

	for sen := range sentences{

		content := strings.Trim(sentences[sen]," ")

		if len(content) > 0 {			
			cleaned = append(cleaned,content)
		}
	}

	return cleaned
}

//**************************************************************

func FractionateThenRankSentence(s_idx int, sentence string) float64 {

	var rrbuffer [MAXCLUSTERS][]string
	var sentence_meaning_rank float64 = 0
	var rank float64

	// split on any punctuation here, because punctuation cannot be in the middle
	// of an n-gram by definition of punctuation's promises
	// THIS IS A PT +/- constraint
	
	re := regexp.MustCompile("[,.;:!?]")
	sentence_frags := re.Split(sentence, -1)
	
	for f := range sentence_frags {
		
		// For one sentence, break it up into codons and sum their importances
		
		clean_sentence := strings.Split(string(sentence_frags[f])," ")
		
		for word := range clean_sentence {
			
			// This will be too strong in general - ligatures and foreign languages etc
			
			m := regexp.MustCompile("[/()?!]*") 
			cleanjunk := m.ReplaceAllString(clean_sentence[word],"") 
			cleanword := strings.Trim(strings.ToLower(cleanjunk)," ")
			
			if len(cleanword) == 0 {
				continue
			}
			
			// Shift all the rolling longitudinal Ngram rr-buffers by one word
			
			rank, rrbuffer = NextWordAndUpdateLTMNgrams(s_idx,cleanword, rrbuffer)
			sentence_meaning_rank += rank
		}
	}
	
	return sentence_meaning_rank
}

//**************************************************************

func SummarizeHistograms(g TT.Analytics) {

	fmt.Println("----- (Intentionality scores) ----------")

	for n := 1; n < MAXCLUSTERS; n++ {

		var max float64 = 0

		for ngram := range STM_NGRAM_RANK[n] {

			if STM_NGRAM_RANK[n][ngram] > max {
				max = STM_NGRAM_RANK[n][ngram]
			}
		}

		fmt.Printf("Max value STM_NGRAM_RANK[%d] --> %f\n",n,max)
	}

	for n := 1; n < MAXCLUSTERS; n++ {

		var sortable []Score

		for ngram := range INTENTIONALITY[n] {

			var item Score
			item.Key = ngram
			item.Score = INTENTIONALITY[n][ngram]
			sortable = append(sortable,item)
		}

		sort.Slice(sortable, func(i, j int) bool {
			return sortable[i].Score < sortable[j].Score
		})
		
		for i := 3 *len(sortable) / 4; i < len(sortable); i++ {
			if n > 2 {
				fmt.Printf("Intention score (%d-gram) %s %f (log scaled)\n",n,sortable[i].Key,sortable[i].Score)
			}
		}

	}

	fmt.Println("--------------------------------")
}

//**************************************************************

func SearchInvariantsAndUpdateImportance(meaning []float64) []float64 {

	var thresh_count [MAXCLUSTERS]map[int][]string

	for n := 1; n < MAXCLUSTERS; n++ {

		fmt.Println("----- LONGITUDINAL INVARIANTS", n)

		var last,delta int
		thresh_count[n] = make(map[int][]string)

		// Search through all sentence ngrams and measure distance between repeated
		// try to indentify any scales that emerge

		for ngram := range LTM_EVERY_NGRAM_OCCURRENCE[n] {

			if (InsignificantPadding(ngram)) {
				continue
			}

			// **** LONG

			occurrences := len(LTM_EVERY_NGRAM_OCCURRENCE[n][ngram])

			// if ngram of occurrences exceeds an expectation threshold in terms of length

			last = 0

			var min_delta int = 9999
			var max_delta int = 0

			for location := 0; location < occurrences; location++ {

				// Foreach occurrence, check proximity to others
				// This is about seeing if an ngram is a recurring input in the stream.
				// Does the subject recur several times over some scale? The scale may be
				// logarithmic like n / log (o1-o2) for occurrence separation
				// Radius = 100 sentences, how many occurrences of this ngram close together?
				
				// Does meaning have an intrinsic radius? It doesn't make sense that it
				// depends on the length of the document. How could we measure this?	
				
				// two one relative to first occurrence (absolute range), one to last occurrence??
				// only the last is invariant on the scale of a story
				
				delta = LTM_EVERY_NGRAM_OCCURRENCE[n][ngram][location] - last			
				last = LTM_EVERY_NGRAM_OCCURRENCE[n][ngram][location]

				if min_delta > delta {
					min_delta = delta
				}

				if max_delta < delta {
					max_delta = delta
				}
			}

			// which ngrams occur in bursty clusters. If completely even, then significance
			// is low or the theme of the whole piece. If cluster span/total span
			// max interdistance >> min interdistance then bursty

			const persistence_factor = 2.5  // measured in sentences

			if (min_delta < LEG_WINDOW/persistence_factor) && (max_delta > LEG_WINDOW*persistence_factor) {

				fmt.Printf("Longitudinal %d-invariant \"%s\" (%d) --  min %d, max %d\n",n,ngram,occurrences,min_delta,max_delta)
				// We keep these separate as we expect them to represent topics within the story

				SCALE_PERSISTENTS[n][ngram] = math.Log(float64(len(ngram)))
			}

			for ngram := range LTM_EVERY_NGRAM_OCCURRENCE[n] {
				
				if (InsignificantPadding(ngram)) {
					continue
				}
				
				occurrences := len(LTM_EVERY_NGRAM_OCCURRENCE[n][ngram])
				
				// if ngram of occurrences exceeds an expectation threshold in terms of length
				
				for location := 0; location < occurrences; location++ {
					
					// Now BOOST/update the relevant sentence scores where they appear
					
					meaning[location] += SCALE_PERSISTENTS[n][ngram]
				}
			}
		}
	}

	return meaning
}

// *****************************************************************

func ReviewAndSelectEvents(filename string) {

	// The importances have now all been measured in realtime, but we review them now...posthoc
	// Now go through the history map chronologically, by sentence only reset the narrative  
        // `leg' counter when it fills up to measure story progress. 
	// This determines the sampling density of "important sentences" - pick a few from each leg

	var steps,leg int

	// Sentences to summarize per leg of the story journey

	steps = 0

	// We rank a leg by summing its sentence ranks

	var rank_sum float64 = 0
	var av_rank_for_leg []float64
	
	// First, coarse grain the narrative into `legs', 
        // i.e. standardized "narrative regions" by meter not syntax

	for s := range SELECTED_SENTENCES {

		// Make list of summed importance ranks for each leg

		rank_sum += SELECTED_SENTENCES[s].rank

		// Once we've summed all the importances and reached the end of the leg
		// define the leg_rank_average as the average over the interval and add this
		// to a list/array indexed by legs sequentially (append)

		if steps > LEG_WINDOW {
			steps = 0
			leg_rank_average := rank_sum / float64(LEG_WINDOW)
			av_rank_for_leg = append(av_rank_for_leg,leg_rank_average)
			rank_sum = 0
		}

		steps++	
	}

	// Don't forget any final "short" leg if there is one (residuals from the loop < LEG_WINDOW)

	leg_rank_average := rank_sum / float64(steps)
	av_rank_for_leg = append(av_rank_for_leg,leg_rank_average)

	// Find the leg of maximum importance

	var max_all_legs float64 = 0

	for l := range av_rank_for_leg {

		if max_all_legs < av_rank_for_leg[l] {

			max_all_legs = av_rank_for_leg[l]
		}
	}

	// Select a sampling rate that's lazy (> 1 sentence per leg) or busy ( <a few)
	// for important legs

	steps = 0
	leg = 0
	var this_leg_av_rank float64 = av_rank_for_leg[0]

	var sentence_id_by_rank = make(map[int]map[float64]int)
	sentence_id_by_rank[0] = make(map[float64]int)

	// Go through all the sentences that haven't been excluded and pick a simpling density that's
	// approximately evenly distributed-- split into LEG_WINDOW intervals

	for s := range SELECTED_SENTENCES {

		sentence_id_by_rank[leg][SELECTED_SENTENCES[s].rank] = s

		if steps > LEG_WINDOW {

			this_leg_av_rank = av_rank_for_leg[leg]

			AnnotateLeg(filename, leg, sentence_id_by_rank[leg], this_leg_av_rank, max_all_legs)

			steps = 0
			leg++

			sentence_id_by_rank[leg] = make(map[float64]int)
		}

		steps++
	}

	// Don't forget the final remainder (catch leg++)

	this_leg_av_rank = av_rank_for_leg[leg]
	
	AnnotateLeg(filename, leg, sentence_id_by_rank[leg], this_leg_av_rank, max_all_legs)
}

//**************************************************************
// TOOLKITS
//**************************************************************

func Intentionality(n int, s string, sentence_count int) float64 {

	// Emotional bias to be added ?

	STM_NGRAM_RANK[n][s]++

	frequency := STM_NGRAM_RANK[n][s] / float64(1 + sentence_count)

	// Things that are repeated too often are not important
	// but length indicates purposeful intent
	// motiovation is multscale AND^n  -> compare

	meaning := math.Log(float64(len(s)) / frequency)

return meaning
}

//**************************************************************

func AnnotateLeg(filename string, leg int, sentence_id_by_rank map[float64]int, this_leg_av_rank, max float64) {

	const leg_trust_threshold = 0.8       // 80/20 rule -- CONTROL VARIABLE
	const intra_leg_sampling_density = 4  // base trust selection

	var sentence_ranks []float64
	var ranks_in_order []int

	key := make(map[float64]int)

	for fl := range sentence_id_by_rank {

		sentence_ranks = append(sentence_ranks,fl)
	}

	var samples_per_leg = len(sentence_ranks)

	if samples_per_leg < 1 {
		return
	}

	// Rank by importance and rescale all as dimensionless between [0,1]

	sort.Float64s(sentence_ranks)
	scale_free_trust := this_leg_av_rank / max

	// We now have an array of sentences whose indices are ascending ordered rankings, max = last
	// and an array of rankings min to max
	// Set up a key = sentence with rank = r as key[r]

	for i := range sentence_ranks {
		key[sentence_ranks[i]] = sentence_id_by_rank[sentence_ranks[i]]
	}

	// Select only the most important remaining in order for the hub
	// Hubs will overlap with each other, so some will be "near" others i.e. "approx" them
	// We want the degree of overlap between hubs TT.CompareContexts()

	fmt.Println(" >> (Rank leg trustworthiness",leg,"=",scale_free_trust,")")

	if scale_free_trust > leg_trust_threshold {

		var start int

		// top intra_leg_sampling_density = count backwards from the end

		if samples_per_leg > intra_leg_sampling_density {

			start = len(sentence_ranks) - intra_leg_sampling_density

		} else {

			start = 0
		}

		for i :=  start; i < len(sentence_ranks); i++ {

			r := key[sentence_ranks[i]]
			ranks_in_order = append(ranks_in_order,r)
		}

		// Put the ranked selections back in sentence order

		sort.Ints(ranks_in_order)

	}

	// Now highest importance in order of occurrence

	for r := range ranks_in_order {

		fmt.Printf("\nEVENT[Leg %d selects %d]: %s\n",leg,ranks_in_order[r],SELECTED_SENTENCES[ranks_in_order[r]].text)
		KEPT++
	}
}

//**************************************************************

func NextWordAndUpdateLTMNgrams(s_idx int, word string, rrbuffer [MAXCLUSTERS][]string) (float64,[MAXCLUSTERS][]string) {

	var rank float64 = 0

	for n := 2; n < MAXCLUSTERS; n++ {
		
		// Pop from round-robin

		if (len(rrbuffer[n]) > n-1) {
			rrbuffer[n] = rrbuffer[n][1:n]
		}
		
		// Push new to maintain length

		rrbuffer[n] = append(rrbuffer[n],word)

		// Assemble the key, only if complete cluster
		
		if (len(rrbuffer[n]) > n-1) {
			
			var key string
			
			for j := 0; j < n; j++ {
				key = key + rrbuffer[n][j]
				if j < n-1 {
					key = key + " "
				}
			}

			if ExcludedByBindings(rrbuffer[n][0],rrbuffer[n][n-1]) {

				continue
			}

			INTENTIONALITY[n][key] = 0.5 * Intentionality(n,key,s_idx) + 0.5 * INTENTIONALITY[n][key]
			rank += INTENTIONALITY[n][key]

			LTM_EVERY_NGRAM_OCCURRENCE[n][key] = append(LTM_EVERY_NGRAM_OCCURRENCE[n][key],s_idx)

		}
	}

	INTENTIONALITY[1][word] = 0.5 * Intentionality(1,word,s_idx) + 0.5 * INTENTIONALITY[1][word]
	rank += INTENTIONALITY[1][word]

	LTM_EVERY_NGRAM_OCCURRENCE[1][word] = append(LTM_EVERY_NGRAM_OCCURRENCE[1][word],s_idx)

	return rank, rrbuffer
}

//**************************************************************
// MISC
//**************************************************************

func NarrationMarker(text string, rank float64, index int) Narrative {

	var n Narrative

	n.text = text
	n.rank = rank
	n.index = index

return n
}

//**************************************************************

func ExcludedByBindings(firstword,lastword string) bool {

	// A standalone fragment can't start/end with these words, because they
	// Promise to bind to something else...
	// Rather than looking for semantics, look at spacetime promises only - words that bind strongly
	// to a prior or posterior word.

	if (len(firstword) == 1) || len(lastword) == 1 {
		return true
	}

	var eforbidden = []string{"but", "and", "the", "or", "a", "an", "its", "it's", "their", "your", "my", "of", "as", "are", "is" }

	for s := range eforbidden {
		if lastword == eforbidden[s] {
			return true
		}
	}

	var sforbidden = []string{"and","or","of"}

	for s := range sforbidden {
		if firstword == sforbidden[s] {
			return true
		}
	}

return false 
}

// *****************************************************************

func InsignificantPadding(word string) bool {

	// This is a shorthand for the most common words and phrases, which may be learned by scanning many docs
	// Earlier, we learned these too, now just cache them

	if len(word) < 3 {
		return true
	}

	var irrel = []string{"hub:", "but", "and", "the", "or", "a", "an", "its", "it's", "their", "your", "my", "of", "if", "we", "you", "i", "there", "as", "in", "then", "that", "with", "to", "is","was", "when", "where", "are", "some", "can", "also", "it", "at", "out", "like", "they", "her", "him", "them", "his", "our", "by", "more", "less", "from", "over", "under", "why", "because", "what", "every", "some", "about", "though", "for", "around", "about", "any", "will","had","all","which" }

	for s := range irrel {
		if irrel[s] == word {
			return true
		}
	}

return false
}

//**************************************************************

func usage() {
    fmt.Fprintf(os.Stderr, "usage: go run scan_text.go [filelist]\n")
    flag.PrintDefaults()
    os.Exit(2)
}
