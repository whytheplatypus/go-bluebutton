package webdriver

import (
	"log"
	"net"
	"strings"

	"github.com/tebeka/selenium"
)

type TokenDriver struct {
	WD          selenium.WebDriver
	RedirectURL string
}

func (td *TokenDriver) Approve(username, password string) error {
	log.Println(username, password)
	wd := td.WD
	// Navigate to the simple playground interface.
	if err := wd.Get(td.RedirectURL); err != nil {
		return err
	}

	if err := wd.Wait(func(wd selenium.WebDriver) (bool, error) {
		elem, err := wd.FindElement(selenium.ByCSSSelector, "#SWEUserName")
		if err != nil {
			return false, err
		}
		if err := elem.SendKeys(username); err != nil {
			return false, err
		}

		elem, err = wd.FindElement(selenium.ByCSSSelector, "#SWEPassword")
		if err != nil {
			return false, err
		}
		if err := elem.SendKeys(password); err != nil {
			return false, err
		}

		elem, err = wd.FindElement(selenium.ByCSSSelector, "#SignIn")
		if err != nil {
			return false, err
		}

		if err := elem.Click(); err != nil {
			return false, err
		}
		return true, nil
	}); err != nil {
		return err
	}
	if err := wd.Wait(func(wd selenium.WebDriver) (bool, error) {
		elem, err := wd.FindElement(selenium.ByCSSSelector, "#approve")
		if err != nil {
			return false, err
		}
		if err := elem.Click(); err != nil {
			return false, err
		}
		return true, nil
	}); err != nil {
		return err
	}
	return wd.Wait(func(wd selenium.WebDriver) (bool, error) {
		currentURL, err := wd.CurrentURL()
		return strings.HasPrefix(currentURL, td.RedirectURL), err
	})
}

func port() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}
	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}
