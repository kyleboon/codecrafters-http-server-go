package main

import (
	"fmt"
	"net"
	"os"
	"strings"
)

type Status string

const (
	OK               = "200 OK"
	NotFound         = "404 Not Found"
	MethodNotAllowed = "405 Method Not Allowed"
)

const HTTP_VERSION = "HTTP/1.1"

type HttpRequest struct {
	method  string
	path    string
	headers map[string]string
	body    string
}

type HttpResponse struct {
	status  Status
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

	// parse request headers
	headers := make(map[string]string)
	for i := 1; i < len(splitted); i++ {
		if splitted[i] == "" {
			break
		}
		header := strings.Split(splitted[i], ": ")
		headers[strings.ToUpper(header[0])] = header[1]
	}

	return HttpRequest{method: method, path: path, headers: headers}
}

func handleRequest(request HttpRequest) HttpResponse {
	if request.method == "GET" {
		if request.path == "/" {
			return HttpResponse{status: OK}
		}

		if request.path == "/user-agent" {
			return HttpResponse{status: OK, body: request.headers["USER-AGENT"]}
		}
		path_parts := strings.Split(request.path, "/")
		if len(path_parts) != 3 {
			return HttpResponse{status: NotFound}
		}
		if path_parts[1] == "echo" {
			return HttpResponse{status: OK, body: path_parts[2]}
		} else {
			return HttpResponse{status: NotFound}
		}

	} else {
		return HttpResponse{status: MethodNotAllowed, body: "Method Not Allowed"}
	}
}

func writeResponse(conn net.Conn, response HttpResponse) {
	response.headers = make(map[string]string)
	response.headers["Content-Type"] = "text/plain"
	response.headers["Content-Length"] = fmt.Sprintf("%d", len(response.body))

	// build up the response string
	response_string := fmt.Sprintf("%s %s\r\n", HTTP_VERSION, response.status)
	response_string += fmt.Sprintf("Content-Type: %s\r\n", response.headers["Content-Type"])
	response_string += fmt.Sprintf("Content-Length: %s\r\n", response.headers["Content-Length"])
	response_string += fmt.Sprintf("\r\n%s", response.body)

	conn.Write([]byte(response_string))
	conn.Close()
}
