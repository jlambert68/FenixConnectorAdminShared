package gcp

import (
	"bytes"
	"encoding/json"
	"fmt"
	uuidGenerator "github.com/google/uuid"
	"github.com/gorilla/pat"
	"github.com/gorilla/sessions"
	"github.com/jlambert68/FenixConnectorAdminShared/common_config"
	fenixSyncShared "github.com/jlambert68/FenixSyncShared"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"google.golang.org/api/idtoken"
	grpcMetadata "google.golang.org/grpc/metadata"
	"html/template"
	"io/ioutil"
	"net/http"
	"os/exec"
	"runtime"
	"time"
)

// GenerateTokenTargetType
// Type used to define
type GenerateTokenTargetType int

// GenerateTokenForExecutionServer
// Constants used to define what Token should be used for
const (
	GenerateTokenForGrpcTowardsExecutionWorker GenerateTokenTargetType = iota
	GenerateTokenForPubSub
	GetTokenForGrpcAndPubSub
)

/*type ServiceAccountInfo struct {
	Aliases []string `json:"aliases"`
	Email   string   `json:"email"`
	Scopes  []string `json:"scopes"`
}

*/

// Key for NewCookieStore
var tempKeyAsHash string

func (gcp *GcpObjectStruct) GenerateGCPAccessToken(ctx context.Context, tokenTarget GenerateTokenTargetType) (
	appendedCtx context.Context, returnAckNack bool, returnMessage string) {

	// Check if GCP auth-token should be received from SPIRE-server in OpenShift
	if common_config.ShouldSpireServerBeUsedForGettingGcpToken == true {

		return ctx, true, ""

		//appendedCtx, returnAckNack, returnMessage = gcp.generateGCPAccessTokenFromOpenShift(ctx)

		//return appendedCtx, returnAckNack, returnMessage
	}

	// Chose correct method for authentication
	switch tokenTarget { // common_config.UseServiceAccount == true {

	case GenerateTokenForGrpcTowardsExecutionWorker:
		// Only use Authorized used when running locally and WorkerServer is on GCP
		if common_config.ExecutionLocationForConnector == common_config.LocalhostNoDocker &&
			common_config.ExecutionLocationForFenixExecutionWorkerServer == common_config.GCP {

			// Use Authorized user when targeting GCP from local
			appendedCtx, returnAckNack, returnMessage = gcp.GenerateGCPAccessTokenForAuthorizedUser(ctx)

		} else {
			// Use Authorized user
			appendedCtx, returnAckNack, returnMessage = gcp.generateGCPAccessToken(ctx)
		}

	case GenerateTokenForPubSub:
		// Only use Authorized used when running locally and WorkerServer is on GCP
		if common_config.ExecutionLocationForConnector == common_config.LocalhostNoDocker {

			// Use Authorized user when targeting GCP from local
			appendedCtx, returnAckNack, returnMessage = gcp.GenerateGCPAccessTokenForAuthorizedUserPubSub(ctx)

		} else {
			// Use Authorized user
			appendedCtx, returnAckNack, returnMessage = gcp.generateGCPAccessTokenPubSub(ctx)
		}

	case GetTokenForGrpcAndPubSub:
		// Only use Authorized used when running locally and WorkerServer is on GCP
		if common_config.ExecutionLocationForConnector == common_config.LocalhostNoDocker {

			// Use Authorized user when targeting GCP from local
			appendedCtx, returnAckNack, returnMessage = gcp.GenerateGCPAccessTokenForAuthorizedUserPubSub(ctx) // gcp.GenerateGCPAccessTokenForOAuthUserPubSub(ctx)

		} else {
			// Use Authorized user
			appendedCtx, returnAckNack, returnMessage = gcp.generateGCPAccessTokenPubSub(ctx)
		}

	}
	return appendedCtx, returnAckNack, returnMessage

}

// Generate Google access token. Used when running in OpenShift
func (gcp *GcpObjectStruct) generateGCPAccessTokenFromOpenShift(ctx context.Context) (appendedCtx context.Context, returnAckNack bool, returnMessage string) {

	// sgcp endpoint for fetching the token
	sgcpEndpoint := common_config.OpenShiftsGcpTokenSourceUrl

	// Create a new HTTP request to fetch the token
	req, err := http.NewRequest("GET", sgcpEndpoint, nil)
	if err != nil {

		common_config.Logger.WithFields(logrus.Fields{
			"ID":  "d0c71fbe-5d02-4d58-b5d8-e5e9d104d45b",
			"err": err,
		}).Fatalln("Failed to create request: %v\n", err)
	}

	// Perform the HTTP request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {

		common_config.Logger.WithFields(logrus.Fields{
			"ID":  "30a6291e-ecf2-4011-87d8-0697389d3a60",
			"err": err,
		}).Fatalln("Failed to perform request: %v\n", err)
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Failed to read response body: %v\n", err)
		common_config.Logger.WithFields(logrus.Fields{
			"ID":  "8f6cdef3-619f-4594-9b37-eabf7f5fb903",
			"err": err,
		}).Fatalln("Failed to read response body: %v\n", err)
	}

	// Print the token
	token := string(body)
	//fmt.Printf("GCP Token: %s\n", token)

	common_config.Logger.WithFields(logrus.Fields{
		"ID": "6ac9a073-195f-42b0-8d9f-3d0a8ebf7ed0",
		//"FenixExecutionWorkerObject.gcpAccessToken": gcp.gcpAccessTokenForServiceAccounts,
	}).Debug("Will use Bearer Token")

	// Add token to GrpcServer Request.
	appendedCtx = grpcMetadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+token)

	return appendedCtx, true, ""

}

// Generate Google access token. Used when running in GCP
func (gcp *GcpObjectStruct) generateGCPAccessToken(ctx context.Context) (appendedCtx context.Context, returnAckNack bool, returnMessage string) {

	// Only create the token if there is none, or it has expired (or 5 minutes before expiration
	var safetyDuration time.Duration
	safetyDuration = -5 * time.Minute
	if gcp.gcpAccessTokenForServiceAccounts == nil || gcp.gcpAccessTokenForServiceAccounts.Expiry.Add(safetyDuration).Before(time.Now()) {

		// Create an identity token.
		// With a global TokenSource tokens would be reused and auto-refreshed at need.
		// A given TokenSource is specific to the audience.
		var tokenSource oauth2.TokenSource
		var err error

		tokenSource, err = idtoken.NewTokenSource(ctx, "https://"+common_config.FenixExecutionWorkerAddress)

		if err != nil {
			common_config.Logger.WithFields(logrus.Fields{
				"ID":  "11b41921-92fa-48ed-914f-0dde41282609",
				"err": err,
			}).Error("Couldn't generate access token")

			return nil, false, "Couldn't generate access token"
		}

		token, err := tokenSource.Token()
		if err != nil {
			common_config.Logger.WithFields(logrus.Fields{
				"ID":  "c1870620-d615-45e8-aaae-a1329d2ff4af",
				"err": err,
			}).Error("Problem getting the token")

			return nil, false, "Problem getting the token"
		} else {
			common_config.Logger.WithFields(logrus.Fields{
				"ID": "fee61402-aefa-4d4a-87ff-04b02c055366",
				//"token": token,
			}).Debug("Got Bearer Token")
		}

		gcp.gcpAccessTokenForServiceAccounts = token

	}

	common_config.Logger.WithFields(logrus.Fields{
		"ID": "52ccc212-601d-409b-a177-28782fa09d6f",
		//"gcp.gcpAccessTokenForServiceAccounts": gcp.gcpAccessTokenForServiceAccounts,
	}).Debug("Will use Bearer Token")

	// Add token to GrpcServer Request.
	appendedCtx = grpcMetadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+gcp.gcpAccessTokenForServiceAccounts.AccessToken)

	return appendedCtx, true, ""

}

// Generate Google access token for Pub Sub
func (gcp *GcpObjectStruct) generateGCPAccessTokenPubSub(ctx context.Context) (appendedCtx context.Context, returnAckNack bool, returnMessage string) {

	// Only create the token if there is none, or it has expired (or 5 minutes before expiration
	var safetyDuration time.Duration
	safetyDuration = -5 * time.Minute

	if gcp.gcpAccessTokenForServiceAccountsPubSub == nil || gcp.gcpAccessTokenForServiceAccountsPubSub.Expiry.Add(safetyDuration).Before(time.Now()) {

		// Create an identity token.
		// With a global TokenSource tokens would be reused and auto-refreshed at need.
		// A given TokenSource is specific to the audience.

		tokenSource, err := idtoken.NewTokenSource(ctx, "https://www.googleapis.com/auth/pubsub")
		if err != nil {
			common_config.Logger.WithFields(logrus.Fields{
				"ID":  "ffb7cdcc-00f1-4560-9fd6-a45d2423230d",
				"err": err,
			}).Error("Couldn't generate access token")

			return nil, false, "Couldn't generate access token"
		}

		token, err := tokenSource.Token()
		if err != nil {
			common_config.Logger.WithFields(logrus.Fields{
				"ID":  "6f335c25-b020-4748-85ab-eda80e53b9a0",
				"err": err,
			}).Error("Problem getting the token")

			return nil, false, "Problem getting the token"
		} else {
			common_config.Logger.WithFields(logrus.Fields{
				"ID":    "a17e40dc-e7fc-4d7e-afbc-072a4c21850b",
				"token": token,
			}).Debug("Got Bearer Token")
		}

		gcp.gcpAccessTokenForServiceAccountsPubSub = token

	}

	common_config.Logger.WithFields(logrus.Fields{
		"ID": "be4b0370-e578-4664-8fd7-4bb5099e6261",
		"gcp.gcpAccessTokenForServiceAccountsPubSub.AccessToken": gcp.gcpAccessTokenForServiceAccountsPubSub.AccessToken,
		"gcp.gcpAccessTokenForServiceAccountsPubSub.Expiry":      gcp.gcpAccessTokenForServiceAccountsPubSub.Expiry,
		"time.Now()": time.Now(),
	}).Info("Will use Bearer Token")

	common_config.Logger.WithFields(logrus.Fields{
		"ID": "7913b32c-70c5-4ae5-841f-04943107131c",
		" gcp.gcpAccessTokenForServiceAccountsPubSub.AccessToken": gcp.gcpAccessTokenForServiceAccountsPubSub.AccessToken,
	}).Debug("Will use Bearer Token")

	// Add token to GrpcServer Request.
	appendedCtx = grpcMetadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+gcp.gcpAccessTokenForServiceAccountsPubSub.AccessToken)

	return appendedCtx, true, ""

}

// DoneChannel - channel used for to close down local web server
var DoneChannel chan bool

func (gcp *GcpObjectStruct) GenerateGCPAccessTokenForAuthorizedUser(ctx context.Context) (appendedCtx context.Context, returnAckNack bool, returnMessage string) {

	// Secure that User is initiated
	gcp.initiateUserObject()

	// Only create the token if there is none, or it has expired (or 5 minutes before expiration
	var safetyDuration time.Duration
	safetyDuration = -5 * time.Minute
	timeToCompareTo := gcp.gcpAccessTokenForAuthorizedAccountsPubSub.ExpiresAt.Add(safetyDuration)
	if gcp.gcpAccessTokenForAuthorizedAccountsPubSub.IDToken != "" && timeToCompareTo.Before(time.Now()) {
		// We already have a ID-token that can be used, so return that
		appendedCtx = grpcMetadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+gcp.gcpAccessTokenForAuthorizedAccounts.IDToken)

		return appendedCtx, true, ""
	} else if gcp.gcpAccessTokenForAuthorizedAccountsPubSub.IDToken != "" && timeToCompareTo.After(time.Now()) {
		// Update with new token

	}

	// Need to create a new ID-token

	key := common_config.ApplicationRunTimeUuid // Replace with your SESSION_SECRET or similar
	maxAge := 86400 * 30                        // 30 days
	isProd := false                             // Set to true when serving over https

	store := sessions.NewCookieStore([]byte(key))
	store.MaxAge(maxAge)
	store.Options.Path = "/"
	store.Options.HttpOnly = true // HttpOnly should always be enabled
	store.Options.Secure = isProd

	gothic.Store = store

	goth.UseProviders(
		// Use 'Fenix End User Authentication'
		google.New(
			common_config.AuthClientId,
			common_config.AuthClientSecret,
			"http://localhost:3000/auth/google/callback",
			"email", "profile"),
	)

	router := pat.New()

	router.Get("/auth/{provider}/callback", func(res http.ResponseWriter, req *http.Request) {

		user, err := gothic.CompleteUserAuth(res, req)
		if err != nil {

			fmt.Fprintln(res, err)

			return
		}
		t, _ := template.ParseFiles("templates/success.html")
		t.Execute(res, user)

		// Save ID-token
		gcp.gcpAccessTokenForAuthorizedAccounts = user

		// Trigger Close of Web Server, and 'true' means that a ID-to
		DoneChannel <- true

	})

	router.Get("/logout/{provider}", func(res http.ResponseWriter, req *http.Request) {
		gothic.Logout(res, req)
		res.Header().Set("Location", "/")
		res.WriteHeader(http.StatusTemporaryRedirect)
	})

	router.Get("/auth/{provider}", func(res http.ResponseWriter, req *http.Request) {
		gothic.BeginAuthHandler(res, req)
	})

	router.Get("/", func(res http.ResponseWriter, req *http.Request) {
		t, _ := template.ParseFiles("templates/index.html")
		t.Execute(res, false)
	})

	// Initiate channel used to stop server
	DoneChannel = make(chan bool, 1)

	// Initiate http server
	localWebServer := &http.Server{
		Addr:    ":3000",
		Handler: router,
	}

	// Start Local Web Server as go routine
	url := "http://localhost:3000"
	go gcp.startLocalWebServer(localWebServer, url)

	common_config.Logger.WithFields(logrus.Fields{
		"ID": "689d42de-3cc0-4237-b1e9-3a6c769f65ea",
	}).Debug("Local webServer Started")

	// Wait for message in channel to stop local web server
	gotIdTokenResult := <-DoneChannel

	// Shutdown local web server
	gcp.stopLocalWebServer(context.Background(), localWebServer)

	// Depending on the outcome of getting a token return different results
	if gotIdTokenResult == true {
		// Success in getting an ID-token
		appendedCtx = grpcMetadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+gcp.gcpAccessTokenForAuthorizedAccounts.IDToken)

		return appendedCtx, true, ""
	} else {
		// Didn't get any ID-token
		return nil, false, "Couldn't generate access token"
	}

}

func (gcp *GcpObjectStruct) GenerateGCPAccessTokenForAuthorizedUserPubSub(ctx context.Context) (appendedCtx context.Context, returnAckNack bool, returnMessage string) {

	// Set Key for NewCookieStore
	if len(tempKeyAsHash) == 0 {
		tempKeyAsHash = fenixSyncShared.HashSingleValue(uuidGenerator.New().String())
	}

	// Secure that User is initiated
	gcp.initiateUserObjectPubSub()

	router := pat.New()
	var url string

	// Initiate http server
	localWebServer := &http.Server{
		Addr:    ":3000",
		Handler: router,
	}

	// Only create the token if there is none, or it has expired (or 5 minutes before expiration
	var safetyDuration time.Duration
	var timeToCompareTo time.Time
	safetyDuration = -5 * time.Minute
	if gcp.gcpAccessTokenForAuthorizedAccountsPubSub.IDToken != "" {

		timeToCompareTo = gcp.refreshTokenResponse.ExpiresAt.Add(safetyDuration)
	}
	if gcp.gcpAccessTokenForAuthorizedAccountsPubSub.IDToken != "" && timeToCompareTo.After(time.Now()) {
		// We already have a ID-token that can be used, so return that
		appendedCtx = grpcMetadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+gcp.refreshTokenResponse.IDToken)

		return appendedCtx, true, ""

	} else if gcp.gcpAccessTokenForAuthorizedAccountsPubSub.IDToken != "" && timeToCompareTo.Before(time.Now()) {
		client := &http.Client{
			// Configure the client if necessary. For example, set a timeout:
			Timeout: time.Second * 30,
		}

		refreshTokenResponseMessage, err := refreshToken(client, gcp.gcpAccessTokenForAuthorizedAccountsPubSub.RefreshToken)
		if err != nil {
			fmt.Println("err: ", err)

			return nil, false, err.Error()

		} else {

			// When no refresh token was received then ask user to close the web browser containing previous log in credentials
			if gcp.gcpAccessTokenForAuthorizedAccountsPubSub.RefreshToken == "" {
				url = "http://localhost:3000/close_this_browser"
				go gcp.startLocalWebServer(localWebServer, url)

				return nil, false, "Missing Refresh token"
			}

			// Store Refresh response
			gcp.refreshTokenResponse = refreshTokenResponseMessage

			//
			appendedCtx = grpcMetadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+gcp.refreshTokenResponse.IDToken)
			return appendedCtx, true, ""
		}

	}

	// Need to create a new ID-token

	key := tempKeyAsHash // Replace with your SESSION_SECRET or similar
	maxAge := 86400 * 30 // 30 days
	isProd := false      // Set to true when serving over https

	store := sessions.NewCookieStore([]byte(key))
	store.MaxAge(maxAge)
	store.Options.Path = "/"
	store.Options.HttpOnly = true // HttpOnly should always be enabled
	store.Options.Secure = isProd

	gothic.Store = store

	goth.UseProviders(
		// Use 'Fenix End User Authentication'

		google.New(
			common_config.AuthClientId,
			common_config.AuthClientSecret,
			"http://localhost:3000/auth/google/callback",
			"email", "profile", "https://www.googleapis.com/auth/pubsub"),
	)

	router.Get("/auth/{provider}/callback", func(res http.ResponseWriter, req *http.Request) {

		user, err := gothic.CompleteUserAuth(res, req)
		if err != nil {

			fmt.Fprintln(res, err)

			return
		}

		// Save ID-token
		gcp.gcpAccessTokenForAuthorizedAccountsPubSub = user

		// When we got an Refresh Token then inform of Success
		// When there was no Refresh Token then inform user to close Browser and restart
		if len(user.RefreshToken) > 0 {
			// Got Refresh Token
			t, _ := template.ParseFiles("templates/success.html")
			t.Execute(res, user)

			// Trigger Close of Web Server, and 'true' means that a ID-to
			DoneChannel <- true

		} else {
			// Didn't get Refresh Token
			t, _ := template.ParseFiles("templates/close_this_browser.html")
			t.Execute(res, false)

			// Trigger Close of Web Server, and 'false' means no Refresh Token
			DoneChannel <- false

		}

	})

	router.Get("/logout/{provider}", func(res http.ResponseWriter, req *http.Request) {
		gothic.Logout(res, req)
		res.Header().Set("Location", "/")
		res.WriteHeader(http.StatusTemporaryRedirect)

	})

	//
	router.Get("/auth/{provider}", func(res http.ResponseWriter, req *http.Request) {
		gothic.BeginAuthHandler(res, req)
	})

	// Start page for web server for user to be able to login into GCP
	router.Get("/", func(res http.ResponseWriter, req *http.Request) {
		//res.Header().Set("state-token", "offline")
		t, _ := template.ParseFiles("templates/index.html")
		t.Execute(res, false)
	})

	// Show Text telling user to close down web browser due to that no Refresh Token can be retrieved
	// as long as browser window is open
	/*
		router.Get("/closethisbrowser", func(res http.ResponseWriter, req *http.Request) {
			//res.Header().Set("state-token", "offline")
			t, _ := template.ParseFiles("templates/close_this_browser.html")
			t.Execute(res, false)

			// Shutdown local web server
			gcp.stopLocalWebServer(context.Background(), localWebServer)

		})

	*/

	// Initiate channel used to stop server
	DoneChannel = make(chan bool, 1)

	// Start Local Web Server as go routine
	url = "http://localhost:3000"
	go gcp.startLocalWebServer(localWebServer, url)

	common_config.Logger.WithFields(logrus.Fields{
		"ID": "689d42de-3cc0-4237-b1e9-3a6c769f65ea",
	}).Debug("Local webServer Started")

	// Wait for message in channel to stop local web server
	gotIdTokenResult := <-DoneChannel

	// Shutdown local web server
	gcp.stopLocalWebServer(context.Background(), localWebServer)

	// Depending on the outcome of getting a token return different results
	if gotIdTokenResult == true {
		// Success in getting an ID-token first time so use RefreshToken to fill RefreshTokenMessage
		client := &http.Client{
			// Configure the client if necessary. For example, set a timeout:
			Timeout: time.Second * 30,
		}
		refreshTokenResponseMessage, err := refreshToken(client, gcp.gcpAccessTokenForAuthorizedAccountsPubSub.RefreshToken)
		if err != nil {
			fmt.Println("err: ", err)

			return nil, false, err.Error()

		} else {

			// When no refresh token was received then ask user to close the web browser containing previous log in credentials
			if gcp.gcpAccessTokenForAuthorizedAccountsPubSub.RefreshToken == "" {
				url = "http://localhost:3000/closethisbrowser"
				gcp.startLocalWebServer(localWebServer, url)

				fmt.Println("Hej")
				time.Sleep(10 * time.Second)

				return nil, false, "Missing Refresh token"
			}

			// Store Refresh response
			gcp.refreshTokenResponse = refreshTokenResponseMessage

			//
			appendedCtx = grpcMetadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+gcp.refreshTokenResponse.IDToken)
			return appendedCtx, true, ""
		}

		appendedCtx = grpcMetadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+gcp.gcpAccessTokenForAuthorizedAccountsPubSub.IDToken)

		return appendedCtx, true, ""
	} else {
		// Didn't get any ID-token
		return nil, false, "Couldn't generate access token"
	}

}

func (gcp *GcpObjectStruct) GetGcpAccessTokenForAuthorizedAccountsPubSub() string {

	// Only use Authorized used when running locally and WorkerServer is on GCP
	if common_config.ExecutionLocationForConnector == common_config.LocalhostNoDocker {

		// Use Authorized user when targeting GCP from local
		return gcp.refreshTokenResponse.AccessToken

	} else {

		_, returnAckNack, returnMessage := gcp.GenerateGCPAccessToken(context.Background(), GenerateTokenForPubSub)
		if returnAckNack == false {

			common_config.Logger.WithFields(logrus.Fields{
				"id":            "e929a3ac-44a3-439a-a820-493e13318489",
				"returnMessage": returnMessage,
			}).Error("Problem when generating a new token. Waiting some time before next try")
		}
		// Use token for Service Account
		return gcp.gcpAccessTokenForServiceAccountsPubSub.AccessToken
	}

}

func (gcp *GcpObjectStruct) GenerateGCPAccessTokenForOAuthUserPubSub(
	ctx context.Context) (appendedCtx context.Context, returnAckNack bool, returnMessage string) {

	// Initiate channel used to stop server
	DoneChannel = make(chan bool, 1)

	http.HandleFunc("/login", handleLogin)
	http.HandleFunc("/auth/google/callback", handleCallback)
	http.ListenAndServe(":3000", nil)
	/*

		// Initiate http server
		localWebServer := &http.Server{
			Addr:    ":3000",
			Handler: router,
		}

		// Start Local Web Server as go routine
		go gcp.startLocalWebServer(localWebServer)

		common_config.Logger.WithFields(logrus.Fields{
			"ID": "689d42de-3cc0-4237-b1e9-3a6c769f65ea",
		}).Debug("Local webServer Started")

		// Wait for message in channel to stop local web server
		gotIdTokenResult := <-DoneChannel

		// Shutdown local web server
		gcp.stopLocalWebServer(context.Background(), localWebServer)


	*/

	appendedCtx = grpcMetadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+gcp.gcpAccessTokenForAuthorizedAccountsPubSub.IDToken)

	return appendedCtx, returnAckNack, returnMessage
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	var conf = &oauth2.Config{
		ClientID:     common_config.AuthClientId,
		ClientSecret: common_config.AuthClientSecret,
		RedirectURL:  "http://localhost:3000/auth/google/callback",
		Scopes:       []string{"email", "profile", "https://www.googleapis.com/auth/pubsub"},
		Endpoint:     google.Endpoint,
	}

	url := conf.AuthCodeURL("state-token", oauth2.AccessTypeOffline)

	fmt.Println("Redirect URL: ", url)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}
func handleCallback(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	// Query parameters from the callback
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")

	// Verify state token
	if state != "state-token" {
		http.Error(w, "State token does not match", http.StatusBadRequest)
		return
	}

	var conf = &oauth2.Config{
		ClientID:     common_config.AuthClientId,
		ClientSecret: common_config.AuthClientSecret,
		RedirectURL:  "http://localhost:3000/auth/google/callback",
		Scopes:       []string{"email", "profile", "https://www.googleapis.com/auth/pubsub"},
		Endpoint:     google.Endpoint,
	}

	token, err := conf.Exchange(ctx, code)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Token contains the access token
	fmt.Fprintf(w, "Access Token: %s", token.AccessToken)
}

// RefreshTokenResponse represents the JSON response from the OAuth2 provider.
type RefreshTokenResponse struct {
	AccessToken  string    `json:"access_token"`
	ExpiresIn    int64     `json:"expires_in"`
	ExpiresAt    time.Time `json:"expires_at"`
	RefreshToken string    `json:"refresh_token"`
	//Scope        string `json:"scope"`
	TokenType string `json:"token_type"`
	IDToken   string `json:"id_token"`
	// Include other fields as necessary
}

func refreshToken(client *http.Client, refreshToken string) (*RefreshTokenResponse, error) {
	// The URL for the token endpoint will vary based on the OAuth2 provider.
	tokenEndpoint := "https://oauth2.googleapis.com/token"

	requestData := map[string]string{
		"client_id":     common_config.AuthClientId,
		"client_secret": common_config.AuthClientSecret,
		"refresh_token": refreshToken,
		"grant_type":    "refresh_token",
	}
	jsonValue, _ := json.Marshal(requestData)

	response, err := http.Post(tokenEndpoint, "application/json", bytes.NewBuffer(jsonValue))
	//response, err := client.Post(tokenEndpoint, "application/json", bytes.NewBuffer(jsonValue))
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		// Handle non-200 responses
		fmt.Println(response.StatusCode)
		return nil, err
	}

	var tokenResponse RefreshTokenResponse
	err = json.NewDecoder(response.Body).Decode(&tokenResponse)
	if err != nil {
		return nil, err
	}

	// Build time when Token expires
	var expireDuration time.Duration
	expireDuration = time.Duration(tokenResponse.ExpiresIn) * time.Second
	tokenResponse.ExpiresAt = time.Now().Add(expireDuration)

	return &tokenResponse, nil
}

// Start and run Local Web Server
func (gcp *GcpObjectStruct) startLocalWebServer(webServer *http.Server, url string) {

	var cmd *exec.Cmd

	/*
		switch runtime.GOOS {
		case "windows":
			// Command for Windows
			cmd = exec.Command("cmd", "/C", "start", "chrome", "--new-window", "--guest", url)
		case "darwin":
			// Command for macOS
			cmd = exec.Command("open", "-a", "Google Chrome", "--args", "--new-window", "--guest", url)
		case "linux":
			// Command for Linux
			cmd = exec.Command("google-chrome", "--new-window", "--guest", url)
		default:
			panic("Unsupported operating system")
		}

	*/
	// Determine the operating system
	switch runtime.GOOS {
	case "windows":
		// Command for Windows
		cmd = exec.Command("cmd", "/C", "start", "chrome", "--incognito", url)
	case "darwin":
		// Command for macOS
		cmd = exec.Command("open", "-a", "Google Chrome", "--args", "--incognito", url)
	case "linux":
		// Command for Linux
		cmd = exec.Command("google-chrome", "--incognito", url)
	default:
		panic("Unsupported operating system")
	}

	// Execute the command
	err := cmd.Start()
	if err != nil {
		panic(err)
	}

	// Print the PID of the process
	fmt.Printf("Chrome started with PID: %d\n", cmd.Process.Pid)

	// Kill the process
	//if err := cmd.Process.Kill(); err != nil {
	//	panic(err)
	//}
	//err := webbrowser.Open("http://localhost:3000")

	if err != nil {
		common_config.Logger.WithFields(logrus.Fields{
			"ID":  "17bc0305-4594-48e1-bb8d-c642579e5e56",
			"err": err,
		}).Fatalf("Couldn't open the web browser")
	}

	// Kill the process before leave
	//	defer func() {
	//		if err := cmd.Process.Kill(); err != nil {
	//			panic(err)
	//		}

	//		fmt.Println("Chrome process killed")
	//	}()

	err = webServer.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		common_config.Logger.WithFields(logrus.Fields{
			"ID": "8226cf74-0cdc-4e29-a441-116504b4b333",
		}).Fatalf("Local Web Server failed to listen: %s\n", err)

	}

	common_config.Logger.WithFields(logrus.Fields{
		"ID":  "844f2c3e-c271-4f95-ba9c-4eec9a206811",
		"err": err.Error(),
	}).Debug("Web Server was stopped")
}

// Navigate to new page
func (gcp *GcpObjectStruct) navigateToNewPage(url string) {

	var cmd *exec.Cmd

	// Determine the operating system
	switch runtime.GOOS {
	case "windows":
		// Command for Windows
		cmd = exec.Command("cmd", "/C", "start", "chrome", "--incognito", url)
	case "darwin":
		// Command for macOS
		cmd = exec.Command("open", "-a", "Google Chrome", "--args", "--incognito", url)
	case "linux":
		// Command for Linux
		cmd = exec.Command("google-chrome", "--incognito", url)
	default:
		panic("Unsupported operating system")
	}

	// Execute the command
	err := cmd.Start()
	if err != nil {
		panic(err)
	}

	// Print the PID of the process
	fmt.Printf("Chrome started with PID: %d\n", cmd.Process.Pid)

	if err != nil {
		common_config.Logger.WithFields(logrus.Fields{
			"ID":  "17bc0305-4594-48e1-bb8d-c642579e5e56",
			"err": err,
		}).Fatalf("Couldn't open the web browser")
	}

	// Kill the process before leave
	defer func() {
		if err := cmd.Process.Kill(); err != nil {
			panic(err)
		}

		fmt.Println("Chrome process killed")
	}()

}

// Close down Local Web Server
func (gcp *GcpObjectStruct) stopLocalWebServer(ctx context.Context, webServer *http.Server) {

	common_config.Logger.WithFields(logrus.Fields{
		"ID": "1f4e0354-2a09-4a1d-be61-67ecda781142",
	}).Debug("Trying to stop local web server")

	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	err := webServer.Shutdown(ctx)
	if err != nil {
		if err != nil {
			common_config.Logger.WithFields(logrus.Fields{
				"ID": "ea06dfab-39b9-4df6-b3ca-7f5f56b3cb91",
			}).Fatalf("Local Web Server Shutdown Failed:%+v", err)

		} else {
			common_config.Logger.WithFields(logrus.Fields{
				"ID": "ea06dfab-39b9-4df6-b3ca-7f5f56b3cb91",
			}).Debug("Local Web Server Exited Properly")
		}

	}

}

// SetLogger
// Set to use the same Logger reference as is used by central part of system
func (gcp *GcpObjectStruct) SetLogger(logger *logrus.Logger) {

	//grpcOutVaraible = GRPCOutStruct{}

	//gcp.logger = logger

	return

}

// initiateUserObject
// Set to use the same Logger reference as is used by central part of system
func (gcp *GcpObjectStruct) initiateUserObject() {

	// Only do initiation if it's not done before

	if gcp.gcpAccessTokenForAuthorizedAccounts.UserID == "" {
		gcp.gcpAccessTokenForAuthorizedAccounts = goth.User{}
	}

	return

}

// initiateUserObject
// Set to use the same Logger reference as is used by central part of system
func (gcp *GcpObjectStruct) initiateUserObjectPubSub() {

	// Only do initiation if it's not done before

	if gcp.gcpAccessTokenForAuthorizedAccountsPubSub.UserID == "" {
		gcp.gcpAccessTokenForAuthorizedAccountsPubSub = goth.User{}
	}

	return

}
