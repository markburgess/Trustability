
//
// Calculate the average values of <N> appearing in the formula for the data sets
//

package main

import (
    "bufio"
    "fmt"
    "log"
    "os"
)

// ************************************************************

func main() {

	r1 := Calc("../Wikipedia/trust.dat","all results")
	r2 := Calc("../WikipediaNoBots/trust.dat","without bots")

	fmt.Println("Ratio nobots/bots",r2/r1)
}

// ************************************************************

func Calc(filename,legend string) float64 {

	file, err := os.Open(filename)
	
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	
	scanner := bufio.NewScanner(file)
	
	var a,b,count,n_tot float64 = 0,0,0,0

	for scanner.Scan() {

		var n float64 = 0
		s := scanner.Text()
		fmt.Sscanf(s,"%f %f %f ",&a,&b,&n)
		n_tot += n
		count++
	}

	result := n_tot/count

	fmt.Println("<N>",result,legend)

	return result
}
