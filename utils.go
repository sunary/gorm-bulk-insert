package bulkinsert

import (
	"strings"
)

func toSnakeCase(input string) string {
	var sb strings.Builder
	var upperCount int
	for i, c := range input {
		switch {
		case isUppercase(c):
			if i > 0 && (upperCount == 0 || nextIsLower(input, i)) {
				sb.WriteByte('_')
			}
			sb.WriteByte(byte(c - 'A' + 'a'))
			upperCount++

		case isLowercase(c):
			sb.WriteByte(byte(c))
			upperCount = 0

		case isDigit(c):
			if i == 0 {
				panic("Identifier must start with a character: `" + input + "`")
			}
			sb.WriteByte(byte(c))

		default:
			panic("Invalid identifier: `" + input + "`")
		}
	}

	return sb.String()
}

// nextIsLower The next character is lower case, but not the last 's'.
// nextIsLower("HTMLFile", 1) expected: "html_file"
// => true
// nextIsLower("URLs", -1) expected: "urls"
// => false
func nextIsLower(input string, i int) bool {
	i++
	if i >= len(input) {
		return false
	}

	c := input[i]
	if c == 's' && i == len(input)-1 {
		return false
	}

	return isLowercase(rune(c))
}

func isDigit(r rune) bool {
	return r >= '0' && r <= '9'
}

func isLowercase(r rune) bool {
	return r >= 'a' && r <= 'z'
}

func isUppercase(r rune) bool {
	return r >= 'A' && r <= 'Z'
}
