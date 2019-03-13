package main

import (
	"github.com/sirupsen/logrus"
	//"github.com/texttheater/golang-levenshtein/levenshtein"

	"log"
	"math/rand"
	"net/http"
	"regexp"
	"time"
)

// Errors for result return
const (
	WL       = iota // 0 Whitelisted site
	BL              // 1 Blacklisted site
	CHKRUN          // 2 Starting check, this must be in pair with CheckID
	PENDING         // 3 Pending checks to finish
	NOTFOUND        // 4 Checkrun with given ID not found
	RED             // 5 Dangereous website
	GREY            // 6 Might be dangerous website
	WHITE           // 7 Safe website
)


// Resp Server response
type Resp struct {
	Success bool   `json:"success"`
	Reason  uint8  `json:"reason,omitempty"`
	CheckID string `json:"checkID,omitempty"`
	URL     string `json:"host,omitempty"`
}


// TODO: Blacklist and whitelist should be somewhere in inmemory storage instead of here so it could be easily updated without recompiling
var bl = []*regexp.Regexp{
	regexp.MustCompile("gooogle.com"),
}

var wl = []*regexp.Regexp{
	regexp.MustCompile("google\\.com\\S{0,3}$"),
	regexp.MustCompile("facebook.com$"),
	regexp.MustCompile("localhost$"),
	regexp.MustCompile("chrome.com$"),
}

var results = make(map[string]*Check)
var running = make(map[string]bool)

func randString(n int) string {
	var src = rand.NewSource(time.Now().UnixNano())
	b := make([]byte, n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}
	return string(b)
}

func checkWL(host string) bool {
	for _, v := range wl {
		if ok := v.MatchString(host); ok {
			return true
		}
	}
	return false
}

func checkBL(host string) bool {
	for _, v := range bl {
		if ok := v.MatchString(host); ok {
			return true
		}
	}
	return false
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)



func main() {
	// Levenshteing check as proof of concept
	//fmt.Println(levenshtein.DistanceForStrings([]rune("google.com"), []rune("goоgle.com"), levenshtein.DefaultOptions))
	logrus.SetLevel(logrus.DebugLevel)
	logrus.Debug("Starting server")
	http.HandleFunc("/check", handleCheckSite)
	http.HandleFunc("/false", handleFalsePositive)
	http.HandleFunc("/hit", handleHit)
	http.HandleFunc("/wbl", handleWbl)
	http.HandleFunc("/getCheck", handleGetResult)
	http.HandleFunc("/phishingPage", handleErrorPage)
	//log.Fatal(http.ListenAndServe(":8080", nil))
	// Cant check with self signed cert since browser deny it
	log.Fatal(http.ListenAndServeTLS(":8080", "server.crt", "server.key", nil))
}
