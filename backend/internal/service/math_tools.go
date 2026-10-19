package service

import "math"

func CalculateFloat64Mean(data []float64) float64 {

	summ := 0.0
	for i := 0; i < len(data); i++ {
		summ += data[i]
	}

	return summ / float64(len(data))

}

func CalculateFloat64Deflection(data []float64, mean float64) float64 {

	summ := 0.0

	for i := 0; i < len(data); i++ {
		summ += math.Pow(data[i]-mean, 2)
	}

	return math.Sqrt(summ)

}

func CalculateCovariation(data1, data2 []float64) float64 {

	covariation := 0.0

	data1Mean := CalculateFloat64Mean(data1)
	data2Mean := CalculateFloat64Mean(data2)

	for i := 0; i < len(data1); i++ {
		covariation += (data1[i] - data1Mean) * (data2[i] - data2Mean)
	}

	return covariation

}

func CalculatePirsonCorrelation(data1, data2 []float64) float64 {

	covariation := 0.0

	data1Mean := CalculateFloat64Mean(data1)
	data2Mean := CalculateFloat64Mean(data2)

	for i := 0; i < len(data1); i++ {
		covariation += (data1[i] - data1Mean) * (data2[i] - data2Mean)
	}

	data1Deflection := CalculateFloat64Deflection(data1, data1Mean)
	data2Deflection := CalculateFloat64Deflection(data2, data2Mean)

	return covariation / (data1Deflection * data2Deflection)
}
