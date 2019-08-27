package cmd

import (
	"fmt"
	"log"
	"testing"

	"gopkg.in/yaml.v3"
)

type StructA struct {
	A string `yaml:"a"`
}

type StructB struct {
	// Embedded structs are not treated as embedded in YAML by default. To do that,
	// add the ",inline" annotation below
	StructA `yaml:",inline"`
	B       string `yaml:"b"`
}

var data = `
a: a string from struct A
b: a string from struct B
`

//type structE struct {
//	encoding map[string]string
//}

type StructC struct {
	name   string   `yaml:"c"`
	things []string `yaml:",inline"`
}

type StructD struct {
	// Embedded structs are not treated as embedded in YAML by default. To do that,
	// add the ",inline" annotation below
	StructC `yaml:"stuff,inline"`
	B       string `yaml:"b"`
}

var dataB = `
a: a string from struct A
b: a string from struct B
stuff: 
  name:goofstuff
  things:
    -bag
    -tea
    -pot
`

func TestYamlv3A(t *testing.T) {

	var b StructB

	err := yaml.Unmarshal([]byte(data), &b)
	if err != nil {
		log.Fatalf("cannot unmarshal data: %v", err)
	}
	fmt.Println(b.A)
	fmt.Println(b.B)
}

func TestYamlv3B(t *testing.T) {

	var d StructD

	err := yaml.Unmarshal([]byte(dataB), &d)
	if err != nil {
		log.Fatalf("cannot unmarshal data: %v", err)
	}
	fmt.Println(d.B)
	fmt.Println(d.name)
}
