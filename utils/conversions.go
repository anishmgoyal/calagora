package utils

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// PriceClientToServer parses a string passed in from the client, and
// returns -1 if the price isn't valid
func PriceClientToServer(input string) (output int, err error) {
	output = 0
	err = nil

	match, _ := regexp.MatchString("^\\$?[0-9]*(\\,[0-9]{3})*(\\.[0-9]{2}|[0-9])$",
		input)

	// Clean up the string
	if match {
		if len(input) > 0 && input[0] == '$' {
			input = input[1:]
		}
		strings.Replace(input, ",", "", -1)
		if len(input) > 2 && input[len(input)-3] == '.' {
			input = input[0:len(input)-3] + input[len(input)-2:]
		} else {
			input = input + "00"
		}
	} else {
		err = errors.New("Invalid format for price")
		return
	}

	output, err = strconv.Atoi(input)
	return
}

// PriceServerToClient converts an int to a price string
func PriceServerToClient(input int) (output string) {
	cents := input % 100
	input = input / 100
	var padding string
	if cents < 10 {
		padding = "0"
	}

	output = fmt.Sprintf("%d.%s%d", input, padding, cents)
	return
}
