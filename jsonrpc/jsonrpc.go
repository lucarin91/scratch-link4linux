/*
Simple implementation of JSONRPC 2.0.
*/
package jsonrpc

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sync/atomic"

	"github.com/rs/zerolog/log"
	"golang.org/x/net/websocket"
)

var MsgID int64 = 0

type Msg struct {
	ID      int64           `json:"id"`
	Jsonrpc string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
}

type Result struct {
	ID       int64       `json:"id"`
	Jsonrpc  string      `json:"jsonrpc"`
	Result   interface{} `json:"result"`
	Encoding string      `json:"encoding,omitempty"`
}

type Error struct {
	ID      int64  `json:"id"`
	Jsonrpc string `json:"jsonrpc"`
	Error   string `json:"error"`
}

func NewMsg(method string, params interface{}) Msg {
	buff, err := json.Marshal(params)
	if err != nil {
		panic(err)
	}
	return Msg{
		ID:      atomic.AddInt64(&MsgID, 1),
		Jsonrpc: "2.0",
		Method:  method,
		Params:  json.RawMessage(buff),
	}
}

func (m Msg) RespondBytes(buf []byte) Result {
	return Result{
		ID:       m.ID,
		Jsonrpc:  "2.0",
		Encoding: "base64",
		Result:   base64.StdEncoding.EncodeToString(buf),
	}
}

func (m Msg) Respond(data interface{}) Result {
	return Result{
		ID:      m.ID,
		Jsonrpc: "2.0",
		Result:  data,
	}
}

func (m Msg) Error(err string) Error {
	return Error{
		ID:      m.ID,
		Jsonrpc: "2.0",
		Error:   err,
	}
}

func (m Msg) DebugParams() map[string]interface{} {
	var out map[string]interface{}
	err := json.Unmarshal(m.Params, &out)
	if err != nil {
		panic(err)
	}
	return out
}

func WsSend[T Msg | Error | Result](c *websocket.Conn, data T) error {
	buff, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}

	_, err = c.Write(buff)
	if err != nil {
		return fmt.Errorf("ws write error: %w", err)
	}

	return nil
}

func WsReadLoop(c *websocket.Conn) <-chan Msg {
	out := make(chan Msg, 100)

	go func() {
		defer close(out)
		for {
			msg, err := wsRead(c)
			if errors.Is(err, io.EOF) {
				break
			}
			if err != nil {
				log.Warn().Err(err).Msg("read loop error, ignore")
				return
			}
			out <- msg
		}
	}()

	return out
}

func wsRead(c *websocket.Conn) (Msg, error) {
	buff := make([]byte, 512)
	var msg Msg
	for {
		n, err := c.Read(buff)
		if err != nil {
			return msg, fmt.Errorf("ws read: %w", err)
		}
		if n >= 512 {
			panic("too big")
		}

		err = json.Unmarshal(buff[:n], &msg)
		if err != nil {
			return msg, fmt.Errorf("ws read error: %w", err)
		}
		if len(msg.Method) == 0 {
			// ignore response messages
			continue
		}

		return msg, nil
	}
}
