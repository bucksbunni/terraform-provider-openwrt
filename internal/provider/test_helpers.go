package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

type rpcHandler func(*rpcRequest) (interface{}, *rpcError)

type TestRPCServer struct {
	Server    *httptest.Server
	Mux       *http.ServeMux
	AuthToken string
}

func NewTestRPCServer(t *testing.T) *TestRPCServer {
	t.Helper()
	mux := http.NewServeMux()
	authToken := "test-token-123"

	mux.HandleFunc("/cgi-bin/luci/rpc/auth", func(w http.ResponseWriter, r *http.Request) {
		resp := rpcResponse{
			ID:     1,
			Result: json.RawMessage(`"` + authToken + `"`),
			Error:  nil,
		}
		_ = json.NewEncoder(w).Encode(&resp)
	})

	return &TestRPCServer{
		Mux:       mux,
		AuthToken: authToken,
		Server:    httptest.NewServer(mux),
	}
}

func (s *TestRPCServer) Close() {
	s.Server.Close()
}

func (s *TestRPCServer) checkAuth(r *http.Request) bool {
	token := r.URL.Query().Get("auth")
	return token == s.AuthToken
}

func (s *TestRPCServer) AddUCIHandler(method string, h rpcHandler) {
	s.Mux.HandleFunc("/cgi-bin/luci/rpc/uci", s.makeUCIHandler(method, h))
}

func (s *TestRPCServer) AddUCIGenericHandler(h func(method string, req *rpcRequest) (interface{}, *rpcError)) {
	s.Mux.HandleFunc("/cgi-bin/luci/rpc/uci", func(w http.ResponseWriter, r *http.Request) {
		if !s.checkAuth(r) {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		var req rpcRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		result, rpcErr := h(req.Method, &req)
		s.writeRPC(w, req.ID, result, rpcErr)
	})
}

func (s *TestRPCServer) AddFSHandler(method string, fn func(*rpcRequest) (interface{}, error)) {
	s.Mux.HandleFunc("/cgi-bin/luci/rpc/fs", s.makeFSHandler(method, fn))
}

func (s *TestRPCServer) AddFSGenericHandler(h func(method string, req *rpcRequest) (interface{}, error)) {
	s.Mux.HandleFunc("/cgi-bin/luci/rpc/fs", func(w http.ResponseWriter, r *http.Request) {
		if !s.checkAuth(r) {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		var req rpcRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		result, err := h(req.Method, &req)
		if err != nil {
			s.writeRPC(w, req.ID, nil, &rpcError{Code: -1, Message: err.Error()})
			return
		}
		s.writeRPC(w, req.ID, result, nil)
	})
}

func (s *TestRPCServer) makeFSHandler(method string, fn func(*rpcRequest) (interface{}, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !s.checkAuth(r) {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		var req rpcRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if req.Method != method {
			http.Error(w, "unexpected method: "+req.Method, http.StatusBadRequest)
			return
		}
		result, err := fn(&req)
		if err != nil {
			s.writeRPC(w, req.ID, nil, &rpcError{Code: -1, Message: err.Error()})
			return
		}
		s.writeRPC(w, req.ID, result, nil)
	}
}

func (s *TestRPCServer) AddIPKGHandler(method string, fn func(*rpcRequest) error) {
	s.Mux.HandleFunc("/cgi-bin/luci/rpc/ipkg", s.makeRPCHandler(method, func(req *rpcRequest) (interface{}, *rpcError) {
		if err := fn(req); err != nil {
			return nil, &rpcError{Code: -1, Message: err.Error()}
		}
		return true, nil
	}))
}

func (s *TestRPCServer) AddSysHandler(method string, fn func(params []interface{}) (interface{}, error)) {
	s.Mux.HandleFunc("/cgi-bin/luci/rpc/sys", func(w http.ResponseWriter, r *http.Request) {
		if !s.checkAuth(r) {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		var req rpcRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if req.Method != method {
			http.Error(w, "unexpected method: "+req.Method, http.StatusBadRequest)
			return
		}
		result, err := fn(req.Params)
		if err != nil {
			s.writeRPC(w, req.ID, nil, &rpcError{Code: -1, Message: err.Error()})
			return
		}
		s.writeRPC(w, req.ID, result, nil)
	})
}

func (s *TestRPCServer) AddSysGenericHandler(h func(method string, params []interface{}) (interface{}, error)) {
	s.Mux.HandleFunc("/cgi-bin/luci/rpc/sys", func(w http.ResponseWriter, r *http.Request) {
		if !s.checkAuth(r) {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		var req rpcRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		result, err := h(req.Method, req.Params)
		if err != nil {
			s.writeRPC(w, req.ID, nil, &rpcError{Code: -1, Message: err.Error()})
			return
		}
		s.writeRPC(w, req.ID, result, nil)
	})
}

func (s *TestRPCServer) makeUCIHandler(method string, h rpcHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !s.checkAuth(r) {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		var req rpcRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if req.Method != method {
			http.Error(w, "unexpected method: "+req.Method, http.StatusBadRequest)
			return
		}
		result, rpcErr := h(&req)
		s.writeRPC(w, req.ID, result, rpcErr)
	}
}

func (s *TestRPCServer) makeRPCHandler(method string, h rpcHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !s.checkAuth(r) {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		var req rpcRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if req.Method != method {
			http.Error(w, "unexpected method: "+req.Method, http.StatusBadRequest)
			return
		}
		result, rpcErr := h(&req)
		s.writeRPC(w, req.ID, result, rpcErr)
	}
}

func (s *TestRPCServer) writeRPC(w http.ResponseWriter, id int64, result interface{}, rpcErr *rpcError) {
	if rpcErr != nil {
		resp := rpcResponse{
			ID:     id,
			Result: nil,
			Error:  rpcErr,
		}
		_ = json.NewEncoder(w).Encode(&resp)
		return
	}
	if result == nil {
		resp := rpcResponse{
			ID:     id,
			Result: json.RawMessage("null"),
			Error:  nil,
		}
		_ = json.NewEncoder(w).Encode(&resp)
		return
	}
	data, err := json.Marshal(result)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	resp := rpcResponse{
		ID:     id,
		Result: data,
		Error:  nil,
	}
	_ = json.NewEncoder(w).Encode(&resp)
}

func (s *TestRPCServer) NewClient(t *testing.T) *JsonRpcClient {
	t.Helper()
	client, err := NewJsonRpcClient(context.Background(), JsonRpcConfig{
		BaseURL:  s.Server.URL,
		Username: "root",
		Password: "password",
	})
	require.NoError(t, err)
	return client
}
