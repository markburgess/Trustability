//
// Calculate the predicted fission hierarchy for various beta
//

package main

import (
	"fmt"
	//"math"
)

func main() {

	Calc(1.0)
	Calc(0.93)
}

// ******************************************************************

func Calc(beta float64) {

	var N float64

	fmt.Printf("%10s %10s  (beta = %f)\n","n_max","<N>",beta)

	for n := 5.0; n <= 600; n = N {

		N = (n-1)*4*beta

		fmt.Printf("%10f %10f\n",n,N)

	}
}