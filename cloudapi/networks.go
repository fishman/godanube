package cloudapi

import (
	"github.com/erigones/godanube/client"
	"github.com/erigones/godanube/errors"
)

// Network represents a network available to a given account
type Network struct {
    GenericDcEntity
    Network     string          `json:"network,omitempty`
    Netmask     string          `json:"netmask,omitempty`
    Gateway     string          `json:"gateway,omitempty`
    Nic_tag     string          `json:"nic_tag,omitempty`
    Nic_tag_type string         `json:"nic_tag_type,omitempty`
    Vlan_id     *int            `json:"vlan_id,omitempty`
    Vxlan_id    *int            `json:"vxlan_id,omitempty`
    Mtu         *int            `json:"mtu,omitempty`
    Resolvers   []string        `json:"resolvers,omitempty`
    Dns_domain  string          `json:"dns_domain,omitempty`
    Ptr_domain  string          `json:"ptr_domain,omitempty`
    Dhcp_passthrough *bool       `json:"dhcp_passthrough,omitempty`
    Dcs         []string        `json:"dcs,omitempty"`      // vDC list where the network is attached
}

/*** STRUCTS FOR NETWORK-SPECIFIC DC RESPONSES ***/
type NetworkResponse struct {
	DcResponse
	Result Network	`json:"result"`
}

type NetworkResponseFull struct {
	DcResponse
	Result []Network	`json:"result"`
}

// ListNetworks lists all the networks
// This call needs SuperAdmin rights. With Admin rights use GetAttachedNetworks()
func (c *Client) ListNetworks() ([]string, error) {
//J
	var resp ResponseList
	req := request{
		method: client.GET,
		url:    "network",
		resp:   &resp,
	}
	if _, err := c.sendRequest(req); err != nil {
		return nil, errors.Newf2(err, resp.Detail, "failed to get list of networks")
	}
	return resp.Result, nil
}

// Returns extended info about all attached networks in current vDC
func (c *Client) GetAttachedNetworks() ([]Network, error) {
//J
	var resp NetworkResponseFull
	filter := NewFilter()
	filter.Set("full", "true")
	req := request{
		method: client.GET,
		filter: filter,
		url:    makeURL("dc", c.client.GetVirtDC(), "network"),
		resp:   &resp,
	}
	if _, err := c.sendRequest(req); err != nil {
		return nil, errors.Newf2(err, resp.Detail, "failed to get attached networks")
	}
	return resp.Result, nil
}

// GetNetwork retrieves an individual network info.
// This call needs SuperAdmin rights. With Admin rights use GetAttachedNetworks()
func (c *Client) GetNetwork(networkName string) (*Network, error) {
//J
	var resp NetworkResponse
	req := request{
		method: client.GET,
		url:    makeURL("network", networkName),
		resp:   &resp,
	}
	if _, err := c.sendRequest(req); err != nil {
		return nil, errors.Newf2(err, resp.Detail, "failed to get network info for \"%s\"", networkName)
	}
	return &resp.Result, nil
}
