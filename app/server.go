package main

import (
	"fmt"
	"net"
	"os"
	"strings"
)

type HttpRequest struct {
	method  string
	path    string
	headers map[string]string
	body    string
}

type HttpResponse struct {
	status  int
	headers map[string]string
	body    string
}

func main() {
	fmt.Println("Logs from your program will appear here!")

	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}
	defer l.Close()

	conn, err := l.Accept()
	if err != nil {
		fmt.Println("Error accepting connection: ", err.Error())
		os.Exit(1)
	}

	httpRequest := parseRequest(conn)
	httpResponse := handleRequest(httpRequest)

	writeResponse(conn, httpResponse)
}

func parseRequest(conn net.Conn) HttpRequest {
	// Create a buffer to read data into
	buffer := make([]byte, 1024)

	// Read data from the client
	_, err := conn.Read(buffer)

	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	splitted := strings.Split(string(buffer), "\r\n")
	requestLine := strings.Split(splitted[0], " ")
	method := requestLine[0]
	path := requestLine[1]

	return HttpRequest{method: method, path: path}
}

func handleRequest(request HttpRequest) HttpResponse {
	if request.method == "GET" {
		if request.path == "/" {
			return HttpResponse{status: 200, body: "OK"}
		} else {
			return HttpResponse{status: 404, body: "Not Found"}
		}

	} else {
		return HttpResponse{status: 405, body: "Method Not Allowed"}
	}
}

func writeResponse(conn net.Conn, response HttpResponse) {
	conn.Write([]byte(fmt.Sprintf("HTTP/1.1 %d %s\r\n\r\n", response.status, response.body)))
	conn.Close()
}
