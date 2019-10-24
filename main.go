package main

import (
	"fmt"
	"net/http"
	"os"

	"encoding/json"

	"context"

	"golang.org/x/oauth2"
)

type Config struct {
	Port    int           `json:"port"`
	Env     string        `json:"env"`
	Dropbox DropboxConfig `json:"dropbox"`
}

type DropboxConfig struct {
	ID          string `json:"id"`
	Secret      string `json:"secret"`
	AuthURL     string `json:"auth_url"`
	TokenURL    string `json:"token_url"`
	RedirectURL string `json:"redirect_url"`
}

func (c Config) IsProd() bool {
	return c.Env == "prod"
}

func DefaultConfig() Config {
	return Config{
		Port: 3000,
		Env:  "dev",
	}
}

func main() {

	cfg := loadConfig()
	dbxOauth := &oauth2.Config{
		ClientID:     cfg.Dropbox.ID,
		ClientSecret: cfg.Dropbox.Secret,
		Endpoint: oauth2.Endpoint{
			AuthURL:  cfg.Dropbox.AuthURL,
			TokenURL: cfg.Dropbox.TokenURL,
		},
		RedirectURL: cfg.Dropbox.RedirectURL,
	}

	dbxRedirect := func(w http.ResponseWriter, r *http.Request) {
		state := "TODO-CREATE_STATE"
		fmt.Println(state)
		url := dbxOauth.AuthCodeURL(state)
		http.Redirect(w, r, url, http.StatusFound)
	}

	http.HandleFunc("/oauth/dropbox/connect", dbxRedirect)

	http.HandleFunc("/authorize", authorize)

	dbxCallback := func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		fmt.Fprintln(w, "", "code:", r.FormValue("code"), "state:", r.FormValue("state"))

		code := r.FormValue("code")
		token, err := dbxOauth.Exchange(context.TODO(), code)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		fmt.Fprintln(w, "token:", token.AccessToken)
		fmt.Fprintln(w, "type:", token.TokenType)
		fmt.Fprintln(w, "refresh token:", token.RefreshToken)
		fmt.Fprintln(w, "expiryf:", token.Expiry)
	}
	http.HandleFunc("/oauth/dropbox/callback", dbxCallback)
	http.ListenAndServe(":3000", nil)

}

func authorize(w http.ResponseWriter, r *http.Request) {

	// cfg := loadConfig()
	// dbxOauth := &oauth2.Config{
	// 	ClientID:     cfg.Dropbox.ID,
	// 	ClientSecret: cfg.Dropbox.Secret,
	// 	Endpoint: oauth2.Endpoint{
	// 		AuthURL:  cfg.Dropbox.AuthURL,
	// 		TokenURL: cfg.Dropbox.TokenURL,
	// 	},
	// 	RedirectURL: cfg.Dropbox.RedirectURL,
	// }

	// fmt.Fprintf(w, "", cfg)

}

func loadConfig() Config {
	file, err := os.Open("oauth_svc.conf")
	if err != nil {
		fmt.Println(err, ": Using default config")
		return DefaultConfig()
	}
	var c Config
	dec := json.NewDecoder(file)
	err = dec.Decode(&c)
	if err != nil {
		panic(err)
	}
	if err != nil {
		panic(err)
	}
	return c

}
