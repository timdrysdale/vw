package cmd

import (
	"fmt"
	"log"
	"net/url"
	"os"

	"github.com/phayes/freeport"
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

func expandCaptureCommands(c *Commands, e Endpoints, v Variables) {

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

	//now do variables
	mapper = func(placeholderName string) string {
		if val, ok := v.Vars[placeholderName]; ok {
			return val
		} else {
			return fmt.Sprintf("${%s}", placeholderName)
		}
	}

	for i, raw := range c.Commands {
		c.Commands[i] = os.Expand(raw, mapper)
	}

}

func expandDestination(destination string, variables Variables) string {

	mapper := func(placeholderName string) string {
		if val, ok := variables.Vars[placeholderName]; ok {
			return val
		} else {
			return fmt.Sprintf("${%s}", placeholderName)
		}
	}
	destination = os.Expand(destination, mapper)
	return destination
}

func expandDestinations(o *Output, v Variables) {

	for i, stream := range o.Streams {
		o.Streams[i].Destination = expandDestination(stream.Destination, v)
		fmt.Println(expandDestination(stream.Destination, v))
	}
	fmt.Printf("Vars: %v", v.Vars)
}

//type HTTPOptions struct {
//	Port      int `yaml:"port"`
//	WaitMS    int `yaml:"waitMS"`
//	FlushMS   int `yaml:"flushMS"`
//	TimeoutMS int `yaml:"timeoutMS"`
//}
//		err = viper.Unmarshal(&httpOpts)
//		if err != nil {
//			log.Fatalf("Viper unmarshal commands failed: %v", err)
//		}
//
//

//func configureChannels(o Output, channelBufferLength int, channelList *[]ChannelDetails, variables Variables) {
//	for _, stream := range o.Streams {
//		for _, feed := range stream.InputNames {
//			newChannel := make(chan Packet, channelBufferLength)
//			destination := expandDestination(stream.Destination, variables)
//			newChannelDetails := ChannelDetails{Channel: newChannel, Feed: feed, Destination: destination}
//			*channelList = append(*channelList, newChannelDetails)
//		}
//	}
//}
//
//func configureFeedMap(channelList *[]ChannelDetails, feedMap FeedMap) {
//
//	for _, channel := range *channelList {
//		if _, ok := feedMap[channel.Feed]; ok {
//			feedMap[channel.Feed] = append(feedMap[channel.Feed], channel.Channel)
//		} else {
//			feedMap[channel.Feed] = []chan Packet{channel.Channel}
//		}
//	}
//}
//
//func configureClientMap(channelList *[]ChannelDetails, clientMap ClientMap) {
//
//	for _, channel := range *channelList {
//		if _, ok := clientMap[channel.Destination]; ok {
//			clientMap[channel.Destination] = append(clientMap[channel.Destination], channel.Channel)
//		} else {
//			clientMap[channel.Destination] = []chan Packet{channel.Channel}
//		}
//	}
//}

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
