package cmd

import (
	"fmt"
	"log"
	"net/url"
	"os"

	"github.com/phayes/freeport"
	"github.com/timdrysdale/vw/config"
)

//func configureEverything() Everything {
//
//}

func constructEndpoint(h *url.URL, inputName string) string {

	h.Path = inputName
	return h.String()
}

func populateInputNames(s *config.Streams) {

	//for each stream, copy each item in Feeds as string into InputNames
	for i, stream := range s.Stream {
		feedSlice, _ := stream.From.([]interface{})
		for _, feed := range feedSlice {
			s.Stream[i].InputNames = append(s.Stream[i].InputNames, slashify(feed.(string)))
		}

	}

}

func mapEndpoints(s config.Streams, h *url.URL) config.Endpoints {
	//go through feeds to collect inputs into map

	var e = make(config.Endpoints)

	for _, v := range s.Stream {
		for _, f := range v.InputNames {
			e[f] = constructEndpoint(h, f)
		}
	}

	return e
}

func expandCaptureCommands(c *config.Commands, e config.Endpoints) {

	// we rely on e being in scope for the mapper when it runs
	mapper := func(placeholderName string) string {

		if val, ok := e[slashify(placeholderName)]; ok {
			return val
		} else {
			// we don't modify unknown env variables
			return fmt.Sprintf("${%s}", placeholderName)
		}

	}

	for i, raw := range c.Commands {
		c.Commands[i] = os.Expand(raw, mapper)
	}

}

func expandDestination(destination string, outurl string, uuid string, session string) string {

	mapper := func(placeholderName string) string {
		switch placeholderName {
		case "outurl":
			return outurl
		case "uuid":
			return uuid
		case "session":
			return session
		default:
			return fmt.Sprintf("${%s}", placeholderName)
		}
	}
	destination = os.Expand(destination, mapper)
	return destination
}

func configureChannels(s config.Streams, channelBufferLength int, channelList *[]config.ChannelDetails, outurl string, uuid string, session string) {

	for _, stream := range s.Stream {
		for _, feed := range stream.InputNames {
			newChannel := make(chan config.Packet, channelBufferLength)
			destination := expandDestination(stream.To, outurl, uuid, session)
			newChannelDetails := config.ChannelDetails{Channel: newChannel, From: feed, To: destination}
			*channelList = append(*channelList, newChannelDetails)
		}
	}
}

func configureFeedMap(channelList *[]config.ChannelDetails, feedMap config.FeedMap) {

	for _, channel := range *channelList {
		if _, ok := feedMap[channel.From]; ok {
			feedMap[channel.From] = append(feedMap[channel.From], channel.Channel)
		} else {
			feedMap[channel.From] = []chan config.Packet{channel.Channel}
		}
	}
}

func configureClientMap(channelList *[]config.ChannelDetails, clientMap config.ClientMap) {

	for _, channel := range *channelList {
		if _, ok := clientMap[channel.To]; ok {
			clientMap[channel.To] = append(clientMap[channel.To], channel.Channel)
		} else {
			clientMap[channel.To] = []chan config.Packet{channel.Channel}
		}
	}
}

func getHost() *url.URL {
	//get a free port
	port, err := freeport.GetFreePort()
	if err != nil {
		log.Printf("Error getting free port %v", err)
	}

	addr := fmt.Sprintf("http://127.0.0.1:%d", port)
	h, err := url.Parse(addr)
	return h
}
