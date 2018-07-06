/*
Package streaming implements a client for the Luno Streaming API.

Example:

	c, err := streaming.Dial(keyID, keySecret, "XBTZAR")
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()

	for {
		seq, bids, asks := c.OrderBookSnapshot()
		log.Printf("%d: %v %v\n", seq, bids[0], asks[0])
		time.Sleep(time.Minute)
	}
*/
package streaming

import (
	"encoding/json"
	"errors"
	"flag"
	"log"
	"math/rand"
	"sync"
	"time"

	"golang.org/x/net/websocket"
)

type connection struct {
	keyID, keySecret string
	pair             string

	ws     *websocket.Conn
	closed bool

	MessageProcessor messageProcessor

	mu sync.Mutex
}

// Dial initiates a connection to the streaming service and starts processing
// data for the given market pair.
// The connection will automatically reconnect on error.
func Dial(keyID, keySecret, pair string, opts ...DialOption) (*connection, error) {
	if keyID == "" || keySecret == "" {
		return nil, errors.New("streaming: streaming API requires credentials")
	}

	c := &connection{
		keyID:     keyID,
		keySecret: keySecret,
		pair:      pair,
	}
	for _, opt := range opts {
		opt(c)
	}

	go c.manageForever()
	return c, nil
}

var wsHost = flag.String(
	"luno_websocket_host", "wss://ws.luno.com", "Luno API websocket host")

func (c *connection) manageForever() {
	attempts := 0
	var lastAttempt time.Time
	for {
		c.mu.Lock()
		closed := c.closed
		c.mu.Unlock()
		if closed {
			return
		}

		lastAttempt = time.Now()
		attempts++
		if err := c.connect(); err != nil {
			log.Printf("luno/streaming: connection error key=%s pair=%s: %v",
				c.keyID, c.pair, err)
		}

		if time.Now().Sub(lastAttempt) > time.Hour {
			attempts = 0
		}
		if attempts > 5 {
			attempts = 5
		}
		wait := 5
		for i := 0; i < attempts; i++ {
			wait = 2 * wait
		}
		wait = wait + rand.Intn(wait)
		dt := time.Duration(wait) * time.Second
		log.Printf("luno/streaming: Waiting %s before reconnecting", dt)
		time.Sleep(dt)
	}
}

func (c *connection) connect() error {
	url := *wsHost + "/api/1/stream/" + c.pair
	ws, err := websocket.Dial(url, "", "http://localhost/")
	if err != nil {
		return err
	}

	defer func() {
		ws.Close()
		c.mu.Lock()
		c.ws = nil
		c.MessageProcessor.Reset()
		c.mu.Unlock()
	}()

	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return nil
	} else {
		c.ws = ws
		c.mu.Unlock()
	}

	cred := credentials{c.keyID, c.keySecret}
	if err := websocket.JSON.Send(ws, cred); err != nil {
		return err
	}

	log.Printf("luno/streaming: connection established key=%s pair=%s",
		c.keyID, c.pair)

	go sendPings(ws)

	for {
		var data []byte
		err := websocket.Message.Receive(c.ws, &data)
		if err != nil {
			return err
		}

		err = c.handleMessage(data)
		if err != nil {
			return err
		}
	}
}

func (c *connection) handleMessage(message []byte) error {
	if string(message) == "\"\"" {
		return nil
	}

	var ob orderBook
	if err := json.Unmarshal(message, &ob); err != nil {
		return err
	}
	if ob.Asks != nil || ob.Bids != nil {
		if err := c.MessageProcessor.receivedOrderBook(ob); err != nil {
			return err
		}
		return nil
	}

	var u Update
	if err := json.Unmarshal(message, &u); err != nil {
		return err
	}
	if err := c.MessageProcessor.receivedUpdate(u); err != nil {
		return err
	}

	return nil
}

func sendPings(ws *websocket.Conn) {
	defer ws.Close()
	for {
		if !sendPing(ws) {
			return
		}
		time.Sleep(time.Minute)
	}
}

func sendPing(ws *websocket.Conn) bool {
	return websocket.Message.Send(ws, "") == nil
}

// Close the connection.
func (c *connection) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.closed = true
	if c.ws != nil {
		c.ws.Close()
	}
}
