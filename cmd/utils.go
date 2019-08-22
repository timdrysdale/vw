package cmd

import "strings"

func clean(in string) string {

	return strings.TrimRight(in, "\r\n")

}
