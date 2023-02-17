package luhnalgorithm

func Validate(number int) bool {
	/*
		Check number by the luhnal algorithm.
		https://en.wikipedia.org/wiki/Luhn_algorithm
	*/
	return (number%10+checksum(number/10))%10 == 0
}

func checksum(number int) int {
	var lNum int

	for i := 0; number > 0; i++ {
		cur := number % 10

		if i%2 == 0 {
			cur = cur * 2
			if cur > 9 {
				cur = cur%10 + cur/10
			}
		}
		lNum += cur
		number = number / 10
	}
	return lNum % 10
}
