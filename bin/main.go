package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"strings"

	bluebutton "github.com/whytheplatypus/go-bluebutton"

	"golang.org/x/oauth2"
)

var (
	RedirectURL   string
	bbURL         string
	CLIENT_ID     string
	CLIENT_SECRET string
)

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
	flag.StringVar(&RedirectURL,
		"redirect-url",
		"http://127.0.0.1:8080/testclient/callback",
		"The url for your oauth2 callback endpoint")
}

func main() {
	var verbose bool
	var server bool
	flag.BoolVar(&verbose, "v", false, "Enable for verbose logging")
	flag.BoolVar(&server, "s", false, "Enable for token callback server")
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

	if server {
		listenForToken(conf)
		return
	}

	var tknClient *http.Client

	tkn, err := ioutil.ReadFile(".tkn")
	if err != nil {
		log.Fatal(err)
	}

	tok := &oauth2.Token{}
	if err := json.Unmarshal(tkn, tok); err != nil {
		log.Fatal(err)
	}
	tknClient = conf.Client(context.Background(), tok)

	rsp, err := tknClient.Get(bbURL + "/v1/fhir/ExplanationOfBenefit/?patient=20140000008325")

	if err != nil {
		log.Fatal(err)
	}
	r, _ := httputil.DumpResponse(rsp, true)
	log.Println(string(r))

	patient, _ := ioutil.ReadAll(rsp.Body)
	var out bytes.Buffer
	json.Indent(&out, patient, "	", "\t")
}

func manageToken(tok *oauth2.Token) error {
	stok, err := json.MarshalIndent(tok, "", "	")
	if err != nil {
		log.Println(err)
		return err
	}

	if err := ioutil.WriteFile(".tkn", stok, 0777); err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func listenForToken(conf *oauth2.Config) {
	http.ListenAndServe(":8080", http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		if code := r.FormValue("code"); code != "" {
			tok, err := conf.Exchange(r.Context(), strings.TrimSpace(code))
			if err != nil {
				log.Println(err, tok)
				http.Error(
					rw,
					err.Error(),
					http.StatusInternalServerError,
				)
				return
			}
			if err := manageToken(tok); err != nil {
				log.Println(err)
				http.Error(
					rw,
					err.Error(),
					http.StatusInternalServerError,
				)
				return
			}

			return
		}
		authURL := conf.AuthCodeURL("state")
		http.Redirect(rw, r, authURL, http.StatusSeeOther)
	}))
}
