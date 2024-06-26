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
	CREATED          = "201 Created"
)

const HTTP_VERSION = "HTTP/1.1"

type HttpRequest struct {
	method  string
	path    string
	headers map[string]string
	body    string
}

type HttpResponse struct {
	status       Status
	content_type string
	headers      map[string]string
	body         string
}

func main() {
	fmt.Println("Logs from your program will appear here!")

	var directory string
	if len(os.Args) == 3 {
		directory = os.Args[2]
	}

	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}
	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}

		go handleConnection(conn, directory)
	}
}

func handleConnection(conn net.Conn, directory string) {
	httpRequest := parseRequest(conn)
	httpResponse := handleRequest(httpRequest, directory)
	writeResponse(conn, httpResponse)
}

func parseRequest(conn net.Conn) HttpRequest {
	// Create a buffer to read data into
	buffer := make([]byte, 1024)

	n, err := conn.Read(buffer)

	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	buffer = buffer[0:n]
	splitted := strings.Split(string(buffer), "\r\n")
	requestLine := strings.Split(splitted[0], " ")
	method := requestLine[0]
	path := requestLine[1]
	var body string

	// parse request headers
	headers := make(map[string]string)
	for i := 1; i < len(splitted); i++ {
		// TODO: I think i need to only read the number of bytes specified in the content-length header
		if splitted[i] == "" {
			fmt.Println("Last line of headers")
			if i+1 < len(splitted) {
				fmt.Println("Body found")
				body = splitted[i+1]
				fmt.Println(body)
				fmt.Println("The was the body")
			}
			break
		}
		header := strings.Split(splitted[i], ": ")
		headers[strings.ToUpper(header[0])] = header[1]
	}

	return HttpRequest{method: method, path: path, headers: headers, body: body}
}

func handleRequest(request HttpRequest, directory string) HttpResponse {
	if request.method == "GET" {
		switch {
		case request.path == "/":
			return HttpResponse{status: OK, content_type: "text/plain"}
		case request.path == "/user-agent":
			return HttpResponse{status: OK, content_type: "text/plain", body: request.headers["USER-AGENT"]}
		case strings.HasPrefix(request.path, "/echo/"):
			return HttpResponse{status: OK, content_type: "text/plain", body: request.path[6:]}
		case strings.HasPrefix(request.path, "/files/"):
			filePath := directory + request.path[6:]
			file_contents, err := os.ReadFile(filePath)
			if err != nil {
				return HttpResponse{status: NotFound, content_type: "text/plain"}
			}
			return HttpResponse{status: OK, content_type: "application/octet-stream", body: string(file_contents)}

		default:
			return HttpResponse{status: NotFound, content_type: "text/plain"}
		}
	} else {
		switch {
		case strings.HasPrefix(request.path, "/files/"):
			filePath := directory + request.path[6:]
			err := os.WriteFile(filePath, []byte(request.body), 0644)
			if err != nil {
				panic(err)
			}

			return HttpResponse{status: CREATED, content_type: "text/plain"}
		default:
			return HttpResponse{status: NotFound, content_type: "text/plain"}

		}
	}
}

func writeResponse(conn net.Conn, response HttpResponse) {
	response.headers = make(map[string]string)
	response.headers["Content-Type"] = response.content_type
	response.headers["Content-Length"] = fmt.Sprintf("%d", len(response.body))

	// build up the response string
	response_string := fmt.Sprintf("%s %s\r\n", HTTP_VERSION, response.status)
	response_string += fmt.Sprintf("Content-Type: %s\r\n", response.headers["Content-Type"])
	response_string += fmt.Sprintf("Content-Length: %s\r\n", response.headers["Content-Length"])
	response_string += fmt.Sprintf("\r\n%s", response.body)

	conn.Write([]byte(response_string))
	conn.Close()
}
