package service

func CalculateLuhn(number int) int {
	checkNumber := checksum(number)

	if checkNumber == 0 {
		return 0
	}
	return 10 - checkNumber
}

func ValidLuhn(number int) bool {
	return (number%10+checksum(number/10))%10 == 0
}

func checksum(number int) int {
	var luhn int

	for i := 0; number > 0; i++ {
		cur := number % 10

		if i%2 == 0 {
			cur = cur * 2
			if cur > 9 {
				cur = cur%10 + cur/10
			}
		}

		luhn += cur
		number = number / 10
	}
	return luhn % 10
}
