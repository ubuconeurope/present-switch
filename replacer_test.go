package main

import "testing"

func TestCanConvertTextWithMap(t *testing.T) {
	var textToUpdate = []byte("Hello Author. Your talk is Title and it is starting at Time.")
	const title = "Go, Go and more Go"
	const author = "John Doe"
	const time = "15:00"
	var newContent = map[string]string{
		"Title":  title,
		"Author": author,
		"Time":   time,
	}

	var expected = "Hello John Doe. Your talk is Go, Go and more Go and it is starting at 15:00."
	var actual = string(replaceBytes(textToUpdate, newContent))

	if expected != actual {
		t.Fail()
		t.Logf("Expected: %s \n Got: %s \n", expected, actual)
	}
}
