package helpers

/*
Takes a given string and desired char length returning a string
of spaces that can be appended to the string to fill the char length
*/
func FillWithSpaces(s string, l int) (spacer string) {
	for i := 0; i <= (l)-len(s); i++ {
		spacer += " "
	}
	return
}
