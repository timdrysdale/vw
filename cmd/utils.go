package cmd

import (
	"fmt"
	"strings"
)

func clean(in string) string {

	return strings.TrimRight(in, "\r\n")

}

func slashify(path string) string {

	//remove trailing slash (that's for directories)
	path = strings.TrimSuffix(path, "/")

	//ensure leading slash without needing it in config
	path = strings.TrimPrefix(path, "/")
	path = fmt.Sprintf("/%s", path)

	return path

}
