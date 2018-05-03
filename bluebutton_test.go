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
	if CLIENT_ID == "" || CLIENT_SECRET == "" {
		t.Skip()
	}
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
	result := `oauth2: cannot fetch token: 401 Unauthorized
Response: {"error": "invalid_grant"}`

	if strings.Compare(err.Error(), result) != 0 {
		t.Error(err)
	}
}

func TestPasswordCredentialsToken(t *testing.T) {
	if CLIENT_ID == "" || CLIENT_SECRET == "" {
		t.Skip()
	}
	conf := &oauth2.Config{
		RedirectURL:  RedirectURL,
		ClientID:     CLIENT_ID,
		ClientSecret: CLIENT_SECRET,
		Endpoint: oauth2.Endpoint{
			AuthURL:  bbURL + "/v1/o/authorize/",
			TokenURL: bbURL + "/v1/o/token/",
		},
	}

	_, err := conf.PasswordCredentialsToken(context.Background(), "blah", "blahblah")
	result := `oauth2: cannot fetch token: 401 Unauthorized
Response: {"error_description": "Invalid credentials given.", "error": "invalid_grant"}`
	altResult := `oauth2: cannot fetch token: 401 Unauthorized
Response: {"error": "invalid_grant", "error_description": "Invalid credentials given."}`

	if strings.Compare(err.Error(), result) != 0 && strings.Compare(err.Error(), altResult) != 0 {
		t.Error(err)
	}
}

func TestRefreshToken(t *testing.T) {
	if CLIENT_ID == "" || CLIENT_SECRET == "" {
		t.Skip()
	}
	conf := &oauth2.Config{
		RedirectURL:  RedirectURL,
		ClientID:     CLIENT_ID,
		ClientSecret: CLIENT_SECRET,
		Endpoint: oauth2.Endpoint{
			AuthURL:  bbURL + "/v1/o/authorize/",
			TokenURL: bbURL + "/v1/o/token/",
		},
	}
	tkn := `{
	"access_token": "old-bogus",
	"token_type": "Bearer",
	"refresh_token": "bogus",
	"expiry": "2018-01-01T00:00:00-05:00"
}`
	tok := &oauth2.Token{}
	if err := json.Unmarshal([]byte(tkn), tok); err != nil {
		t.Fatal(err)
	}

	tknClient := conf.Client(context.Background(), tok)

	_, err := tknClient.Get(bbURL + "/v1/fhir/ExplanationOfBenefit/?patient=20140000008325")
	//r, _ := httputil.DumpResponse(rsp, true)
	//log.Println(string(r))

	result := `Get https://sandbox.bluebutton.cms.gov/v1/fhir/ExplanationOfBenefit/?patient=20140000008325: oauth2: cannot fetch token: 401 Unauthorized
Response: {"error": "invalid_grant"}`

	if strings.Compare(err.Error(), result) != 0 {
		t.Error(err)
	}
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
		//fmt.Println(out.String())
	}
}
