package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
)

func handleHit(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "success")
}

func handleCheckSite(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "success")
}

func handleFalsePositive(w http.ResponseWriter, r *http.Request) {
	host := r.FormValue("host")
	fmt.Printf("Website %s submitted for review", host)
	fmt.Fprint(w, "The website is submitted to review. Thank you for your cooperation.")
}

// handleWbl checks for White and Black lists entries and returns result to the caller
// if no result found then checks runner in started and check ID returned with the result
func handleWbl(w http.ResponseWriter, r *http.Request) {
	enc := json.NewEncoder(w)
	hostname := r.FormValue("hostname")
	fmt.Printf("hostname received: %s \n", hostname)

	if checkWL(hostname) {
		enc.Encode(Resp{Success: true, Reason: WL, URL: hostname})
		return
	} else if checkBL(hostname) {
		enc.Encode(Resp{Success: false, Reason: BL})
		return
	}
	p := new(Params)
	p.URL = r.FormValue("url")
	p.PageHash = r.FormValue("hash")
	c := Check{P: p, ID: randString(8)}
	checks := make(map[string]func([]byte, chan bool))
	checks["levenshtein"] = LevenshteinCheck
	c.C = checks
	running[c.ID] = true
	go c.Runner()
	enc.Encode(Resp{Success: false, Reason: CHKRUN, CheckID: c.ID})
}

// handleGetResult checking for the result of given checks if it available
// write error to response if it not available and 'success' = false
func handleGetResult(w http.ResponseWriter, r *http.Request) {
	enc := json.NewEncoder(w)
	id := r.FormValue("cID") // get cID from param
	if running[id] {
		enc.Encode(Resp{Success: false, Reason: PENDING})
		return
	}
	if v, ok := results[id]; ok {
		enc.Encode(Resp{Success: v.Result(), Reason: v.R})
		delete(running, id)
		return
	}
	enc.Encode(Resp{Success: false, Reason: NOTFOUND})
}

// Shows explanation page for blocked resource
func handleErrorPage(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("danger.tmpl")
	if err != nil {
		log.Fatal(err.Error())
	}
	c := struct {
		Host string
	}{
		Host: r.FormValue("host"),
	}
	t.Execute(w, c)
}
