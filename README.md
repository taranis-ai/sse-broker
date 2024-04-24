# SSE Broker

SSE Broker is a robust Server-Sent Events (SSE) broker designed to facilitate real-time communication between a distrubuted system and web clients.

![sse_broker](https://github.com/taranis-ai/sse-broker/assets/6696618/0ec6d3e7-ffd5-42d5-b8dd-b848c0e88700)

## Installation

```sh
git clone https://github.com/taranis-ai/sse-broker.git
cd sse-broker
# Build the application:

go build -o sse-broker

# Run the server:

./sse-broker
```

## Usage

**Publishing Events**

To publish an event, use the `/publish` endpoint with the required headers and JSON payload:

```sh
curl http://localhost:8088/publish -H 'Content-Type: application/json' -H 'X-API-KEY: supersecret' -d '{"data": "Hello, world!", "event": "greeting"}'
```

**Subscribing to Events**

Clients can subscribe to events by connecting to the `/events` endpoint. A valid JWT must be included in the Authorization header:

```sh
curl http://localhost:8088/events  -H "Accept: text/event-stream" -H "Authorization: $sse_token"
```

## Configuration
You can configure the application using environment variables. Below are the available settings:

`JWT_SECRET_KEY`: Secret key for JWT validation. Default is "supersecret".  
`API_KEY`: Secret API key for publishing messages. Default is "supersecret".  
`PORT`: Server port. Default is "8088".  
`SSE_PATH`: Path for the SSE endpoint. Default is "/events".  
`PUBLISH_PATH`: Path for the publishing endpoint. Default is "/publish".  
`TOPICS`: Array of topics that clients can subscribe to.  
