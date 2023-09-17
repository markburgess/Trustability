/*

// Plot the data fit graph 

go run AB.go > ABout.dat
gnuplot
set xlabel "n"
set ylabel "psi(N)"
set term pdf monochrome
unset key
set output "datafit2.pdf"
plot [0:40] "ABout.dat" using 1:2:3 w yerr, "ABout.dat" using 1:4 w line

*/

package main

import (
	"fmt"
	"math"
)

func main() {

	curve := []float64{ 
		0,
		0.039679,
		0.106834,
		0.109947,
		0.110932,
		0.093690,
		0.082964,
		0.073335,
		0.060549,
		0.049465,
		0.040217,
		0.035985,
		0.029894,
		0.024856,
		0.021474,
		0.018496,
		0.014398,
		0.012450,
		0.010883,
		0.010099,
		0.007412,
		0.006225,
		0.005956,
		0.004635,
		0.004210,
		0.003471,
		0.003336,
		0.002262,
		0.002172,
		0.001747,
		0.001635,
		0.001590,
		0.001411,
		0.001052,
	}

	curve2 := []float64{ 
		0.0,
		0.042644,
		0.114430,
		0.120455,
		0.112812,
		0.095440,
		0.081679,
		0.072489,
		0.056875,
		0.047896,
		0.040722,
		0.033150,
		0.029610,
		0.024030,
		0.019927,
		0.017044,
		0.014652,
		0.011511,
		0.010315,
		0.008205,
		0.006869,
		0.005931,
		0.005181,
		0.004173,
		0.003517,
		0.003141,
		0.002508,
		0.001899,
		0.001946,
		0.001266,
		0.001571,
		0.001196,
		0.000844,
	}

	//A := 1.0/3.5
	//m := 1.0/3.5 //  (N-1)/2

	var sumdat float64 = 0
	var sumgr float64 = 0

	for N := 1; N < 41; N++ {

		n := float64(N)
		g_data := 0.0
		g_data2 := 0.0

		if N < 31 {
			g_data = curve[N]
			g_data2 = curve2[N]
		}

		g_intim   := Curve(n,4)
		g_expt    := Curve(n,8)
		g_tribe   := Curve(n,30)
		g_friends := Curve(n,150)
		g_ufriends := Curve(n,500)

		fmt.Println(n,g_data,g_data2-g_data,g_expt,g_intim,g_tribe,g_friends,g_ufriends)

		sumdat += g_data
		sumgr += g_expt

	}
}

// ************************************************************

func Curve(n,nbar float64) float64 {
	
	const rootpi = 1.77

	nu := 2*(n-1)/nbar

	f := 4.0 * math.Sqrt(nu) * math.Exp(-nu)/((nbar)*rootpi)

	return f
}
