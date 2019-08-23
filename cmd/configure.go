package cmd

import (
	"fmt"
	"net/url"
	"os"
)

//func configureEverything() Everything {
//
//}

func constructEndpoint(h *url.URL, inputName string) string {

	h.Path = inputName
	return h.String()
}

func populateInputNames(o *Output) {

	//for each stream, copy each item in Feeds as string into InputNames
	for i, s := range o.Streams {
		feedSlice, _ := s.Feeds.([]interface{})
		for _, feed := range feedSlice {
			o.Streams[i].InputNames = append(o.Streams[i].InputNames, slashify(feed.(string)))
		}

	}

}

func mapEndpoints(o Output, h *url.URL) Endpoints {
	//go through feeds to collect inputs into map

	var e = make(Endpoints)

	for _, v := range o.Streams {
		for _, f := range v.InputNames {
			e[f] = constructEndpoint(h, f)
		}
	}

	return e
}

func expandCaptureCommands(c *Commands, e Endpoints) {

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

func configureChannels(o Output, channelBufferLength int, channelList *[]ChannelDetails) {

	for _, stream := range o.Streams {
		for _, feed := range stream.InputNames {
			newChannel := make(chan Packet, channelBufferLength)
			newChannelDetails := ChannelDetails{Channel: newChannel, Feed: feed, Destination: stream.Destination}
			*channelList = append(*channelList, newChannelDetails)
		}
	}
}

//type ChannelDetails struct {
//	Channel     chan Packet
//	Feed        string
//	Destination string
//}
//type FeedMap map[string][]chan Packet

func configureFeedMap(channelList *[]ChannelDetails, feedMap FeedMap) {

	for _, channel := range *channelList {
		if _, ok := feedMap[channel.Feed]; ok {
			feedMap[channel.Feed] = append(feedMap[channel.Feed], channel.Channel)
		} else {
			feedMap[channel.Feed] = []chan Packet{channel.Channel}
		}
	}

}

//func Configureclientlist(channelList *[]ChannelDetails, clientList *ClientList) {
