package response

import (
	"errors"
	"fmt"
	"io"
	"netter/internal/headers"
	"strconv"
	"strings"
)

type StatusCode int

const (
	StatusOK                  StatusCode = 200
	StatusBadRequest                     = 400
	StatusInternalServerError            = 500
)

var StatusCodeReasonPhrase = map[StatusCode]string{
	StatusOK:                  "OK",
	StatusBadRequest:          "Bad Request",
	StatusInternalServerError: "Internal Server Error",
}

type WriterState int

const (
	StateStatusLine WriterState = iota
	StateHeader
	StateBody
)

type Writer struct {
	Conn  io.Writer
	State WriterState
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	if w.State != StateStatusLine {
		err := fmt.Sprintf("wrong state for writer should be %v is %v", StateStatusLine, w.State)
		return errors.New(err)
	}

	statusLine := ""
	switch statusCode {
	case StatusOK:
		statusLine = fmt.Sprintf("HTTP/1.1 %d %s\r\n", statusCode, StatusCodeReasonPhrase[statusCode])
	case StatusBadRequest:
		statusLine = fmt.Sprintf("HTTP/1.1 %d %s\r\n", statusCode, StatusCodeReasonPhrase[statusCode])
	case StatusInternalServerError:
		statusLine = fmt.Sprintf("HTTP/1.1 %d %s\r\n", statusCode, StatusCodeReasonPhrase[statusCode])
	default:
		return errors.New("invalid status code")
	}

	_, err := w.Conn.Write([]byte(statusLine))

	w.State = StateHeader
	return err
}

func (w *Writer) WriteHeaders(headers headers.Headers) error {
	if w.State != StateHeader {
		err := fmt.Sprintf("wrong state for writer should be %v is %v", StateHeader, w.State)
		return errors.New(err)
	}

	for k, v := range headers {
		resp := fmt.Sprintf("%s: %s\r\n", k, v)
		_, err := w.Conn.Write([]byte(resp))
		if err != nil {
			return err
		}
	}
	_, err := w.Conn.Write([]byte("\r\n"))
	w.State = StateBody
	return err
}

func (w *Writer) WriteBody(p []byte) (int, error) {
	if w.State != StateBody {
		return 0, fmt.Errorf("wrong state for writer")
	}

	return w.Conn.Write(p)
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	h := headers.NewHeaders()

	h.Set("content-length", strconv.Itoa(contentLen))
	h.Set("connection", "close")
	h.Set("content-type", "text/plain")

	return h
}

func (w *Writer) WriteChunkedBody(p []byte) (int, error) {
	hex := fmt.Sprintf("%x \r\n %s \r\n", len(p), string(p))
	n, err := w.WriteBody([]byte(hex))
	if err != nil {
		return 0, err
	}

	return n, nil
}

func (w *Writer) WriteChunkedBodyDone() (int, error) {
	return w.WriteBody([]byte("0\r\n\r\n"))
}

func (w *Writer) WriteTrailers(h headers.Headers) error {
	trailers := h.Get("trailer")
	if len(trailers) <= 0 {
		return nil
	}

	tSlice := strings.Split(trailers, ",")
	for _, v := range tSlice {
		v = strings.TrimSpace(v)
		resp := fmt.Sprintf("%s: %s \r\n", v, h.Get(v))
		_, err := w.Conn.Write([]byte(resp))
		if err != nil {
			return err
		}
	}
	_, err := w.Conn.Write([]byte("\r\n"))

	return err
}
