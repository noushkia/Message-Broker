<a href="https://github.com/noushkia">Kianoush Arshi 810198438</a>

## Introduction
In this project, we’ll implement a message queue broker in Go.<br>
The broker has the following features:<br>

* [Synchronous messaging](#Synchronous-messaging)
* [Asynchronous messaging](#Asynchronous-messaging)
* [Overflow handling](#Overflow-handling)
* [Two-way communication](#Two-way-communication)
* [Concurrent connection](#Concurrent-connection)

## How to run the project
The project includes a server, client and a broker.<br>
To run the broker use the following command:<br>
`go run connection.go -port <port_num>-maxl <max queue length> -wsize <number of workers>`<br>
And for running servers and clients use:<br>
`go run client.go -host localhost -port <port_num>`<br>
`go run server.go -host localhost -port <port_num>`<br>

## Connection
The project runs on a tcp server; the broker listens on a socket and the clients/servers 
connect to that socket and communicate with the broker.

## Synchronous messaging
The broker contains a queue of messages i.e. payloads. Each time the server requests a publishing, 
the broker appends the message on the queue with the respectable topic.<br>
This appending is a critical section, thus we use a mutex lock to preserve integrity.<br>
Whenever possible, the messages are sent to the topic's subscribers using subscriber handler.<br>

## Asynchronous messaging
In the async method, the broker sends each message (event) to a channel which 
will then be used by the workers (go routines) to be sent to the subscribers.<br>
Note that the client is non-blocking and with each request, it doesn't expect a response. 
This was made possible by using a go routine for the response handling.

## Overflow handling
If there is a new publish request, the broker checks if the queue of the request topic is full or not.<br>
If the queue was full, it will return an OverflowError and cancel the publishing.

## Two way communication
The client can both subscribe and publish to a topic using -subscribe and -publish commands.<br>

## Concurrent connection
Many servers and clients can connect to a broker, this is made possible using go routines.<br>
Each time a new client is connected to the socket, a new go routine for the specific client is created
and the communication is resumed in that routine.

## References
<div style="position: relative; right: -20px;">
    <li> https://github.com/aaronbieber/tcp-server-client-go </li>
    <li> https://github.com/asim/go-micro </li>
</div>
