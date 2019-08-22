package cmd

import (
	"github.com/spf13/viper"
)

func unmarshalConfig(v *viper.Viper, o interface{}) error {

	err := v.Unmarshal(&o)
	return err

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
