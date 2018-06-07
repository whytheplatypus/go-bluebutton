package webdriver

import (
	"log"

	"github.com/tebeka/selenium"
)

type TokenFetcher struct {
	Jobs        chan [2]string
	WD          selenium.WebDriver
	RedirectURL string
}

func (tf *TokenFetcher) Work() error {
	log.Printf("[DEBUG] spinning up worker %s", tf.WD.SessionID())

	td := TokenDriver{tf.WD, tf.RedirectURL}

	for job := range tf.Jobs {
		log.Printf("[DEBUG] %s working on %v+ \n",
			tf.WD.SessionID(),
			job)
		err := td.Approve(job[0], job[1])
		if err != nil {
			tf.WD.Close()
			return err
		}
	}
	return nil
}
