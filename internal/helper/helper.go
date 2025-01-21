package helper

const (
	MinLengthOrderNumber = 2
	MaxDig               = 9
	MinDig               = 0
)

func LuhnCheck(orderNumber string) bool {
	l := len([]rune(orderNumber))
	if l < MinLengthOrderNumber {
		return false
	}
	sum := 0
	for pos, chr := range orderNumber {
		dig := int(chr - '0')
		if dig < MinDig || dig > MaxDig {
			return false
		}
		if pos%2 == l%2 {
			dig *= 2
			if dig > MaxDig {
				dig -= 9
			}
		}
		sum += dig
	}
	return sum%10 == 0
}
