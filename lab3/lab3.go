package lab3

import (
	"strconv"
	"strings"
)

func GenerateMessages(producerAmount, messageAmount int) [][]string {
	letter := 'A'
	result := make([][]string, producerAmount)

	for p := 0; p < producerAmount; p++ {
		messages := make([]string, messageAmount)
		for m := 0; m < messageAmount; m++ {
			message := strings.Builder{}
			for i := 0; i < 4; i++ {
				message.WriteRune(letter)
			}
			message.WriteString(strconv.Itoa(m + 1))
			messages[m] = message.String()
		}
		result[p] = messages
		letter++
	}
	return result
}

func GenerateIntMessages(producerAmount, messageAmount int) [][]int {
	letter := '1'
	result := make([][]int, producerAmount)

	for p := 0; p < producerAmount; p++ {
		messages := make([]int, messageAmount)
		for m := 0; m < messageAmount; m++ {
			message := strings.Builder{}
			for i := 0; i < 4; i++ {
				message.WriteRune(letter)
			}
			message.WriteString(strconv.Itoa(m + 1))
			number, _ := strconv.Atoi(message.String())
			messages[m] = number
		}
		result[p] = messages
		letter++
	}
	return result
}
