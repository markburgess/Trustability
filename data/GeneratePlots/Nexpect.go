package main

import (
    "bufio"
    "fmt"
    "log"
    "os"
)

func main() {

	r1 := Calc("../Wikipedia/episodeclusters.dat","all")
	r2 := Calc("../WikipediaNoBots/episodeclusters.dat","no bots")

	fmt.Println("Ratio nbots/bots =",r2/r1)
}

// **********************************************************

func Calc(filename,legend string) float64 {

	file, err := os.Open(filename)
	
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	
	var n_expect float64 = 0

	scanner := bufio.NewScanner(file)
	
	for scanner.Scan() {

		var n,p float64 = 0,0

		s := scanner.Text()
		fmt.Sscanf(s,"%f %f %f ",&n,&p)
		n_expect += n*p
	}

	fmt.Println("N expected sum pn = ",n_expect,legend)

	return n_expect

}
