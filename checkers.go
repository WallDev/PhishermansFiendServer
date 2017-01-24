package main

// All checkers in this file should be moved to be plugins after 1.8 release
// Plugins have an ability to load dynamically, this way they can be updated
// without ever restarting the main service.
// Also, as a thought, the actual checkers should be running as different
// microservice to reduce load on main server... But this is in a long run
// lets run with the MADNESSSSS

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/sirupsen/logrus"
	"github.com/texttheater/golang-levenshtein/levenshtein"
)

// Checker interface
type Checker interface {
	/* TODO: add methods */
	Runner()
	Result() bool
}

// Params holds the parameteres of the website to check received from the extension
type Params struct {
	URL      string
	PageHash string
}

// Check type
type Check struct {
	C      map[string]func([]byte, chan bool)
	P      *Params
	R      uint8
	ID     string
	result string
}

func FakeCheck(b []byte, c chan bool) {
	time.Sleep(5 * time.Second)
}

// LevenshteinCheck processes all the URLs found on page and if there is a url host which doesn't match href returns fail
func LevenshteinCheck(b []byte, c chan bool) {
	hrefs := getHrefs(b)
	for k, v := range hrefs {
		ku, err := url.Parse(k)
		if err != nil {
			continue
		}
		vu, err := url.Parse(v)
		if err != nil {
			continue
		}
		if vu.Host == "" || ku.Host == "" {
			logrus.WithFields(logrus.Fields{
				"href": ku.Host,
				"text": vu.Host,
			}).Debug("Skipping")
			continue
		}
		distance := levenshtein.DistanceForStrings([]rune(ku.Host), []rune(vu.Host), levenshtein.DefaultOptions)
		if distance > 0 {
			logrus.WithFields(logrus.Fields{"href": ku.Host, "text": vu.Host}).Debugf("Distance is %d", distance)
			c <- false
		}
	}
	time.Sleep(5 * time.Second)
	logrus.Debug("All good!")
	c <- true
}

// Result implementation for Checker interface
func (c *Check) Result() bool {
	if c.R == 0 || c.R == 7 {
		return true
	}
	return false
}

// Runner implementation for Checker interface
func (c *Check) Runner() {

	lenChecks := len(c.C)
	done := make(chan bool, lenChecks) // Might be a good idea to use GOMAXPROCS instead... but oh well, for now.

	res := map[bool]int{
		true:  0,
		false: 0,
	}
	running[c.ID] = true
	logrus.Debugf("Getting html for page %s", c.P.URL)
	// Get the HTML to pass it to plugins
	tree, err := fetchHTML(c.P.URL)
	if err != nil {
		c.result = err.Error()
		c.R = GREY
		goto CleanCheck
	}

	logrus.Debugf("Executing checks for url %s with hash %s\n", c.P.URL, c.P.PageHash)
	for k, v := range c.C {
		logrus.Debugf("Executing check %s", k)
		// Starting goroutine for all the checks
		go v(tree, done)
		v(tree, done)
	}
	// wait for channels and assign to result
	for i := 0; i < lenChecks; i++ {
		r := <-done
		logrus.Debug("Received channel")
		res[r] = res[r] + 1
	}

	// calculate result, currently the highest wins, might well be any other algorithm
	if res[true] > res[false] {
		c.R = WHITE
	} else if res[true] == res[false] {
		c.R = GREY
	} else {
		c.R = RED
	}

CleanCheck:
	logrus.Debugf("%+v", c)
	results[c.ID] = c
	delete(running, c.ID)
}

func getHrefs(body []byte) map[string]string {
	var result = make(map[string]string)
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return nil
	}
	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		val, ok := s.Attr("href")
		if ok {
			if s.Text() != "" {
				result[val] = s.Text()
			}
		}
	})
	return result
}

func fetchHTML(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		logrus.Error(err.Error())
		return nil, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		resp.Body.Close()
		logrus.Error(err.Error())
		return nil, err
	}
	return body, nil
}
