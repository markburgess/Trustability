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
	"strconv"
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

// Promise bindings in English. This domain knowledge saves us a lot of training analysis

var FORBIDDEN_ENDING = []string{"but", "and", "the", "or", "a", "an", "its", "it's", "their", "your", "my", "of", "as", "are", "is", "with", "using", "that", "who", "to" ,"no", "because","at","but"}

var FORBIDDEN_STARTER = []string{"and","or","of","the","it","because","in","that","these","those","is","are","was","were","but"}

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

var MISTRUST_THRESHOLD float64 = 0.8
const DETAIL_PER_LEG_POLICY = 3

// ***************************************************************************

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
var TOPICS = make(map[string]float64)
var EXCLUSIONS []string

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
	
	if len(args) < 2 {
		usage()
		os.Exit(1);
	}

	for i := 1; i < MAXCLUSTERS; i++ {

		STM_NGRAM_RANK[i] = make(map[string]float64)
		LTM_EVERY_NGRAM_OCCURRENCE[i] = make(map[string][]int)
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

		MISTRUST_THRESHOLD = threshold
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

	if strings.HasSuffix(args[0],".dat") {

		ReadSentenceStream(args[0])
		ReviewAndSelectEvents(args[0])		
		RankByIntent()
	}

	LongitudinalPersistentConcepts()
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

	m2 := regexp.MustCompile("\\\\[^ –\n]+") 
	stripped2 := m2.ReplaceAllString(stripped1," ") 

	// Non-English alphabet (tricky), but leave ?!:;

	m3 := regexp.MustCompile("[–{&}““”#%^+_#”=$’~‘/<>\"&]*") 
	stripped3 := m3.ReplaceAllString(stripped2,"") 

	m4 := regexp.MustCompile("[:;]+")
	stripped4 := m4.ReplaceAllString(stripped3,".")

	m5 := regexp.MustCompile("([^.,: ][\n])+")
	stripped5 := m5.ReplaceAllString(stripped4,"$0:")

	m6 := regexp.MustCompile("[^- a-zA-ZåøæÅØÆ.:,()!?\n]*")
	stripped6 := m6.ReplaceAllString(stripped5,"")

	m7 := regexp.MustCompile("[?!.]+")
	mark := m7.ReplaceAllString(stripped6,"$0#")

	m8 := regexp.MustCompile("[ \n]+")
	cleaned := m8.ReplaceAllString(mark," ")

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

		meaning[s_idx] = FractionateThenRankSentence(s_idx,sentences[s_idx],len(sentences))
	}

	// Some things to note: importance tends to be clustered around the start and the end of
	// a story. The start is automatically weakner in this method, due to lack of data. We can
	// compensate by weighting the start and the end by sentence number.

	midway := len(sentences) / 2

	for s_idx := range sentences {

		scale_factor := 1.0 + float64((midway - s_idx) * (midway - s_idx)) / float64(midway*midway)

		n := NarrationMarker(sentences[s_idx], meaning[s_idx] * scale_factor, s_idx)
			
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

func FractionateThenRankSentence(s_idx int, sentence string, total_sentences int) float64 {

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
			
			rank, rrbuffer = NextWordAndUpdateLTMNgrams(s_idx,cleanword, rrbuffer,total_sentences)
			sentence_meaning_rank += rank
		}
	}

	return sentence_meaning_rank
}

//**************************************************************

func RankByIntent() {

	sentences := len(SELECTED_SENTENCES)

	fmt.Println("--------- Sumarize ngram Intentionality threshold selection ---------------------------")

	for n := 1; n < MAXCLUSTERS; n++ {

		var last,delta int

		// Search through all sentence ngrams and measure distance between repeated
		// try to indentify any scales that emerge

		for ngram := range LTM_EVERY_NGRAM_OCCURRENCE[n] {

			occurrences := len(LTM_EVERY_NGRAM_OCCURRENCE[n][ngram])

			intent := Intentionality(n,ngram,sentences)

			if intent < 0.3  {
				continue
			}

			fmt.Println(n,ngram,occurrences,STM_NGRAM_RANK[n][ngram],"---------",intent)

			// if ngram of occurrences exceeds an expectation threshold in terms of length

			last = 0

			var min_delta int = 9999
			var max_delta int = 0
			var sum_delta int = 0

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

				sum_delta += delta

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
			
			if min_delta == 0 {
				continue
			}

			av_delta := float64(sum_delta)/float64(occurrences)

			if (av_delta > 3) && (av_delta < float64(LEG_WINDOW) * 4) {

				TOPICS[ngram] = intent
			}
		}
	}
}

// *****************************************************************

func LongitudinalPersistentConcepts() {
	
	fmt.Println("----- Emergent Longitudinally Stable Concept Fragments ---------")
	
	var sortable []Score
	
	for ngram := range TOPICS {
		
		var item Score
		item.Key = ngram
		item.Score = TOPICS[ngram]
		sortable = append(sortable,item)
	}
	
	sort.Slice(sortable, func(i, j int) bool {
		return sortable[i].Score < sortable[j].Score
	})
	
	// The score is the average interval between repetitions
	// If this is very long, the focus is spurious, so we look at the
	// shortest sample
	
	for i := 0; i < len(sortable); i++ {
		
		fmt.Printf("Particular theme/topic \"%s = %f\"\n", sortable[i].Key, sortable[i].Score)
	}
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

	// Go through all the sentences that haven't been excluded and pick a sampling density that's
	// approximately evenly distributed-- split into LEG_WINDOW intervals

	for s := range SELECTED_SENTENCES {

		sentence_id_by_rank[leg][SELECTED_SENTENCES[s].rank] = s

		if steps > LEG_WINDOW {

			this_leg_av_rank = av_rank_for_leg[leg]

			// At the start of a long doc, there's insufficient weight to make an impact, so
			// we need to compensate

			const ramp_up = 60
			
			if (leg < ramp_up) {
				this_leg_av_rank *= float64(LEG_WINDOW/ramp_up)
			}

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

	// Summarize	

	fmt.Println("------------------------------------------")
	fmt.Println("Notable events = ",KEPT,"of total ",ALL_SENTENCE_INDEX,"efficiency = ",100*float64(ALL_SENTENCE_INDEX)/float64(KEPT),"%")
	fmt.Println("------------------------------------------")
}

//**************************************************************
// TOOLKITS
//**************************************************************

func Intentionality(n int, s string, sentence_count int) float64 {

	occurrences := STM_NGRAM_RANK[n][s]
	work := float64(len(s))
	legs := float64(sentence_count) / float64(LEG_WINDOW)

	if occurrences < 3 {
		return 0
	}

	if work < 5 {
		return 0
	}

	// lambda should have a cutoff for insignificant words, like "a" , "of", etc that occur most often

	lambda := occurrences / float64(LEG_WINDOW)

	// This constant is tuned to give words a growing importance up to a limit
	// or peak occurrences, then downgrade

	// Things that are repeated too often are not important
	// but length indicates purposeful intent

	meaning := lambda * work / (1.0 + math.Exp(lambda-legs))

return meaning
}

//**************************************************************

func AnnotateLeg(filename string, leg int, sentence_id_by_rank map[float64]int, this_leg_av_rank, max float64) {

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

	fmt.Println(" >> (Rank leg untrustworthiness (anomalous interest)",leg,"=",scale_free_trust,")")

	if scale_free_trust > MISTRUST_THRESHOLD {

		var start int

		// top intra_leg_sampling_density = count backwards from the end

		if samples_per_leg > DETAIL_PER_LEG_POLICY {

			start = len(sentence_ranks) - DETAIL_PER_LEG_POLICY

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

	// Now highest importance within the lef, in order of occurrence

	for r := range ranks_in_order {

		fmt.Printf("\nEVENT[Leg %d selects %d]: %s\n",leg,ranks_in_order[r],SELECTED_SENTENCES[ranks_in_order[r]].text)
		KEPT++
	}
}

//**************************************************************

func NextWordAndUpdateLTMNgrams(s_idx int, word string, rrbuffer [MAXCLUSTERS][]string,total_sentences int) (float64,[MAXCLUSTERS][]string) {

	// Word by word, we form a superposition of scores from n-grams of different lengths
	// as a simple sum. This means lower lengths will dominate as there are more of them
	// so we define intentionality proportional to the length also as compensation

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

			STM_NGRAM_RANK[n][key]++
			rank += Intentionality(n,key,total_sentences)

			LTM_EVERY_NGRAM_OCCURRENCE[n][key] = append(LTM_EVERY_NGRAM_OCCURRENCE[n][key],s_idx)

		}
	}

	STM_NGRAM_RANK[1][word]++
	rank += Intentionality(1,word,total_sentences)

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

	for s := range FORBIDDEN_ENDING {
		if lastword == FORBIDDEN_ENDING[s] {
			return true
		}
	}

	for s := range FORBIDDEN_STARTER {
		if firstword == FORBIDDEN_STARTER[s] {
			return true
		}
	}

return false 
}


//**************************************************************

func usage() {
	
	fmt.Fprintf(os.Stderr, "usage: go run scan_text.go [file].dat [1-100]\n")
	flag.PrintDefaults()
	os.Exit(2)
}
