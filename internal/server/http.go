package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

func NewHTTPServer(addr string) *http.Server {
	httpsrv := newHTTPServer()
	r := mux.NewRouter()
	r.HandleFunc("/", httpsrv.handleProduce).Methods("POST")
	r.HandleFunc("/", httpsrv.handleConsume).Methods("GET")
	// return pointer to Server, user will call cmd/server/ListenAndServe() to
	// listen for and handle incoming requests
	return &http.Server{
		Addr:    addr,
		Handler: r,
	}
}

// server state consists of a Log:
//
//	type Log struct {
//	  mu      sync.Mutex
//	  records []Record
//	}
type httpServer struct {
	Log *Log
}

func newHTTPServer() *httpServer {
	return &httpServer{
		Log: NewLog(),
	}
}

type ProduceRequest struct {
	Record Record `json:"record"`
}
type ProduceResponse struct {
	Offset uint64 `json:"offset"`
}

type ConsumeRequest struct {
	Offset uint64 `json:"offset"`
}
type ConsumeResponse struct {
	Record Record `json:"record"`
}

// Produce handler: write to the log
func (s *httpServer) handleProduce(w http.ResponseWriter, r *http.Request) {
	var request ProduceRequest
	// unmarshal the request into a struct
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	fmt.Printf("Server: handle Produce request, request: %v\n", request)

	// append record to server state - Log
	offset, err := s.Log.Append(request.Record)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// put offset into response, encode response
	response := ProduceResponse{Offset: offset}
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Printf("  responded with offset: %d\n\n", offset)
}

// Consume handler: read from the log
func (s *httpServer) handleConsume(w http.ResponseWriter, r *http.Request) {
	var request ConsumeRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// read record
	fmt.Printf("Server: handle Consume request w/ offset: %d\n", request.Offset)
	record, err := s.Log.Read(request.Offset)
	if err == ErrOffsetNotFound {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	response := ConsumeResponse{Record: record}
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Printf("  responded with:%v\n\n", response)
}
