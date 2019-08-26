package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
)

// ReplaceInTemplate will take care of importing a given template and replace the
// variable strings as the user wants.
// With testing in mind, the actual replacement of variables will occur in the replaceByte method.
func ReplaceInTemplate(id string, content map[string]string) {
	var path = fmt.Sprintf("romm_%s_template", id)
	dat, err := ioutil.ReadFile(fmt.Sprintf("/templates/%s", path))
	if err != nil {
		log.Fatal("Errors occurred when reading file. Does it exist?")
	}
	ioutil.WriteFile(path, replaceBytes(dat, content), 0664)
}

func replaceBytes(original []byte, newContent map[string]string) []byte {
	for key, val := range newContent {
		original = bytes.ReplaceAll(original, []byte(key), []byte(val))
	}
	return original
}
