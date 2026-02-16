package luhn

import "strconv"

// Valid checks number using Luhn algorithm
func Valid(number string) bool {
	if len(number) == 0 {
		return false
	}

	sum := 0
	isSecond := false

	for i := len(number) - 1; i >= 0; i-- {
		digit, err := strconv.Atoi(string(number[i]))
		if err != nil {
			return false
		}

		if isSecond {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}

		sum += digit
		isSecond = !isSecond
	}

	return sum%10 == 0
}
