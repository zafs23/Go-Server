## Simple Go Server
This project builds a simple server built in Go (Golang) which listens to a TCP server on localhost:$Port  and processes client requests.

### Getting Started
Before running this project on your local machine for development and testing, complete the following steps. 

#### Prerequisites
Before running the server, have Go and Git installed on your machine.  The project is built on ```Go version 1.22.0.```
[Go installation offical website](https://go.dev/learn/)
[Git installtion](https://git-scm.com/book/en/v2/Getting-Started-Installing-Git)

#### Installation
To set up the project, clone the repositoy to your local machine:
```
git clone https://github.com/zafs23/Go-Server.git
cd Go-Server/task-server
```
Then, run the server using: 
```
go run main.go
```
The server will listen to http://localhost:3000. But additional ports can be added to the main.go. 

### Services
This server accepts a task request  terminated by a new line to execute a command where the command absolute path is given in a JSON format. 
Client-side task request format:
```json
{
"command": ["./cmd", "--flag", "argument1", "argument2"],
"timeout": 500
}
```
Upon receiving and accepting a request, the server will handle the task and return a response. 
Task result (server response) format: 
```json
{
"command": ["./cmd", "--flag", "argument1", "argument2"],
"executed_at": 0,
"duration_ms": 0.0,
"exit_code": 0,
"output": "",
"error": "",
}
```
The server processes the task requests in parallel. It can accept and start processing multiple requests in parallel, but the scheduler handles each task synchronously. 

#### Testing
To execute the automated tests, run the following command from the project directory:
```
go test -v ./test/...
```
