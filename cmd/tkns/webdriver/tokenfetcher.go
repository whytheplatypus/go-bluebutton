package webdriver

import (
	"fmt"
	"time"

	"github.com/tebeka/selenium"
	"github.com/tebeka/selenium/firefox"
)

func FetchToken(cg CredGenerator, RedirectURL string) error {
	var portNum int

	Retry(5, time.Second/2, func() error {
		var err error
		portNum, err = port()
		if err != nil {
			return err
		}
		return nil
	})

	var opts []selenium.ServiceOption
	s, err := selenium.NewGeckoDriverService("geckodriver", portNum, opts...)
	if err != nil {
		return err
	}
	defer s.Stop()
	// Connect to the WebDriver instance running locally.
	caps := selenium.Capabilities{"browserName": "firefox"}
	caps.AddFirefox(firefox.Capabilities{
		Args: []string{"-headless"},
	})
	wd, err := selenium.NewRemote(caps, fmt.Sprintf("http://localhost:%d", portNum))
	if err != nil {
		return err
	}
	defer wd.Quit()
	td := TokenDriver{wd, RedirectURL}
	return td.Approve(cg.GetCreds())
}

func Retry(tries int, sleep time.Duration, fn func() error) error {
	if err := fn(); err != nil {
		if s, ok := err.(stop); ok {
			return s.error
		}

		if tries--; tries > 0 {
			time.Sleep(sleep)
			return Retry(tries, 2*sleep, fn)
		}
		return err
	}
	return nil
}

type stop struct {
	error
}

type CredGenerator interface {
	GetCreds() (username, password string)
}
