/*
Package cloudapi interacts with the Cloud API (http://apidocs.joyent.com/cloudapi/).

Licensed under the Mozilla Public License version 2.0

Copyright (c) Joyent Inc.
*/
package cloudapi

import (
	"net/http"
	"net/url"
	"path"
	//"fmt"

	"github.com/erigones/godanube/client"
	jh "github.com/erigones/godanube/http"
)

const (
	// DefaultAPIVersion defines the default version of the Cloud API to use
	DefaultAPIVersion = "~4.2"

	// CloudAPI URL parts
	apiKeys                    = "keys"
	apiPackages                = "packages"
	apiImages                  = "images"
	apiDatacenters             = "datacenters"
	apiMachines                = "machines"
	apiMetadata                = "metadata"
	apiSnapshots               = "snapshots"
	apiTags                    = "tags"
	apiAnalytics               = "analytics"
	apiInstrumentations        = "instrumentations"
	apiInstrumentationsValue   = "value"
	apiInstrumentationsRaw     = "raw"
	apiInstrumentationsHeatmap = "heatmap"
	apiInstrumentationsImage   = "image"
	apiInstrumentationsDetails = "details"
	apiUsage                   = "usage"
	apiAudit                   = "audit"
	apiFirewallRules           = "fwrules"
	apiFirewallRulesEnable     = "enable"
	apiFirewallRulesDisable    = "disable"
	apiNetworks                = "networks"
	apiFabricVLANs             = "fabrics/default/vlans"
	apiFabricNetworks          = "networks"
	apiNICs                    = "nics"
	apiServices                = "services"

	// CloudAPI actions
	actionExport    = "export"
	actionStop      = "stop"
	actionStart     = "start"
	actionReboot    = "reboot"
	actionResize    = "resize"
	actionRename    = "rename"
	actionEnableFw  = "enable_firewall"
	actionDisableFw = "disable_firewall"
)

// Client provides a means to access the Joyent CloudAPI
// Final object that is returned to the caller by cloudapi.New() and interfaces all API calls by URL.
type Client struct {
	client client.Client
}

// New creates a new Client.
func New(client client.Client) *Client {
	return &Client{client}
}

// Filter represents a filter that can be applied to an API request.
type Filter struct {
	v url.Values
}

// NewFilter creates a new Filter.
func NewFilter() *Filter {
	return &Filter{make(url.Values)}
}

// Set a value for the specified filter.
func (f *Filter) Set(filter, value string) {
	f.v.Set(filter, value)
}

// Add a value for the specified filter.
func (f *Filter) Add(filter, value string) {
	f.v.Add(filter, value)
}

// request represents an API request
type request struct {
	method         string
	url            string
	filter         *Filter
	reqValue       interface{}
	reqHeader      http.Header
	resp           interface{}
	respHeader     *http.Header
	expectedStatus int
	expectedStatuses []int
}

//TODO presunut
/*** GENERIC DC STRUCTS ***/
type GenericDcEntity struct {
    Name         string    `json:"name,omitempty"`      // Unique readable identifier
    Alias        string    `json:"alias,omitempty"`     // Friendly name
    Uuid         string    `json:"uuid,omitempty"`
    Owner        string    `json:"owner,omitempty"`     // Object owner
    Access       int       `json:"access,omitempty"`    // public/private
    Desc         string    `json:"desc,omitempty"`      // Longer description
    Created      string    `json:"created,omitempty"`   // Object creation time
}
/*** STRUCTS FOR GENERIC DC RESPONSES ***/
type DcResponse struct {
	Status	string
	Task_id	string
	Detail	string
	//Result interface{}
}
type ResponseList struct {
	DcResponse
	Result []string	`json:"result"`
}

/*** STRUCTS FOR GENERIC DC REQUESTS ***/
type ReqData struct {
	Dc		string		`json:"dc,omitempty"`
	Force	bool		`json:"force,omitempty"`
}

/*
type StatusQuery struct {
	Url		string
	Status	int
	Method	string
	Text	TaskResponse
}
*/


//DELME start
/*
func (c *Client) DajVmList(url string) (*VmList, error) { //was: SendRequest
	var resp VmList
	myreq := request{
		method:     client.GET,
		url:        url,
		//reqHeader:  http.Header{},
		resp:		&resp,
	}
	c.sendRequest(myreq)
	return &resp, nil
}
/*DELME
func addUrlParams(params map[string]string) string {
	var buf string
	first := true
	for k, v := range params {
		if first {
			buf = "?"
		} else {
			buf += "&"
		}
		buf += k + "=" + v
	}
	return buf
}*/
//func (c *Client) WaitForTask(machineID string) (*DcResponse, error) {
//DELME end
//TODO implement timeouts
// Helper method to send an API request
func (c *Client) sendRequest(req request) (*jh.ResponseData, error) {
	request := jh.RequestData{}

	if req.method == client.GET {
		if req.filter == nil {
			request.Params = &NewFilter().v
		} else {
			request.Params = &req.filter.v
		}
	} else if req.reqValue == nil {
		req.reqValue = &ReqData{}
	}

	request.ReqValue = req.reqValue
	request.ReqHeaders = req.reqHeader

	if len(req.expectedStatuses) == 0 {
		if req.expectedStatus != 0 {
			req.expectedStatuses = []int{req.expectedStatus}
		} else {
			req.expectedStatuses = []int{http.StatusOK}
		}
	}
	respData := jh.ResponseData{
		RespValue:      req.resp,
		RespHeaders:    req.respHeader,
		ExpectedStatus: req.expectedStatuses,
	}
	err := c.client.SendRequest(req.method, req.url, "", &request, &respData)
	return &respData, err
}

// Helper method to create the API URL
func makeURL(parts ...string) string {
	return path.Join(parts...)
}

func (c *Client) SwitchVirtDC(virtDC string) {
	c.client.SwitchVirtDC(virtDC)
}

