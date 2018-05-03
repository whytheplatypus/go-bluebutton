package webdriver

import (
	"fmt"

	"github.com/tebeka/selenium"
	"github.com/tebeka/selenium/firefox"
)

func FetchToken(cg CredGenerator, RedirectURL string) error {
	port, err := port()
	if err != nil {
		return err
	}

	var opts []selenium.ServiceOption
	s, err := selenium.NewGeckoDriverService("geckodriver", port, opts...)
	if err != nil {
		return err
	}
	defer s.Stop()
	// Connect to the WebDriver instance running locally.
	caps := selenium.Capabilities{"browserName": "firefox"}
	caps.AddFirefox(firefox.Capabilities{
		Args: []string{"-headless"},
	})
	wd, err := selenium.NewRemote(caps, fmt.Sprintf("http://localhost:%d", port))
	if err != nil {
		return err
	}
	defer wd.Quit()
	td := TokenDriver{wd, RedirectURL}
	return td.Approve(cg.GetCreds())
}

type CredGenerator interface {
	GetCreds() (username, password string)
}
