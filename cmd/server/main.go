package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"

	"github.com/coreos/go-oidc/v3/oidc"
)

const (
	// oidc discovery endpoint
	IssuerUrlEnvVar            string = "ISSUER_URL"
	OidcIntendedAudienceEnvVar string = "OIDC_INTENDED_AUDIENCE"
)

func echoHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	}
}

func OIDCVerifyTokenHandler(idTokenVerifier *oidc.IDTokenVerifier, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		var err error
		if token == "" {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Authorization header missing"))
			return
		}
		_, err = idTokenVerifier.Verify(context.Background(), token)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(err.Error()))
			return
		}
		h.ServeHTTP(w, r)
	})
}

func main() {
	fmt.Println("starting server")
	mux := http.NewServeMux()

	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	oidcCtx := oidc.ClientContext(context.Background(), http.DefaultClient)
	issuer := os.Getenv(IssuerUrlEnvVar)
	provider, err := oidc.NewProvider(oidcCtx, issuer)
	if err != nil {
		fmt.Printf("%v", err)
		os.Exit(1)
	}

	verifier := provider.Verifier(&oidc.Config{ClientID: os.Getenv(OidcIntendedAudienceEnvVar)})

	mux.Handle("/echo", OIDCVerifyTokenHandler(verifier, echoHandler()))
	s := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
	err = s.ListenAndServe()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
