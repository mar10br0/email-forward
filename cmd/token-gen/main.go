package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"golang.org/x/oauth2"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"

	"github.com/google/uuid"

	gmailconfig "github.com/mar10br0/email-forward/internal/gmail/config"
)

var LAMBDA_PACKAGE_FILENAME = "email-forwarder-lambda.zip"
var LAMBDA_BUCKET = os.Getenv("LAMBDA_BUCKET")

func main() {
	var err error

	var config *oauth2.Config
	if config, _, err = gmailconfig.Load(); err != nil {
		log.Failure(err, "")
		return
	}

	state := uuid.New().String()
	authURL := config.AuthCodeURL(state, oauth2.AccessTypeOffline)
	log.Progress("Go to the following link in your browser for obtaining an authorization code:\n%s", authURL)

	server := &http.Server{}
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Progress("Processing received token code")
		if !r.URL.Query().Has("state") || !r.URL.Query().Has("code") || !r.URL.Query().Has("scope") {
			log.Failure(errors.New("missing state, code or scope"), "")
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if r.URL.Query().Get("state") != state {
			log.Failure(errors.New("invalid state"), "")
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if r.URL.Query().Get("scope") != strings.Join(gmailconfig.SCOPES, " ") {
			log.Failure(errors.New("invalid scope"), "\nExpected: %s\nReceived: %s", strings.Join(gmailconfig.SCOPES, " "), r.URL.Query().Get("scope"))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		var token *oauth2.Token
		var err error
		if token, err = config.Exchange(r.Context(), r.URL.Query().Get("code")); err != nil {
			log.Failure(err, "Exchanging authorization token for an access token")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		var f *os.File
		if f, err = os.OpenFile(gmailconfig.TOKEN_FILE, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600); err != nil {
			log.Failure(err, "Creating access token file %s", gmailconfig.TOKEN_FILE)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer f.Close()
		if err = json.NewEncoder(f).Encode(token); err != nil {
			log.Failure(err, "Writing access token to %s", gmailconfig.TOKEN_FILE)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		log.Progress("Re-packaging lambda function to %s", LAMBDA_PACKAGE_FILENAME)
		var pckgOut []byte
		if pckgOut, err = exec.Command("build-lambda-zip.exe", "-o", LAMBDA_PACKAGE_FILENAME, "lambda", gmailconfig.CREDENTIALS_FILE, gmailconfig.TOKEN_FILE).Output(); err != nil {
			log.Failure(err, string(pckgOut))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		log.Debug(string(pckgOut))

		log.Progress("Uploading %s to S3", LAMBDA_PACKAGE_FILENAME)
		s3Svc := s3.New(session.Must(session.NewSession()), aws.NewConfig().WithLogLevel(log.LEVEL))
		var pckgData []byte
		if pckgData, err = os.ReadFile(LAMBDA_PACKAGE_FILENAME); err != nil {
			log.Failure(err, "reading lambda-package")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		var pckgBody io.ReadSeeker = io.ReadSeeker(strings.NewReader(string(pckgData)))
		if _, err = s3Svc.PutObject(&s3.PutObjectInput{Bucket: &LAMBDA_BUCKET, Key: &LAMBDA_PACKAGE_FILENAME, Body: pckgBody}); err != nil {
			log.Failure(err, "uploading lambda-package")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		log.Success("Access token generated and updated lambda-package uploaded to %s", LAMBDA_BUCKET)
		w.Write([]byte(fmt.Sprintf("Access token generated and updated lambda-package uploaded to %s", LAMBDA_BUCKET)))

		server.Shutdown(r.Context())
	})
	server.ListenAndServe()
}
