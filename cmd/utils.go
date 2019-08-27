package cmd

import (
	"fmt"
	"strings"

	"github.com/timdrysdale/vw/config"
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

func filterClients(clients []config.ClientDetails, filter config.ClientDetails) []config.ClientDetails {
	filteredClients := clients[:0]
	for _, client := range clients {
		if client.Name != filter.Name {
			filteredClients = append(filteredClients, client)
		}
	}
	return filteredClients
}
