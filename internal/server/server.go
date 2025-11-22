package server

import (
	"io"
	"log"
	"net"
	"netter/internal/request"
	"netter/internal/response"
	"strconv"
)

type Server struct {
	listener net.Listener
	handler  Handler
}

type HandlerError struct {
	message    []byte
	statusCode response.StatusCode
}

type Handler func(w *response.Writer, req *request.Request)

func Serve(port int, handler Handler) (*Server, error) {
	listener, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		return nil, err
	}

	server := &Server{
		listener: listener,
		handler:  handler,
	}

	go server.listen()

	return server, nil
}

func (s *Server) listen() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			log.Println("listener closed:", err)
			return
		}
		log.Println("Connection has been accepted")

		go s.handle(conn)
	}
}

func (s *Server) handle(conn io.ReadWriteCloser) {
	defer conn.Close()

	req, err := request.RequestFromReader(conn)
	if err != nil {
		log.Println(err)
	}

	respW := response.Writer{
		Conn:  conn,
		State: response.StateStatusLine,
	}

	s.handler(&respW, req)
}

func (s *Server) Close() error {
	return s.listener.Close()
}
