package bluebutton

import (
	"os"
)

var (
	CLIENT_ID      string
	CLIENT_SECRET  string
	BB_URL         string
	AUTHORIZE_PATH = "/v1/o/authorize/"
	TOKEN_PATH     = "/v1/o/token/"
)

func init() {
	CLIENT_ID = os.Getenv("BLUE_BUTTON_CLIENT_ID")
	CLIENT_SECRET = os.Getenv("BLUE_BUTTON_CLIENT_SECRET")
	BB_URL = os.Getenv("BLUE_BUTTON_URL")
}

func AuthURL() string {
	return BB_URL + AUTHORIZE_PATH
}

func TokenURL() string {
	return BB_URL + TOKEN_PATH
}
