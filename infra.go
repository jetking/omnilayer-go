package omnilayer

import (
	"encoding/json"
	"fmt"
	"net/http"
)

const sendPostBufferSize = 100

type rawResponse struct {
	Result json.RawMessage `json:"result"`
	Error  *rpcError       `json:"error"`
}

func (r rawResponse) result() (result []byte, err error) {
	if r.Error != nil {
		return nil, r.Error
	}
	return r.Result, nil
}

type sendPostDetails struct {
	httpRequest *http.Request
	jsonRequest *jsonRequest
}

type jsonRequest struct {
	id           uint64
	cmd          command
	body         []byte
	responseChan chan *response
}

type response struct {
	result []byte
	err    error
}

func receive(resp chan *response) ([]byte, error) {
	r := <-resp
	return r.result, r.err
}

func newFutureError(err error) chan *response {
	responseChan := make(chan *response, 1)
	responseChan <- &response{err: err}
	return responseChan
}

type ConnConfig struct {
	Host                 string
	Endpoint             string
	User                 string
	Pass                 string
	Proxy                string
	ProxyUser            string
	ProxyPass            string
	Certificates         []byte
	DisableAutoReconnect bool
	DisableConnectOnNew  bool
	EnableBCInfoHacks    bool
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *rpcError) Error() string {
	return fmt.Sprintf("%d: %s", e.Code, e.Message)
}

type command interface {
	ID() string
	Method() string
	Params() []interface{}
}

func marshalCmd(cmd command) ([]byte, error) {
	rawCmd, err := newRpcRequest(cmd)
	if err != nil {
		return nil, err
	}

	return json.Marshal(rawCmd)
}

func newRpcRequest(cmd command) (*rpcRequest, error) {
	method := cmd.Method()
	params := cmd.Params()
	id := cmd.ID()

	rawParams := make([]json.RawMessage, len(params))

	for i := range params {
		msg, err := json.Marshal(params[i])
		if err != nil {
			return nil, err
		}

		rawParams[i] = json.RawMessage(msg)
	}

	return &rpcRequest{
		JsonRPC: "1.0",
		ID:      id,
		Method:  method,
		Params:  rawParams,
	}, nil
}

type rpcRequest struct {
	JsonRPC string            `json:"jsonrpc"`
	Method  string            `json:"method"`
	Params  []json.RawMessage `json:"params"`
	ID      string            `json:"id"`
}
