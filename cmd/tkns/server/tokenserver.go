package server

import (
	"log"
	"net/http"
	"strings"

	"golang.org/x/oauth2"
)

type TokenHandlerFunc func(*oauth2.Token) error

func TokenHandler(conf *oauth2.Config, th TokenHandlerFunc) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		//breq, _ := httputil.DumpRequest(r, true)
		//log.Println(string(breq))
		if code := r.FormValue("code"); code != "" {
			tok, err := conf.Exchange(r.Context(), strings.TrimSpace(code))
			if err != nil {
				log.Println("[ERROR] bad token exchange", err, tok)
				http.Error(
					rw,
					err.Error(),
					http.StatusInternalServerError,
				)
				return
			}
			if err := th(tok); err != nil {
				log.Println("[ERROR] bad token handling", err, tok)
				http.Error(
					rw,
					err.Error(),
					http.StatusInternalServerError,
				)

				return
			}

			rw.Write([]byte("well done!"))
			rw.WriteHeader(http.StatusOK)

			return
		}
		if err := r.FormValue("error"); err != "" {
			http.Error(
				rw,
				err,
				http.StatusBadRequest,
			)
			return
		}
		authURL := conf.AuthCodeURL("bobert")
		//oauth2.SetAuthURLParam("response_type", "token"))
		log.Println(authURL)
		http.Redirect(rw, r, authURL, http.StatusSeeOther)
	})
}
