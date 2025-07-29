package lunavalidate

import (
	"log"
	"strconv"
)

// валидация алгоритмом луна
func Validate(number string) bool {
	log.Println(number)
	sum := 0
	if len(number) == 0 {
		return false
	}
	for i := 0; i < len(number); i++ {
		n, err := strconv.Atoi(string(number[i]))
		if err != nil {
			return false
		}
		if i%2 == 1 {
			sum += n
		} else {
			prom := n * 2
			if prom > 9 {
				prom -= 9
			}
			sum += prom
		}

	}
	log.Println(sum)
	return sum%10 == 0
}
