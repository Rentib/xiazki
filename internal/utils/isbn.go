package utils

import (
	"fmt"
	"strconv"
	"strings"
)

func StringToISBN(isbn string) (string, error) {
	result := isbn
	result = strings.ReplaceAll(result, "-", "")
	result = strings.ReplaceAll(result, " ", "")

	if !IsValidISBN(isbn) {
		return isbn, fmt.Errorf("invalid ISBN: %s", isbn)
	}

	return result, nil
}

func IsValidISBN(isbn string) bool {
	return IsValidISBN10(isbn) || IsValidISBN13(isbn)
}

func IsValidISBN10(isbn string) bool {
	if len(isbn) != 10 {
		return false
	}

	sum := 0
	for i := range 9 {
		digit, err := strconv.Atoi(string(isbn[i]))
		if err != nil {
			return false
		}
		sum += (i + 1) * digit
	}

	checkChar := isbn[9]
	var checkValue int
	if checkChar == 'X' {
		checkValue = 10
	} else {
		digit, err := strconv.Atoi(string(checkChar))
		if err != nil {
			return false
		}
		checkValue = digit
	}
	sum += 10 * checkValue

	return sum%11 == 0
}

func IsValidISBN13(isbn string) bool {
	if len(isbn) != 13 {
		return false
	}

	sum := 0
	for i := range 12 {
		digit, err := strconv.Atoi(string(isbn[i]))
		if err != nil {
			return false
		}
		if i%2 == 0 {
			sum += digit
		} else {
			sum += 3 * digit
		}
	}

	checkDigit, err := strconv.Atoi(string(isbn[12]))
	if err != nil {
		return false
	}
	calculatedCheckDigit := (10 - (sum % 10)) % 10

	return checkDigit == calculatedCheckDigit
}
