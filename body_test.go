package main

import "testing"

func TestCanConvertStructToMapOfStrings(t *testing.T) {
	const title = "Talk Test"
	const author = "Test Doe"
	const time = "15:00"

	var b = RequestBody{title, author, time}

	var expected = map[string]string{
		"Title":  title,
		"Author": author,
		"Time":   time,
	}

	var actual = b.convertToMap()

	if actual == nil {
		t.Fail()
		t.Log("Failed generating output.")
	}

	for k := range actual {
		if actual[k] != expected[k] {
			t.Fail()
			t.Logf("Expected %s and got %s", actual[k], expected[k])
		}
	}
}
