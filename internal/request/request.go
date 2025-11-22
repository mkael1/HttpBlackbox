package request

import (
	"errors"
	"fmt"
	"io"
	"netter/internal/headers"
	"strconv"
	"strings"
	"unicode"
)

type RequestState int

const (
	Initialized RequestState = iota
	ParsingHeaders
	ParsingBody
	Done
)

type Request struct {
	RequestLine RequestLine
	Headers     headers.Headers
	Body        []byte
	state       RequestState
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

const ErrRequestMalformed = "request start-line is malformed"

func RequestFromReader(reader io.Reader) (*Request, error) {
	request := Request{
		state:   Initialized,
		Headers: headers.NewHeaders(),
	}

	// This buffer could overflow with a large body / request
	buf := make([]byte, 128)
	bufIdx := 0
	for request.state != Done {
		numBytesRead, err := reader.Read(buf[bufIdx:])
		if numBytesRead == 0 && err == nil {
			request.state = Done
			break
		}
		if err != nil {
			if errors.Is(err, io.EOF) {
				// If it reaches the end of file and didn't get the expected content-length, throw an error
				ctLength, _ := strconv.Atoi(request.Headers.Get("content-length"))
				if len(request.Body) < ctLength {
					err := fmt.Sprintf("body exceeds declared content-length, received %v, expected %v", len(request.Body), ctLength)
					return nil, errors.New(err)
				}
				request.state = Done
				break
			}
			return nil, err
		}
		bufIdx += numBytesRead
		readN, err := request.parse(buf[:bufIdx])
		if err != nil {
			return nil, err
		}

		// Copies the leftover bytes that weren't used in the parser
		// This is to make sure they aren't lost
		copy(buf, buf[readN:bufIdx])
		bufIdx -= readN
	}

	return &request, nil
}

func (r *Request) parse(data []byte) (int, error) {
	bytesCount := 0
	var err error
	for r.state != Done {
		n := 0
		n, err = r.parseSingle(data[bytesCount:])
		if n == 0 {
			break
		}
		bytesCount += n
	}

	return bytesCount, err
}

func (r *Request) parseSingle(data []byte) (int, error) {
	switch r.state {
	case Initialized:
		rl, n, err := parseRequestLine(string(data))
		if err != nil {
			return 0, err
		}

		// No valid request line — keep waiting for more bytes
		if n <= 0 {
			return 0, nil
		}

		r.RequestLine = *rl
		r.state = ParsingHeaders

		// Add CRLF length (there is always a CRLF after the request line)
		return n, nil

	case ParsingHeaders:
		n, done, err := r.Headers.Parse(data)
		if err != nil {
			return 0, err
		}

		if done {
			r.state = ParsingBody
		}

		return n, nil
	case ParsingBody:
		clHeader := r.Headers.Get("content-length")
		if clHeader == "" {
			// No body expected
			r.state = Done
			return 0, nil
		}

		contentLength, err := strconv.Atoi(clHeader)
		if err != nil || contentLength < 0 {
			return 0, fmt.Errorf("invalid content-length header")
		}

		r.Body = data
		current := len(r.Body)
		if current > contentLength {
			err := fmt.Sprintf("body exceeds declared content-length, received %v, expected %v", current, contentLength)
			return 0, errors.New(err)
		}

		if current == contentLength {
			// Full body received
			r.state = Done
			return current, nil
		}

		// Not done yet, need more body data
		return 0, nil
	case Done:
		return 0, nil

	default:
		return 0, fmt.Errorf("unknown request state: %v", r.state)
	}
}

func parseRequestLine(request string) (*RequestLine, int, error) {
	idx := strings.Index(request, "\r\n")
	if idx == -1 {
		return nil, 0, nil
	}

	data := request[:idx]
	requestLineUnparsed := strings.Split(data, " ")
	if len(requestLineUnparsed) != 3 {
		return nil, 0, errors.New(ErrRequestMalformed)
	}
	method := requestLineUnparsed[0]
	for _, char := range method {
		if !unicode.IsLetter(char) {
			return nil, 0, errors.New(ErrRequestMalformed)
		}
	}
	if strings.ToUpper(method) != method {
		return nil, 0, errors.New(ErrRequestMalformed)
	}

	requestTarget := requestLineUnparsed[1]

	httpVersion := strings.Split(requestLineUnparsed[2], "/")
	if len(httpVersion) != 2 || httpVersion[1] != "1.1" {
		return nil, 0, errors.New(ErrRequestMalformed)
	}

	return &RequestLine{
		Method:        method,
		RequestTarget: requestTarget,
		HttpVersion:   httpVersion[1],
	}, len(data) + len(headers.CRLF), nil
}
