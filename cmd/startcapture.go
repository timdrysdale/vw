package cmd

import (
	"net/url"
	"os"
)

//func unmarshalConfig(v *viper.Viper, o interface{}) error {
//
//	err := v.Unmarshal(&o)
//	return err
//
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
			o.Streams[i].InputNames = append(o.Streams[i].InputNames, feed.(string))
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
		if val, ok := e[placeholderName]; ok {
			return val
		} else {
			return placeholderName //don't change what we don't know about
		}

	}

	for i, raw := range c.Commands {
		c.Commands[i] = os.Expand(raw, mapper)
	}

}
