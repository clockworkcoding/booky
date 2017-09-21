package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/campoy/apiai"
)

func lookUpHandler(w http.ResponseWriter, r *http.Request) {

	var req apiai.Request

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "could not decode request: "+err.Error(), http.StatusBadRequest)
		return
	}

	log.Println("Entered look up handler", req)

	res := apiai.Response{
		Speech: fmt.Sprintf("Hello World!"),
	}
	b, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		log.Println("could not encode response: " + err.Error())
		return
	}
	fmt.Fprintf(w, "%s", b)
}
