package cmd

import (
	"fmt"
	"net/url"
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

func mapEndpoints(o Output, h *url.URL) Endpoints {

	//go through feeds to collect inputs into map

	var e = make(Endpoints)

	for _, v := range o.Streams {
		fmt.Println(v)

		//https://github.com/go-yaml/yaml/issues/282
		//a, _ := d.Data.([]interface{})
		//fmt.Println(len(a))
		//for _, x := range a {
		//	y := x.(string)
		//	fmt.Println(y)
		//}
		//for _, f := range v.InputNames {
		//	fmt.Println("\t", f)
		//	e[f] = constructEndpoint(h, f)
		//
		//}
	}

	return e

}

func expandCaptureCommands(o interface{}, c interface{}) error {
	//
	//	if err != nil {
	//		fmt.Println("Didnt unpack streams config")
	//		return err
	//	} else {
	//		for _, stream := range outs.Streams {
	//			fmt.Printf("destination:%v\n", stream.Destination)
	//			for _, name := range stream.InputNames {
	//				inputAddresses[name] = fmt.Sprintf("%s/%s/", listen, name)
	//				fmt.Printf("%v\v", inputAddresses[name])
	//
	//			} //for
	//
	//		} //for
	//
	//		for _, name := range inputAddresses {
	//			inputChannels[name] = make(chan Packet)
	//			fmt.Printf("%s:%s\n", name, inputAddresses[name])
	//		}
	//
	//	} //else
	return nil
}
