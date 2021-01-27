package regexparser

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"regexp"
)

//Schema structure
type Schema struct {
	Command   string    `json:"command"`
	HwType    string    `json:"hw_type"`
	SwType    string    `json:"sw_type"`
	SwVersion []string  `json:"sw_version"`
	Prompt    string    `json:"prompt"`
	Config    *[]Config `json:"config"`
}

//Config regex structure
type Config struct {
	Match    string    `json:"match"`
	Level    int64     `json:"level"`
	Submatch *[]Config `json:"submatch"`
}

//Parser structure
type Parser struct {
	Match    regexp.Regexp
	Level    int64
	Submatch *[]Parser
}

//Tag structure
type Tag struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

//Node structure
type Node struct {
	Label      string
	Bucket     string
	Properties []Tag
}

//ReadConfig function to convert json to Config struct
func ReadConfig(filename string) []Config {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Print(err)
	}
	var output []Config
	err = json.Unmarshal(data, &output)
	if err != nil {
		fmt.Print(err)
	}
	return output
}

//ReadSchema function to convert json to Schema struct
func ReadSchema(filename string) []Schema {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Print(err)
	}
	var output []Schema
	err = json.Unmarshal(data, &output)
	if err != nil {
		fmt.Print(err)
	}
	return output
}

//GetSchema returns schema file from repo
func GetSchema() []Schema {
	output := ReadSchema("./schema.json")
	return output
}

//Compile config into regex parser
func Compile(config []Config) *[]Parser {

	output := []Parser{}

	for n := range config {

		r := regexp.MustCompile(config[n].Match)
		submatch := *config[n].Submatch
		s := Compile(submatch)

		input := Parser{Match: *r, Level: config[n].Level, Submatch: s}

		output = append(output, input)

	}

	return &output
}

//ParseText function to extract regexp from body of text
func ParseText(text string, parser []Parser, level int64) [][][]Tag {

	output := [][][]Tag{}

	for n := range parser {

		if parser[n].Level <= level {

			search := [][]Tag{}

			r := parser[n].Match

			if len(r.SubexpNames()) > 1 {

				response := r.FindAllStringSubmatch(text, -1)

				if len(response) > 0 {
					for i := range response {
						tags := []Tag{}
						for j, name := range r.SubexpNames() {
							if name != "" {
								tag := Tag{Name: name, Value: response[i][j]}
								tags = append(tags, tag)
							}
						}
						search = append(search, tags)
						submatch := *parser[n].Submatch
						result := ParseText(response[i][0], submatch, level)
						for j := range result {
							if len(result[j]) > 0 {
								log.Println(result[j])
								output = append(output, result[j])
							}
						}
					}
				}

			} else {

				response := r.FindAllString(text, -1)
				if len(response) > 0 {
					for i := range response {
						submatch := *parser[n].Submatch
						result := ParseText(response[i], submatch, level)
						for j := range result {
							if len(result[j]) > 0 {
								output = append(output, result[j])
							}
						}
					}
				}

			}

			if len(search) > 0 {
				output = append(output, search)
			}

		}

	}

	return output

}

//Merge regex matches of equal index to a slice of []Tag
func Merge(input [][][]Tag, index int, output [][]Tag) [][]Tag {
	tag := []Tag{}
	for n := range input {
		if len(input[n]) > index {
			tag = append(tag, input[n][index]...)
		}
	}
	index++
	if len(tag) > 0 {
		output = Merge(input, index, output)
		output = append(output, tag)
	}
	return output
}

//Format data into Node struct
func Format(input [][]Tag, node string, labels []string, keys []string) []Node {
	nodes := []Node{}
	for n := range input {

		label := ``
		for i := range labels {
			if i == 0 {
				label += labels[i]
			} else {
				label += "_" + labels[i]
			}
		}
		for j := range keys {
			for m := range input[n] {
				if keys[j] == input[n][m].Name {
					if len(label) == 0 {
						label += input[n][m].Value
					} else {
						label += "_" + input[n][m].Value
					}
				}
			}
		}

		entry := Node{Label: label, Bucket: node, Properties: input[n]}

		nodes = append(nodes, entry)
	}

	return nodes

}

//Parse text with regex, data output in a slice of Node structs
func Parse(text string, parser []Parser, level int64, node string, labels []string, keys []string) []Node {
	input := ParseText(text, parser, level)
	output := Merge(input, 0, [][]Tag{})
	nodes := Format(output, node, labels, keys)
	return nodes
}
