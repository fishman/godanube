//
// gocommon - Go library to interact with the JoyentCloud
//
//
// Copyright (c) 2013 Joyent Inc.
//
// Written by Daniele Stroppa <daniele.stroppa@joyent.com>
//

package client

import (
	//"fmt"
	"log"
	//"net/url"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/erigones/godanube/auth"
	danubehttp "github.com/erigones/godanube/http"
)

const (
	// The HTTP request methods.
	GET    = "GET"
	POST   = "POST"
	PUT    = "PUT"
	DELETE = "DELETE"
	HEAD   = "HEAD"
	COPY   = "COPY"
)

// Client implementations sends service requests to the Danube Cloud.
type Client interface {
	SendRequest(method, apiCall, rfc1123Date string, request *danubehttp.RequestData, response *danubehttp.ResponseData) (err error)
	SwitchVirtDC(virtDC string)
	GetVirtDC() string
	SetTrace(traceEnabled bool)
	GetTrace() bool
	// MakeServiceURL prepares a full URL to a service endpoint, with optional
	// URL parts. It uses the first endpoint it can find for the given service type.
	//DELME MakeServiceURL(parts []string) string
	SignURL(path string, expires time.Time) (string, error)
	Logger() *log.Logger
}

// This client sends requests without authenticating.
type client struct {
	creds      *auth.Credentials
	httpClient *danubehttp.Client
	mu         sync.Mutex
	logger     *log.Logger
}

var _ Client = (*client)(nil)

func newClient(credentials *auth.Credentials, httpClient *danubehttp.Client, logger *log.Logger) Client {
	client := client{creds: credentials, logger: logger, httpClient: httpClient}
	return &client
}

func NewClient(credentials *auth.Credentials, apiVersion string, logger *log.Logger) Client {
	sharedHttpClient := danubehttp.New(credentials, apiVersion, logger)
	return newClient(credentials, sharedHttpClient, logger)
}

func (c *client) sendRequest(method, url, rfc1123Date string, request *danubehttp.RequestData, response *danubehttp.ResponseData) (err error) {
	err = c.httpClient.JsonRequest(method, url, rfc1123Date, request, response)
	return err
	/*DELME
	if request.ReqValue != nil || response.RespValue != nil {
		err = c.httpClient.JsonRequest(method, url, rfc1123Date, request, response)
	} else {
		err = c.httpClient.BinaryRequest(method, url, rfc1123Date, request, response)
	}
	return
	*/
}

func (c *client) SendRequest(method, apiCall, rfc1123Date string, request *danubehttp.RequestData, response *danubehttp.ResponseData) (err error) {
	//DELME url := c.MakeServiceURL([]string{c.creds.UserAuthentication.User, apiCall})
	url := makeURL(c.creds.ApiEndpoint.URL, []string{apiCall})
	if c.creds.VirtDatacenter != "" {
		if request.Params != nil && request.Params.Get("dc") == "" {
			// set default VirtDatacenter in GET params
			request.Params.Set("dc", c.creds.VirtDatacenter)
		}
		//if request.ReqValue != nil && request.ReqValue.Dc == "" {
		if request.ReqValue != nil {
			// set default VirtDatacenter in data json (non-GET call)
			s := reflect.ValueOf(&request.ReqValue)
			/**
			The most common source of problems here is passing an object that has no 'Dc' member.
			(means ReqData is not inherited).
			*/
			//println(s.String())                      // should return <*interface {} Value>
			//println(s.Elem().String())               // should return <interface {} Value>
			//println(s.Elem().Elem().String())        // should return <*cloudapi. ... Value>
			//println(s.Elem().Elem().Elem().String()) // should return <cloudapi. ... Value>
			s = s.Elem().Elem().Elem()
			req_dc := s.FieldByName("Dc").Interface()
			//fmt.Println(req_dc)
			if req_dc == "" {
				s.FieldByName("Dc").SetString(c.creds.VirtDatacenter)
			}
		}
	}
	err = c.sendRequest(method, url, rfc1123Date, request, response)
	return err
}

func makeURL(base string, parts []string) string {
	if !strings.HasSuffix(base, "/") && len(parts) > 0 {
		base += "/"
	}
	var url string
	if len(parts) == 1 {
		url = base + parts[0]
	} else {
		url = base + strings.Join(parts, "/")
	}
	// make sure the call always ends with slash
	if !strings.HasSuffix(url, "/") {
		url += "/"
	}
	return url
}

/* DELME
func (c *client) MakeServiceURL(parts []string) string {
	return makeURL(c.baseURL, parts)
}
*/

func (c *client) SignURL(path string, expires time.Time) (string, error) {
	return "", nil
	/* DELME
	parsedURL, err := url.Parse(c.baseURL)
	if err != nil {
		return "", fmt.Errorf("bad Manta endpoint URL %q: %v", c.baseURL, err)
	}
	userAuthentication := c.creds.UserAuthentication
	userAuthentication.Algorithm = "RSA-SHA1"
	keyId := url.QueryEscape(fmt.Sprintf("/%s/keys/%s", userAuthentication.User, c.creds.MantaKeyId))
	params := fmt.Sprintf("algorithm=%s&expires=%d&keyId=%s", userAuthentication.Algorithm, expires.Unix(), keyId)
	signingLine := fmt.Sprintf("GET\n%s\n%s\n%s", parsedURL.Host, path, params)

	signature, err := auth.GetSignature(userAuthentication, signingLine)
	if err != nil {
		return "", fmt.Errorf("cannot generate URL signature: %v", err)
	}
	signedURL := fmt.Sprintf("%s%s?%s&signature=%s", c.baseURL, path, params, url.QueryEscape(signature))
	return signedURL, nil
	*/
}

func (c *client) SwitchVirtDC(virtDC string) {
	c.creds.VirtDatacenter = virtDC
}

func (c *client) GetVirtDC() string {
	return c.creds.VirtDatacenter
}

func (c *client) SetTrace(traceEnabled bool) {
	c.httpClient.SetTrace(traceEnabled)
}

func (c *client) GetTrace() bool {
	return c.httpClient.GetTrace()
}

func (c *client) Logger() *log.Logger {
	return c.logger
}
