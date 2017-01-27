# Intro
This is server for PhishermansFiend chrome extension.

Research about the phishing can be found here https://walldev.github.io/PhishermansFiendExtension/

The server is part of http://github.com/WallDev/PhisermansFiendExtension project.
Written on golang and prepared for go 1.8 with it's plugin support the server is very fast and reliable.
It can run lots of checks in paralel at very low overhead cost as goroutines are very cheap.

## Content
* `main.go` - an entry point to application. It sets up handlers and waiting for connection.
* `handlers.go` - a handlers collection, provides views to the application
* `checkers.go` - checks conection and utilities functions. After go 1.8 release should be separated to plugins and helpers. Currently 2 checks exists, one levenshtein distance counter to find malicious links and the other dummy checker which hangs for random time to simulate lag and return random result for debugging.
* `danger.tmpl` - a template for page that shown when extension redirects due to phishing website

## How it works
1. Before browser makes a request to the domain that user asked for, extension sends request to the server and it checks the domain with whitelist and if not found in blacklist.
2. If the domain is not whitelisted or blacklisted - starts checks on the website that the user requested and responding to extension to load the website.
3. Extension unblock the original request and allows browser to load the website while disabling all input method on the page to not to allow user to input data into non-checked website
4. Extension polling the server for when checks finish. As soon as server returns response with final result the extension takes corresponding actions:
  1. If the website is marked as WHITE then all the inputs are re-enabled and user allowed to proceed
  2. If the website is marked as GREY then the extension allows user to use the site but inverting all the colors of the page and showing topbar explaining that this website might be dangerous and advices to proceed with care. This notification is clickable and click on it will revert everything to normal state.
  3. If the website is marked as RED then extension redirects user to special page explaining the website is dangerous and blocked to use.

## Notes
* There is no need in separate API to report hit. As while server running the check it has the url of the website hitted. According to the result server itself can add hits or misses to database. Currently implemented as simple log to stout
* Current server setup is for testing and debugging. 
For whitelisting it has only:
```
google.com
facebook.com
localhost
chrome.com

```
For blacklistion only `gooogle.com` is available. Every other website is checked with random results and might be marked as any severity.
* While "danger" page is featuring "report" of wrongly marked red website - it does not implement it as entire application is not connected to database at this moment.

## Extension
For extension documentation please look [here](https://github.com/WallDev/PhishermansFiendExtension)
