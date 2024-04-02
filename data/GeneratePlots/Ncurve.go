package main

import (
    "bufio"
    "fmt"
    "log"
    "os"
)

func main() {

	Calc("../Wikipedia/trust.dat")
}

// **********************************************************

func Calc(filename string) {

	file, err := os.Open(filename)
	
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	
	scanner := bufio.NewScanner(file)

	var I[100] float64
	
	for scanner.Scan() {

		var n,x,i float64 = 0,0,0

		s := scanner.Text()
		fmt.Sscanf(s,"%f %f %f ",&x,&x,&n,&x,&x,&x,&i)

		N := int(n+0.5)
		I[N]++

	}

	for N := 1; N < 15; N++ {
		fmt.Printf("%d %f\n",N,600*I[N])
	}

}
