package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"sync"
	"time"

	"golang.org/x/oauth2"

	bluebutton "github.com/whytheplatypus/go-bluebutton"
	"github.com/whytheplatypus/go-bluebutton/cmd/tkns/server"
	"github.com/whytheplatypus/go-bluebutton/cmd/tkns/webdriver"
)

var (
	RedirectURL   string = "http://localhost:8080/"
	bbURL         string
	CLIENT_ID     string
	CLIENT_SECRET string
)

const BENE_MIN = 0
const BENE_MAX = 30000
const NUM_WORKERS = 4

func init() {
	flag.StringVar(&bbURL,
		"url",
		"http://localhost:8000",
		"The URL of the bluebutton API")
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
	var verbose bool
	var tknCount int
	flag.BoolVar(&verbose, "v", false, "Enable for verbose logging")
	flag.IntVar(&tknCount, "n", 1, "The number of tokens to generate")
	flag.Parse()
	if verbose {
		log.SetFlags(log.Lshortfile | log.LstdFlags)
	} else {
		log.SetOutput(ioutil.Discard)
	}

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

	tknChan := make(chan Token)
	defer close(tknChan)

	// Fans out
	go http.ListenAndServe(":8080", server.TokenHandler(conf, func(tkn *oauth2.Token) error {
		tknChan <- Token{*tkn}
		return nil
	}))
	// Funnel in
	go WriteTkn(os.Stdout, tknChan)

	generator := &RandomCred{}

	creds := make(chan struct{})
	// worker pool
	wg := &sync.WaitGroup{}
	wg.Add(NUM_WORKERS)
	for i := 0; i < NUM_WORKERS; i++ {
		go func(creds <-chan struct{}) {
			defer wg.Done()
			for range creds {
				if err := webdriver.FetchToken(generator, RedirectURL); err != nil {
					log.Fatal(err)
				}
			}
		}(creds)
	}

	for i := 0; i < tknCount; i++ {
		creds <- struct{}{}
	}
	close(creds)
	wg.Wait()
}

func WriteTkn(w io.Writer, tknChan <-chan Token) {
	for tkn := range tknChan {
		fmt.Fprintln(w, &tkn)
	}
}

type Token struct {
	oauth2.Token
}

func (tok *Token) String() string {
	//stok, err := json.MarshalIndent(tok, "", "	")
	stok, err := json.Marshal(tok)
	if err != nil {
		log.Println(err)
		return ""
	}

	return string(stok)
}

type RandomCred struct{}

func (rc *RandomCred) GetCreds() (username, password string) {
	u := rand.Intn(BENE_MAX)
	username = fmt.Sprintf("BBUser%05d", u)
	password = fmt.Sprintf("PW%05d!", u)
	return
}
