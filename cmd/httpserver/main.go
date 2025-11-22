package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"netter/internal/request"
	"netter/internal/response"
	"netter/internal/server"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const port = 42069

func main() {

	s, err := server.Serve(port, HandlerFunc)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer s.Close()

	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}

func HandlerFunc(w *response.Writer, r *request.Request) {
	headers := response.GetDefaultHeaders(0)
	headers.Set("content-type", "text/html")
	var buf bytes.Buffer

	if strings.HasPrefix(r.RequestLine.RequestTarget, "/httpbin") {
		route := strings.TrimPrefix(r.RequestLine.RequestTarget, "/httpbin")
		resp, err := http.Get(fmt.Sprintf("https://httpbin.org%s", route))
		if err != nil {
			data := ErrorPageData{
				Status:  response.StatusInternalServerError,
				Reason:  response.StatusCodeReasonPhrase[response.StatusInternalServerError],
				Message: "Internal server error",
			}

			w.WriteStatusLine(response.StatusInternalServerError)
			w.WriteHeaders(headers)

			if err := errorTemplate.Execute(&buf, data); err != nil {
				panic(err)
			}

			return
		}

		headers.Unset("Content-Length")
		headers.Set("transfer-encoding", "chunked")
		headers.Set("trailer", "X-Content-SHA256, X-Content-Length")
		w.WriteStatusLine(response.StatusOK)
		w.WriteHeaders(headers)

		chunkBuf := make([]byte, 64)
		var fullBody []byte
		for {
			n, err := resp.Body.Read(chunkBuf)
			time.Sleep(time.Millisecond * 20)

			if err != nil && err != io.EOF {
				data := ErrorPageData{
					Status:  response.StatusInternalServerError,
					Reason:  response.StatusCodeReasonPhrase[response.StatusInternalServerError],
					Message: "Internal server error",
				}

				w.WriteStatusLine(response.StatusInternalServerError)
				w.WriteHeaders(headers)

				if err := errorTemplate.Execute(&buf, data); err != nil {
					panic(err)
				}

				break
			}
			if n == 0 {
				w.WriteChunkedBodyDone()
				break
			}
			w.WriteChunkedBody(chunkBuf[:n])
			fullBody = append(fullBody, chunkBuf[:n]...)
		}
		hash := sha256.Sum256(fullBody)
		headers.Set("x-content-sha256", base64.StdEncoding.EncodeToString(hash[:]))
		headers.Set("x-content-length", strconv.Itoa(len(fullBody)))
		w.WriteTrailers(headers)
		return
	}

	switch r.RequestLine.RequestTarget {
	case "/yourproblem":
		data := ErrorPageData{
			Status:  response.StatusBadRequest,
			Reason:  response.StatusCodeReasonPhrase[response.StatusBadRequest],
			Message: "Your request honestly kinda sucked",
		}

		if err := errorTemplate.Execute(&buf, data); err != nil {
			panic(err)
		}

		w.WriteStatusLine(response.StatusBadRequest)
		headers.Set("content-length", strconv.Itoa(len(buf.Bytes())))
		w.WriteHeaders(headers)
		w.WriteBody(buf.Bytes())
	case "/myproblem":
		data := ErrorPageData{
			Status:  response.StatusInternalServerError,
			Reason:  response.StatusCodeReasonPhrase[response.StatusInternalServerError],
			Message: "Okay, you know what? This one is on me.",
		}

		if err := errorTemplate.Execute(&buf, data); err != nil {
			panic(err)
		}

		w.WriteStatusLine(response.StatusInternalServerError)
		headers.Set("content-length", strconv.Itoa(len(buf.Bytes())))
		w.WriteHeaders(headers)
		w.WriteBody(buf.Bytes())

	case "/video":
		headers.Set("content-type", "video/mp4")
		data, err := os.ReadFile("E:\\Go\\httpfromtcp\\assets\\vim.mp4")
		if err != nil {
			panic(err)
		}

		w.WriteStatusLine(response.StatusOK)
		headers.Set("content-length", strconv.Itoa(len(data)))
		w.WriteHeaders(headers)
		w.WriteBody(data)
	default:
		data := ErrorPageData{
			Status:  int(response.StatusOK),
			Reason:  "Success!",
			Message: "Your request was an absolute banger.",
		}

		if err := errorTemplate.Execute(&buf, data); err != nil {
			panic(err)
		}

		w.WriteStatusLine(response.StatusOK)
		headers.Set("content-length", strconv.Itoa(len(buf.Bytes())))
		w.WriteHeaders(headers)
		w.WriteBody(buf.Bytes())
	}
}

var errorTemplate = template.Must(template.New("error").Parse(`<html>
  <head>
    <title>{{.Status}} {{.Reason}}</title>
  </head>
  <body>
    <h1>{{.Reason}}</h1>
    <p>{{.Message}}</p>
  </body>
</html>`))

type ErrorPageData struct {
	Status  int
	Reason  string
	Message string
}
