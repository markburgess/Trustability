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
//      go run wikipedia_ml_query.go
//
// Generates graph data based on machine learning analysis of wiki corpus
// ***********************************************************

package main

import (
	"fmt"
	"math"
	"TT"
)

// ********************************************************************************

func main() {
	
	TT.InitializeSmartSpaceTime()

	var dbname string = "SST-ML"
	var dburl string = "http://localhost:8529"
	var user string = "root"
	var pwd string = "mark"

	g := TT.OpenAnalytics(dbname,dburl,user,pwd)

	interactions := TT.GetAllWeekMemory(g, "interactions") 
	contention :=  TT.GetAllWeekMemory(g, "contention") 
	episodes :=  TT.GetAllWeekMemory(g, "BeginEndLocks") 

	var q_av,q_var float64 = 0,0

	for t := range interactions {
		q_av = 0.7 * interactions[t] + 0.3 * q_av
		q_var = 0.7 * (interactions[t]-q_av)*(interactions[t]-q_av) + 0.3 * q_var
		s := fmt.Sprintf("%d %f %f\n",t,q_av,math.Sqrt(q_var))
		TT.AppendStringToFile("../data/ML/interactions.dat", s)
	}

	q_av = 0
	q_var = 0

	for t := range contention {
		q_av = 0.7 * contention[t] + 0.3 * q_av
		q_var = 0.7 * (contention[t]-q_av)*(contention[t]-q_av) + 0.3 * q_var
		s := fmt.Sprintf("%d %f %f\n",t,q_av,math.Sqrt(q_var))
		TT.AppendStringToFile("../data/ML/contention.dat", s)
	}

	q_av = 0
	q_var = 0

	for t := range episodes {
		q_av = 0.7 * episodes[t] + 0.3 * q_av
		q_var = 0.7 * (episodes[t]-q_av)*(episodes[t]-q_av) + 0.3 * q_var
		s := fmt.Sprintf("%d %f %f\n",t,q_av,math.Sqrt(q_var))
		TT.AppendStringToFile("../data/ML/episodes.dat", s)
	}
}

