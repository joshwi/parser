package regexparser

import (
	"testing"
)

func TestParse(t *testing.T) {

	text := `test-device-gw1#show interface
	GigabitEthernet0/0/0 is administratively down, line protocol is down`

	schema := ReadSchema("./schema.json")
	config := []Config{}

	for n := range schema {
		if schema[n].HwType == "iosxe" && schema[n].Command == "show interface" {
			config = *schema[n].Config
		}
	}

	parser := Compile(config)
	output := Parse(text, *parser, 1, "test", []string{"test-device-gw1"}, []string{"interface"})

	if len(output) == 0 {
		t.Log("Output should return a slice of 1 Node struct")
		t.Fail()
	}
}
