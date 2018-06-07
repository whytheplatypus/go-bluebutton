package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"time"

	"golang.org/x/oauth2"

	"github.com/tebeka/selenium"
	"github.com/tebeka/selenium/firefox"
	"github.com/whytheplatypus/errgroup"
	bluebutton "github.com/whytheplatypus/go-bluebutton"
	"github.com/whytheplatypus/go-bluebutton/cmd/tkns/server"
	"github.com/whytheplatypus/go-bluebutton/cmd/tkns/webdriver"
)

var (
	RedirectURL   string
	bbURL         string
	CLIENT_ID     string
	CLIENT_SECRET string
)

const BENE_MIN = 0
const BENE_MAX = 30000

func init() {
	flag.StringVar(&bbURL,
		"url",
		"http://localhost:8000",
		"The URL of the bluebutton API")
	flag.StringVar(&RedirectURL,
		"redirect-uri",
		"http://localhost:8080/",
		"The URI for the redirect to the local server")
	flag.StringVar(&CLIENT_SECRET,
		"secret",
		"",
		"The oauth2 client secret for your bluebutton application")
	flag.StringVar(&CLIENT_ID,
		"id",
		"",
		"The oauth2 client id for your bluebutton application")
	rand.Seed(time.Now().UnixNano())
}

func main() {
	var verbose, random bool
	var tknCount int
	var numWorkers int
	var driverStartPort int
	flag.BoolVar(&random, "random", false, "Should we generate random beneficiary credentials?")
	flag.BoolVar(&verbose, "v", false, "Enable for verbose logging")
	flag.IntVar(&tknCount, "n", 1, "The number of tokens to generate")
	flag.IntVar(&numWorkers, "w", 1, "The number of workers to use")
	flag.IntVar(&driverStartPort, "p", 4444, "The start port for geckodriver, starting with this port a port will be assigned to each worker process")
	flag.Parse()
	if verbose {
		log.SetFlags(log.Lshortfile | log.LstdFlags)
		log.SetOutput(os.Stderr)
	} else {
		log.SetOutput(ioutil.Discard)
	}

	// Time how long this takes
	// -------------------------
	start := time.Now()

	defer func() {
		elapsed := time.Since(start)
		log.Printf("Generated %d tokens in %s \n", tknCount, elapsed)
	}()

	// Spinup webserver to fetch tokens
	// ---------------------------------
	bluebutton.BB_URL = bbURL
	conf := &oauth2.Config{
		RedirectURL:  RedirectURL,
		ClientID:     CLIENT_ID,
		ClientSecret: CLIENT_SECRET,
		Endpoint: oauth2.Endpoint{
			AuthURL:  bluebutton.AuthURL(),
			TokenURL: bluebutton.TokenURL(),
		},
	}
	handler := server.TokenHandler(conf, server.WriteTkn(os.Stdout))
	go func(handler http.Handler) {
		url, err := url.Parse(RedirectURL)
		if err != nil {
			log.Fatal(err)
		}
		if err := http.ListenAndServe(fmt.Sprintf(":%s", url.Port()), handler); err != nil {
			log.Fatal(err)
		}
	}(handler)

	// Spin up token generation jobs
	// ------------------------------
	jobs := generateCredentials(tknCount, random)

	var g errgroup.Group
	for i := 0; i < numWorkers; i++ {
		s, wd := buildWebServiceAndDriver(driverStartPort + i)
		defer s.Stop()
		defer wd.Close()
		tf := &webdriver.TokenFetcher{
			Jobs:        jobs,
			WD:          wd,
			RedirectURL: RedirectURL,
		}
		g.Go(tf.Work)
	}

	if err := g.Wait(); err != nil {
		log.Println("[ERROR]", err)
		os.Exit(1)
	}
}

func generateCredentials(tknCount int, random bool) chan [2]string {
	jobs := make(chan [2]string, tknCount)
	defer close(jobs)

	var generator Cred
	if random == true {
		generator = &RandomCred{}
	} else {
		generator = &SerialCred{}
	}

	for i := 0; i < tknCount; i++ {
		u, p := generator.GetCreds()
		jobs <- [2]string{u, p}
	}
	return jobs
}

func buildWebServiceAndDriver(portNum int) (*selenium.Service, selenium.WebDriver) {
	var opts []selenium.ServiceOption
	s, err := selenium.NewGeckoDriverService("geckodriver", portNum, opts...)
	if err != nil {
		log.Fatal(err)
	}
	// Connect to the WebDriver instance running locally.
	caps := selenium.Capabilities{"browserName": "firefox"}
	caps.AddFirefox(firefox.Capabilities{
		Args: []string{"-headless"},
	})
	wd, err := selenium.NewRemote(caps, fmt.Sprintf("http://localhost:%d", portNum))
	if err != nil {
		log.Fatal(err)
	}
	return s, wd
}

type Cred interface {
	GetCreds() (username, password string)
}

type SerialCred struct {
	Idx int
}

func (sc *SerialCred) GetCreds() (username, password string) {
	username = fmt.Sprintf("User%05d", sc.Idx)
	password = fmt.Sprintf("PW%05d!", sc.Idx)
	sc.Idx++
	return
}

type RandomCred struct{}

func (rc *RandomCred) GetCreds() (username, password string) {
	u := rand.Intn(BENE_MAX)
	username = fmt.Sprintf("BBUser%05d", u)
	password = fmt.Sprintf("PW%05d!", u)
	return
}
