package utils

import (
	"math/rand/v2"
	"strconv"
)


func RandomStringNumber(n int) (str string) {
	number := rand.IntN(n)
	str = strconv.FormatInt(int64(number), 12)
	return str
}