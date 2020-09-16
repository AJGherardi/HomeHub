package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/url"
	"sync"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/gorilla/websocket"
	"github.com/vektah/gqlparser/gqlerror"
)

// This file containes a cut down and adapted variant of the web socket transport from the gqlgen library

// Message types
const (
	connectionInitMsg      = "connection_init"
	connectionTerminateMsg = "connection_terminate"
	startMsg               = "start"
	stopMsg                = "stop"
	connectionAckMsg       = "connection_ack"
	connectionErrorMsg     = "connection_error"
	dataMsg                = "data"
	errorMsg               = "error"
	completeMsg            = "complete"
	connectionKeepAliveMsg = "ka"
)

type (
	wsConnection struct {
		KeepAlivePingInterval time.Duration
		ctx                   context.Context
		conn                  *websocket.Conn
		active                map[string]context.CancelFunc
		mu                    sync.Mutex
		keepAliveTicker       *time.Ticker
		exec                  graphql.GraphExecutor
	}
	operationMessage struct {
		Payload json.RawMessage `json:"payload,omitempty"`
		ID      string          `json:"id,omitempty"`
		Type    string          `json:"type"`
	}
)

// ConnectAndServe Creates a connection to the HomeService and serves the graphql schema
func ConnectAndServe(exec graphql.GraphExecutor) {
	// Connect to service
	u := url.URL{Scheme: "ws", Host: "localhost:8080", Path: "/hub"}
	log.Printf("connecting to %s", u.String())
	ws, _, _ := websocket.DefaultDialer.Dial(u.String(), nil)
	// Create connection
	conn := wsConnection{
		active:                map[string]context.CancelFunc{},
		conn:                  ws,
		ctx:                   context.TODO(),
		exec:                  exec,
		KeepAlivePingInterval: 10 * time.Second,
	}
	// Handle initial message
	if !conn.init() {
		return
	}
	// Read and handle messages
	conn.run()
}

func (c *wsConnection) init() bool {
	message := c.readOp()
	// Close on undecodable message
	if message == nil {
		c.close(websocket.CloseProtocolError, "decoding error")
		return false
	}
	// Match type
	switch message.Type {
	case connectionInitMsg:
		// Write ack and keep alive message
		c.write(&operationMessage{Type: connectionAckMsg})
		c.write(&operationMessage{Type: connectionKeepAliveMsg})
	case connectionTerminateMsg:
		c.close(websocket.CloseNormalClosure, "terminated")
		return false
	default:
		c.sendConnectionError("unexpected message %s", message.Type)
		c.close(websocket.CloseProtocolError, "unexpected message")
		return false
	}

	return true
}

// Writes a message to the websocket
func (c *wsConnection) write(msg *operationMessage) {
	c.mu.Lock()
	c.conn.WriteJSON(msg)
	c.mu.Unlock()
}

// Handels messages and keeps web socket alive
func (c *wsConnection) run() {
	// Cancel ctx on function exit
	ctx, cancel := context.WithCancel(c.ctx)
	defer func() {
		cancel()
		c.close(websocket.CloseAbnormalClosure, "unexpected closure")
	}()
	// Make keep alive ticker
	if c.KeepAlivePingInterval != 0 {
		c.mu.Lock()
		c.keepAliveTicker = time.NewTicker(c.KeepAlivePingInterval)
		c.mu.Unlock()

		go c.keepAlive(ctx)
	}
	// Keep reading messages until terminate or error
	for {
		start := graphql.Now()
		message := c.readOp()
		if message == nil {
			return
		}
		// Match type
		switch message.Type {
		case startMsg:
			c.subscribe(start, message)
		case stopMsg:
			c.mu.Lock()
			closer := c.active[message.ID]
			c.mu.Unlock()
			if closer != nil {
				closer()
			}
		case connectionTerminateMsg:
			c.close(websocket.CloseNormalClosure, "terminated")
			return
		default:
			c.sendConnectionError("unexpected message %s", message.Type)
			c.close(websocket.CloseProtocolError, "unexpected message")
			return
		}
	}
}

// Writes keep alive messages to the web socket
func (c *wsConnection) keepAlive(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			c.keepAliveTicker.Stop()
			return
		case <-c.keepAliveTicker.C:
			c.write(&operationMessage{Type: connectionKeepAliveMsg})
		}
	}
}

// Handels a query
func (c *wsConnection) subscribe(start time.Time, message *operationMessage) {
	// Start trace
	ctx := graphql.StartOperationTrace(c.ctx)
	// Decode into raw parms
	var params *graphql.RawParams
	if err := jsonDecode(bytes.NewReader(message.Payload), &params); err != nil {
		c.sendError(message.ID, &gqlerror.Error{Message: "invalid json"})
		c.complete(message.ID)
		return
	}
	// Add read time to parms
	params.ReadTime = graphql.TraceTiming{
		Start: start,
		End:   graphql.Now(),
	}
	// Create ctx
	rc, err := c.exec.CreateOperationContext(ctx, params)
	if err != nil {
		c.sendResponse(message.ID, &graphql.Response{Errors: err})
		c.complete(message.ID)
		return
	}
	ctx = graphql.WithOperationContext(ctx, rc)
	// Create cancel function
	ctx, cancel := context.WithCancel(ctx)
	// Add to active list
	c.mu.Lock()
	c.active[message.ID] = cancel
	c.mu.Unlock()
	// Start operation
	go func() {
		// Catch and send err
		defer func() {
			if r := recover(); r != nil {
				userErr := rc.Recover(ctx, r)
				c.sendError(message.ID, &gqlerror.Error{Message: userErr.Error()})
			}
		}()
		// Run operation
		responses, ctx := c.exec.DispatchOperation(ctx, rc)
		for {
			response := responses(ctx)
			if response == nil {
				break
			}
			// Send response
			c.sendResponse(message.ID, response)
		}
		// Send complete message
		c.complete(message.ID)
		// Remove from active list
		c.mu.Lock()
		delete(c.active, message.ID)
		c.mu.Unlock()
		// Cancel context
		cancel()
	}()
}

// Encodes and writes a response
func (c *wsConnection) sendResponse(id string, response *graphql.Response) {
	b, err := json.Marshal(response)
	if err != nil {
		panic(err)
	}
	c.write(&operationMessage{
		Payload: b,
		ID:      id,
		Type:    dataMsg,
	})
}

// Encodes and writes a complete message
func (c *wsConnection) complete(id string) {
	c.write(&operationMessage{ID: id, Type: completeMsg})
}

// Encodes and writes a error
func (c *wsConnection) sendError(id string, errors ...*gqlerror.Error) {
	errs := make([]error, len(errors))
	for i, err := range errors {
		errs[i] = err
	}
	b, err := json.Marshal(errs)
	if err != nil {
		panic(err)
	}
	c.write(&operationMessage{Type: errorMsg, ID: id, Payload: b})
}

// Encodes and writes a connection error
func (c *wsConnection) sendConnectionError(format string, args ...interface{}) {
	b, err := json.Marshal(&gqlerror.Error{Message: fmt.Sprintf(format, args...)})
	if err != nil {
		panic(err)
	}

	c.write(&operationMessage{Type: connectionErrorMsg, Payload: b})
}

// Reads a operationMessage from the websocket
func (c *wsConnection) readOp() *operationMessage {
	_, r, err := c.conn.NextReader()
	if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseNoStatusReceived) {
		return nil
	} else if err != nil {
		c.sendConnectionError("invalid json: %T %s", err, err.Error())
		return nil
	}
	message := operationMessage{}
	if err := jsonDecode(r, &message); err != nil {
		c.sendConnectionError("invalid json")
		return nil
	}

	return &message
}

// Closes the web socket
func (c *wsConnection) close(closeCode int, message string) {
	c.mu.Lock()
	_ = c.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(closeCode, message))
	c.mu.Unlock()
	_ = c.conn.Close()
}

// Converts json to a object
func jsonDecode(r io.Reader, val interface{}) error {
	dec := json.NewDecoder(r)
	dec.UseNumber()
	return dec.Decode(val)
}
