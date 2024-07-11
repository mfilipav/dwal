package main

import (
	"fmt"
	"log"

	"github.com/mfilipav/dwal/internal/server"
)

var PORT int = 8080

func main() {
	fmt.Printf("Server listening at port: %d\n", PORT)
	srv := server.NewHTTPServer(
		fmt.Sprintf(":%d", PORT))
	log.Fatal(srv.ListenAndServe())
}
