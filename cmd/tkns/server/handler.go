package server

import (
	"encoding/json"
	"fmt"
	"io"
	"log"

	"golang.org/x/oauth2"
)

func WriteTkn(w io.Writer) TokenHandlerFunc {
	return func(tkn *oauth2.Token) error {
		t := &Token{*tkn}
		_, err := fmt.Fprintln(w, t)
		return err
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
