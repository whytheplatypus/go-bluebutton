package bluebutton_test

import (
	"context"
	"encoding/json"
	"flag"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"
	"testing"

	"golang.org/x/oauth2"

	bluebutton "github.com/whytheplatypus/go-bluebutton"
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

func TestMain(m *testing.M) {
	flag.Parse()
	os.Exit(m.Run())
}

func TestExchangeWithBadCode(t *testing.T) {
	code := "bogus"
	conf := &oauth2.Config{
		RedirectURL:  RedirectURL,
		ClientID:     CLIENT_ID,
		ClientSecret: CLIENT_SECRET,
		Endpoint: oauth2.Endpoint{
			AuthURL:  bbURL + "/v1/o/authorize/",
			TokenURL: bbURL + "/v1/o/token/",
		},
	}
	_, err := conf.Exchange(context.Background(), code)
	t.Log(err)
	if strings.Contains(err.Error(), "500 Internal Server Error") {
		t.Error("500")
	}
}

func TestPasswordCredentialsToken(t *testing.T) {
	conf := &oauth2.Config{
		RedirectURL:  RedirectURL,
		ClientID:     CLIENT_ID,
		ClientSecret: CLIENT_SECRET,
		Endpoint: oauth2.Endpoint{
			AuthURL:  bbURL + "/v1/o/authorize/",
			TokenURL: bbURL + "/v1/o/token/",
		},
	}

	tok, err := conf.PasswordCredentialsToken(context.Background(), "blah", "blahblah")
	if strings.Contains(err.Error(), "500 Internal Server Error") {
		t.Error("Exhange Error")
	}
	t.Log(tok)
}

func BenchmarkEOB(b *testing.B) {
	for i := 0; i < b.N; i++ {
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
		tkn, err := ioutil.ReadFile(".tkn")
		if err != nil {
			b.Fatal(err)
		}

		tok := &oauth2.Token{}
		if err := json.Unmarshal(tkn, tok); err != nil {
			b.Fatal(err)
		}
		tknClient := conf.Client(context.Background(), tok)

		rsp, err := tknClient.Get(bbURL + "/v1/fhir/ExplanationOfBenefit/")
		//rsp, err := tknClient.Get(bbURL + "/v1/fhir/Patient/20140000008325")
		//rsp, err := tknClient.Get(bbURL + "/v1/fhir/Nothing")

		if err != nil {
			b.Fatal(err)
		}

		if rsp.StatusCode != http.StatusOK {
			b.Fatal(rsp)
		}

		httputil.DumpResponse(rsp, true)

		eob, _ := ioutil.ReadAll(rsp.Body)
		b.Log(len(eob))
		//var out bytes.Buffer
		//json.Indent(&out, eob, "	", "\t")
	}
}
