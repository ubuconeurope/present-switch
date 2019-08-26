package main

type RequestBody struct {
	Title  string
	Author string
	Time   string
}

func (b *RequestBody) convertToMap() map[string]string {
	return map[string]string{
		"Title":  b.Title,
		"Author": b.Author,
		"Time":   b.Time,
	}
}
