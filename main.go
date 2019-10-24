package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/google/uuid"

	"strconv"

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

type OauthRequest struct {
	JSONConfig  *Config
	OauthConfig *oauth2.Config
}

//IsProd method checks if the environment is Prod. If so, it might eventually do something.
func (c Config) IsProd() bool {
	return c.Env == "prod"
}

//OautCallback method receives the Activation Code from the IDP, then exchanges it for the Token.
func (o *OauthRequest) OauthCallback(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	fmt.Fprintln(w, "", "code:", r.FormValue("code"), "state:", r.FormValue("state"))

	code := r.FormValue("code")
	token, err := o.OauthConfig.Exchange(context.TODO(), code)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	fmt.Fprintln(w, "token:", token.AccessToken)
	fmt.Fprintln(w, "type:", token.TokenType)
	fmt.Fprintln(w, "refresh token:", token.RefreshToken)
	fmt.Fprintln(w, "expiryf:", token.Expiry)
}

//OauthConnect method sets a random state, then connects to the IDP's AuthCodeURL provided in the config file.
func (o *OauthRequest) OauthConnect(w http.ResponseWriter, r *http.Request) {
	state, err := uuid.NewRandom()
	if err != nil {
		fmt.Println("Error generating random UUID for state: ", err)
	}
	fmt.Println(state)
	url := o.OauthConfig.AuthCodeURL(state.String())
	http.Redirect(w, r, url, http.StatusFound)
}

func main() {

	//Load configuration
	cfg := loadConfig()

	//Create an OauthRequest object and populate it with configuration data.
	oauth2req := OauthRequest{
		OauthConfig: &oauth2.Config{
			ClientID:     cfg.Dropbox.ID,
			ClientSecret: cfg.Dropbox.Secret,
			Endpoint: oauth2.Endpoint{
				AuthURL:  cfg.Dropbox.AuthURL,
				TokenURL: cfg.Dropbox.TokenURL,
			},
			RedirectURL: cfg.Dropbox.RedirectURL,
		},
	}

	//Setup HTTP server and routing.
	http.HandleFunc("/oauth/dropbox/connect", oauth2req.OauthConnect)
	http.HandleFunc("/oauth/dropbox/callback", oauth2req.OauthCallback)
	http.ListenAndServe(":"+strconv.Itoa(cfg.Port), nil)

}

//loadConfig loads configuration from file.
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

//DefaultConfig sets the configuration to use if config file isn't loaded.
func DefaultConfig() Config {
	return Config{
		Port: 3000,
		Env:  "dev",
	}
}
