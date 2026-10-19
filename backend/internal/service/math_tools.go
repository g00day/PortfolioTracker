package service

import "math"

func CalculateFloat64Mean(data []float64) float64 {

	summ := 0.0
	for i := 0; i < len(data); i++ {
		summ += data[i]
	}

	return summ / float64(len(data))

}

func CalculateVariance(data []float64) float64 {
	if len(data) <= 1 {
		return 0.0
	}

	summ := 0.0
	mean := CalculateFloat64Mean(data)

	for i := 0; i < len(data); i++ {
		deflection := data[i] - mean
		summ += deflection * deflection
	}

	variance := summ / float64(len(data)-1)

	return variance
}

func CalculateFloat64Deflection(data []float64, mean float64) float64 {

	summ := 0.0

	for i := 0; i < len(data); i++ {
		deflection := data[i] - mean
		summ += deflection * deflection
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

	return covariation / float64(len(data1)-1)

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
