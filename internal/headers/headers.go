package headers

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
)

const CRLF = "\r\n"

var SpecialCharacters = map[byte]struct{}{
	'!': {}, '#': {}, '$': {}, '%': {}, '^': {},
	'&': {}, '*': {}, '-': {}, '_': {},
	'=': {}, '+': {}, '~': {}, '.': {}, '\'': {},
}

type Headers map[string]string

func NewHeaders() Headers {
	return map[string]string{}
}

func (h Headers) Get(k string) string {
	return h[strings.ToLower(k)]
}

func (h Headers) Set(k string, v string) string {
	h[strings.ToLower(k)] = v

	return h.Get(k)
}

func (h Headers) Unset(k string) {
	delete(h, strings.ToLower(k))
}

func (h Headers) Parse(data []byte) (int, bool, error) {
	n := 0
	done := false
	var err error
	for done == false {
		endOfLineIdx := bytes.Index(data, []byte(CRLF))
		if endOfLineIdx == -1 {
			break
		}

		if endOfLineIdx == 0 {
			done = true
			n += len(CRLF)
			break
		}

		line := data[:endOfLineIdx]
		data = data[endOfLineIdx+len(CRLF):]

		splitIdx := bytes.IndexByte(line, ':')
		if i := bytes.IndexByte(line[:splitIdx], ' '); i != -1 {
			err = errors.New("key contains a white space")
			break
		}

		keyUnParsed := line[:splitIdx] // always store the key in lower-case
		if !isValidKey(keyUnParsed) {
			err = errors.New("line contains non parsable chars")
			break
		}

		value := string(line[splitIdx+1:])
		value = strings.TrimSpace(value)
		key := strings.ToLower(string(keyUnParsed))

		// RFC 9110 5.2 says you can stack multiple strings for the same header, e.g
		// Host: localhost
		// Host: test
		// Result: Host: localhost, test
		if _, exists := h[key]; exists {
			value = fmt.Sprintf("%s, %s", h[key], value)
		}

		h[key] = value
		n += endOfLineIdx + len(CRLF)
	}

	return n, done, err
}

func isValidKey(b []byte) bool {
	for _, c := range b {
		if (c >= 'a' && c <= 'z') ||
			(c >= 'A' && c <= 'Z') ||
			(c >= '0' && c <= '9') {
			continue
		}
		if _, ok := SpecialCharacters[c]; !ok {
			return false
		}
	}
	return true
}
