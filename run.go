package ngrok_runner

import (
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
	"regexp"
	"time"
)

// GetNgrokURL first checks to see if ngrok is already running, if so, it will get the ngork url and return it. If not,
// it will first start ngrok, then get the url and return it. There is a timeout on starting ngrok of 10 seconds. With
// a slow internet connection, you may need this to be longer.
func StartNgrok(port string) (string, error) {
	log.Println("starting ngrok...")
	resp, err := http.Get("http://localhost:4040/api/tunnels")
	if err != nil {
		done := make(chan bool)
		go startNgrok(port, done)

		timeout := make(chan bool)
		go func() {
			time.Sleep(10 * time.Second)
			timeout <- true
		}()
		select {
		case <-done:
		case <-timeout:
			return "", errors.New("timeout when attempting to get ngrok URL (just try again, it might work this time)")
		}

		resp, err = http.Get("http://localhost:4040/api/tunnels")
		if err != nil {
			return "", err
		}
	}
	defer resp.Body.Close()
	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	re := regexp.MustCompile(`https:\/\/[a-z0-9]*.ngrok.io`)
	ngrokURL := re.FindString(string(content))
	if ngrokURL == "" {
		log.Println(string(content))
		return "", errors.New("problem getting ngork url (maybe there is already something running on port 4040 besides ngrok?)")
	}
	log.Println("ngrok started")
	return ngrokURL, nil
}

func startNgrok(port string, done chan bool) {
	command := exec.Command("ngrok", "http", port)
	command.Start()
	command.Wait()
	done <- true
}
