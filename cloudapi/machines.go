package cloudapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/erigones/godanube/client"
	"github.com/erigones/godanube/errors"
)


const (
    OsTypeLinux     = 1
    OsTypeSunOs     = 2
    OsTypeBSD       = 3
    OsTypeWindows   = 4
    OsTypeSunosZone = 5
    OsTypeLinuxZone = 6

    VmDeployTimeout = 300 / TaskQuerySleepTime   // seconds
	VmDeleteTimeout = 100 / TaskQuerySleepTime
)

// Machine represent a provisioned virtual machines
type VmDetails struct { // "result": [
	Hostname	string
	Uuid		string
	Alias		string
	Node		string
	Owner		string
	Status		string
	Node_status	string
	Vcpus		int
	Ram			int
	Disk		int
	Ips			[]string
	Uptime		int
	Locked		bool
	Tags		[]string
	Snapshot_define_inactive	int
	Snapshot_define_active		int
	Snapshots					int
	Backup_define_inactive		int
	Backup_define_active		int
	Backups						int
	Slaves						int
	Size_snapshots				int
	Size_backups				int
	Changed						bool
}

//DELME
type Machine struct {
	Id              string            // Unique identifier for the image
	Name            string            // Machine friendly name
	Type            string            // Machine type, one of 'smartmachine' or 'virtualmachine'
	State           string            // Current state of the machine
	Dataset         string            // The dataset URN the machine was provisioned with. For new images/datasets this value will be the dataset id, i.e, same value than the image attribute
	Memory          int               // The amount of memory the machine has (in Mb)
	Disk            int               // The amount of disk the machine has (in Gb)
	IPs             []string          // The IP addresses the machine has
	Metadata        map[string]string // Map of the machine metadata, e.g. authorized-keys
	Tags            map[string]string // Map of the machine tags
	Created         string            // When the machine was created
	Updated         string            // When the machine was updated
	Package         string            // The name of the package used to create the machine
	Image           string            // The image id the machine was provisioned with
	PrimaryIP       string            // The primary (public) IP address for the machine
	Networks        []string          // The network IDs for the machine
	FirewallEnabled bool              `json:"firewall_enabled"` // whether or not the firewall is enabled
	DomainNames     []string          `json:"dns_names"` // The domain names of this machine
}

// MachineDefinition represent the option that can be specified
// when creating a new machine.
// https://docs.danubecloud.org/api-reference/api/vm_define.html#post--vm-(hostname_or_uuid)-define
type MachineDefinition struct {
	// int values of zero (or bool false) mean Danube's default (because of omitempty)
    ReqData
	Name            string          `json:"name"`                 // VM name or UUID
    Alias           string          `json:"alias,omitempty"`      // DNS name without a domain
    DnsDomain       string          `json:"dns_domain,omitempty"` // DNS domain part
    Template        string          `json:"template,omitempty"`   // apply defaults from template
    OsType          int             `json:"ostype,omitempty"`
    Vcpus           int             `json:"vcpus,omitempty"`
    Ram             int             `json:"ram,omitempty"`        // in MB
    Note            string          `json:"note,omitempty"`
    Owner           string          `json:"owner,omitempty"`
    Node            string          `json:"node,omitempty"`
    Tags            []string        `json:"tags,omitempty"`
	//Tags            map[string]string `json:"-"`
    Monitored       bool            `json:"monitored,omitempty"`
    //MonitoredInternal bool         `json:"monitored_internal,omitempty"`
    Installed       bool            `json:"installed,omitempty"`// mark server installed (no deploy)
    SnapshotLimitManual int         `json:"snapshot_limit_manual,omitempty"`
    SnapshotSizeLimits  int         `json:"snapshot_size_limit,omitempty"`
    Zpool           string          `json:"zpool,omitempty"`
    CpuShares       int             `json:"cpu_shares,omitempty"`
    ZfsIoPriority   int             `json:"zfs_io_priority,omitempty"`
    CpuType         string          `json:"cpu_type,omitempty"`
    Vga             string          `json:"vga,omitempty"`
	Routes          map[string]string `json:"routes,omitempty"`
	MonitoringHostgroups []string   `json:"monitoring_hostgroups,omitempty"`
	MonitoringTemplates  []string   `json:"monitoring_templates,omitempty"`
	Mdata           map[string]string `json:"mdata,omitempty"`

    // Not settable, only for querying:
    Uuid            string
    Resolvers       []string
    Locked          bool
    Created         string
    Changed         bool
    Cpu_shares      int
}

type VmNicDefinition struct {
	// int values of zero (or bool false) mean Danube's default (because of omitempty)
    ReqData
	NicId                    int           `json:"nic_id,omitempty"`
	Net                      string        `json:"net,omitempty"`
	Ip                       string        `json:"ip,omitempty"`
	Model                    string        `json:"model,omitempty"`
	Dns                      bool          `json:"dns,omitempty"`
	UseNetDns                bool          `json:"use_net_dns,omitempty"`
	Mac                      string        `json:"mac,omitempty"`
	Mtu                      int           `json:"mtu,omitempty"`
	Primary                  bool          `json:"primary,omitempty"`
	AllowDhcpSpoofing        bool          `json:"allow_dhcp_spoofing,omitempty"`
	AllowIpSpoofing          bool          `json:"allow_ip_spoofing,omitempty"`
	AllowMacSpoofing         bool          `json:"allow_mac_spoofing,omitempty"`
	AllowRestrictedTraffic   bool          `json:"allow_restricted_traffic,omitempty"`
	AllowUnfilteredPromisc   bool          `json:"allow_unfiltered_promisc,omitempty"`
	AllowedIps               []string      `json:"allowed_ips,omitempty"`
	Monitoring               bool          `json:"monitoring,omitempty"`
	SetGateway               bool          `json:"set_gateway,omitempty"`

    // Not settable, only for querying:
	Netmask                  string        `json:"netmask,omitempty"`
}

type VmDiskDefinition struct {
    ReqData
	DiskId                   int           `json:"disk_id,omitempty"`
	Size                     int           `json:"size,omitempty"`
	Image                    string        `json:"image,omitempty"`
	Model                    string        `json:"model,omitempty"`
	BlockSize                int           `json:"block_size,omitempty"`
	UseNetDns                bool          `json:"use_net_dns,omitempty"`
	Compression              string        `json:"compression,omitempty"`
	Zpool                    string        `json:"zpool,omitempty"`
	Boot                     bool          `json:"boot,omitempty"`
	Refreservation           int           `json:"refreservation,omitempty"`
	ImageTagsInherit         bool          `json:"image_tags_inherit,omitempty"`
}

type CreateMachineOpts struct {
	Vm		MachineDefinition
	Disks	[]VmDiskDefinition
	Nics	[]VmNicDefinition
}

/*** STRUCTS FOR VM-SPECIFIC DC RESPONSES ***/
type VmResponse struct {
	DcResponse
	Result VmDetails            `json:"result"`
}

type VmsResponse struct {
	DcResponse
	Result []VmDetails            `json:"result"`
}

type CreateMachineResponse struct {
	DcResponse
	Result MachineDefinition    `json:"result"`
}

type VmNicResponse struct {
	DcResponse
	Result VmNicDefinition      `json:"result"`
}

type VmNicsResponse struct {
	DcResponse
	Result []VmNicDefinition    `json:"result"`
}

type VmDisksResponse struct {
	DcResponse
	Result []VmDiskDefinition   `json:"result"`
}

type VmDiskResponse struct {
	DcResponse
	Result VmDiskDefinition     `json:"result"`
}


// Equals compares two machines. Ignores state and timestamps.
func (m Machine) Equals(other Machine) bool {
	if m.Id == other.Id && m.Name == other.Name && m.Type == other.Type && m.Dataset == other.Dataset &&
		m.Memory == other.Memory && m.Disk == other.Disk && m.Package == other.Package && m.Image == other.Image &&
		m.compareIPs(other) && m.compareMetadata(other) {
		return true
	}
	return false
}

// Helper method to compare two machines IPs
func (m Machine) compareIPs(other Machine) bool {
	if len(m.IPs) != len(other.IPs) {
		return false
	}
	for i, v := range m.IPs {
		if v != other.IPs[i] {
			return false
		}
	}
	return true
}

// Helper method to compare two machines metadata
func (m Machine) compareMetadata(other Machine) bool {
	if len(m.Metadata) != len(other.Metadata) {
		return false
	}
	for k, v := range m.Metadata {
		if v != other.Metadata[k] {
			return false
		}
	}
	return true
}

// AuditAction represents an action/event accomplished by a machine.
type AuditAction struct {
	Action     string                 // Action name
	Parameters map[string]interface{} // Original set of parameters sent when the action was requested
	Time       string                 // When the action finished
	Success    string                 // Either 'yes' or 'no', depending on the action successfulness
	Caller     Caller                 // Account requesting the action
}

// Caller represents an account requesting an action.
type Caller struct {
	Type  string // Authentication type for the action request. One of 'basic', 'operator', 'signature' or 'token'
	User  string // When the authentication type is 'basic', this member will be present and include user login
	IP    string // The IP addresses this from which the action was requested. Not present if type is 'operator'
	KeyId string // When authentication type is either 'signature' or 'token', SSH key identifier
}

// appendJSON marshals the given attribute value and appends it as an encoded value to the given json data.
// The newly encode (attr, value) is inserted just before the closing "}" in the json data.
func appendJSON(data []byte, attr string, value interface{}) ([]byte, error) {
	newData, err := json.Marshal(&value)
	if err != nil {
		return nil, err
	}
	strData := string(data)
	result := fmt.Sprintf(`%s, "%s":%s}`, strData[:len(strData)-1], attr, string(newData))
	return []byte(result), nil
}

type jsonOpts MachineDefinition

/*
// MarshalJSON turns the given MachineDefinition into JSON
func (opts MachineDefinition) MarshalJSON() ([]byte, error) {
	jo := jsonOpts(opts)
	status := 
	if err != nil {
		return nil, err
	}
	for k, v := range opts.Tags {
		if !strings.HasPrefix(k, "tag.") {
			k = "tag." + k
		}
		data, err = appendJSON(data, k, v)
		if err != nil {
			return nil, err
		}
	}
	for k, v := range opts.Metadata {
		if !strings.HasPrefix(k, "metadata.") {
			k = "metadata." + k
		}
		data, err = appendJSON(data, k, v)
		if err != nil {
			return nil, err
		}
	}
	return data, nil
}
*/

// ListMachines lists all machines on record for an account.
// You can paginate this API by passing in offset, and limit
// See API docs: http://apidocs.joyent.com/cloudapi/#ListMachines
func (c *Client) ListMachines() ([]string, error) {
//J
	var resp ResponseList
	req := request{
		method: client.GET,
		url:    "vm",
		resp:   &resp,
	}
	if _, err := c.sendRequest(req); err != nil {
		return nil, errors.Newf2(err, resp.Detail, "failed to get list of machines")
	}
	return resp.Result, nil
}

func (c *Client) ListMachinesFilteredFull(vmfilter VmDetails) ([]VmDetails, error) {
//J
	var resp VmsResponse
	filter := NewFilter()
	filter.Set("extended", "true")
	req := request{
		method: client.GET,
		url:    "vm",
		filter:	filter,
		resp:   &resp,
	}
	if _, err := c.sendRequest(req); err != nil {
		return nil, errors.Newf2(err, resp.Detail, "failed to get list of machines")
	}

	allVMs := resp.Result
	var vmListFiltered []VmDetails

	// iterate over tags to filter VMs
	for _, vm := range allVMs {
		if
		(vmfilter.Hostname != "" && strings.Contains(vm.Hostname, vmfilter.Hostname)) ||
		(vmfilter.Uuid != "" && vmfilter.Uuid == vm.Uuid) ||
		(vmfilter.Alias != "" && strings.Contains(vm.Alias, vmfilter.Alias)) ||
		(vmfilter.Node != "" && strings.Contains(vm.Node, vmfilter.Node)) ||
		(vmfilter.Owner != "" && vmfilter.Owner == vm.Owner) ||
		(vmfilter.Status != "" && vmfilter.Status == vm.Status) {
			vmListFiltered = append(vmListFiltered, vm)

		} else if vmfilter.Tags != nil {
			someTagNotFound := false
			for _, tag := range vmfilter.Tags {
				found := false
				for _, vmtag := range vm.Tags {
					if vmtag == tag {
						found = true
						break
					}
				}
				if !found {
					someTagNotFound = true
					break
				}
			}
			if !someTagNotFound {
				vmListFiltered = append(vmListFiltered, vm)
			}
		}
	}

	if c.client.GetTrace() {
		c.client.Logger().Printf("Filtered VM list: %+v", GetVmUuids(vmListFiltered))
	}

	return vmListFiltered, nil
}

func GetVmUuids(vmDetails []VmDetails) []string {
	var vmList []string
	for _, vm := range vmDetails {
		vmList = append(vmList, vm.Uuid)
	}
	return vmList
}


// returns simplified version - only list of names
func (c *Client) ListMachinesFiltered(vmfilter VmDetails) ([]string, error) {
	vmListFull, err := c.ListMachinesFilteredFull(vmfilter)
	if err != nil {
		return nil, err
	} else {
		return GetVmUuids(vmListFull), nil
	}
}

// CountMachines returns the number of machines on record for an account.
// See API docs: http://apidocs.joyent.com/cloudapi/#ListMachines
func (c *Client) CountMachines() (int, error) {
	var resp int
	req := request{
		method: client.HEAD,
		url:    apiMachines,
		resp:   &resp,
	}
	if _, err := c.sendRequest(req); err != nil {
		return -1, errors.Newf(err, "failed to get count of machines")
	}
	return resp, nil
}

// GetMachine returns the machine specified by machineId.
// See API docs: http://apidocs.joyent.com/cloudapi/#GetMachine
/*
func (c *Client) GetMachine(machineID string) (*Machine, error) {
	var resp Machine
	req := request{
		method: client.GET,
		url:    makeURL(apiMachines, machineID),
		resp:   &resp,
	}
	if _, err := c.sendRequest(req); err != nil {
		return nil, errors.Newf(err, "failed to get machine with id: %s", machineID)
	}
	return &resp, nil
}*/
// DELME was GetVmExtended
func (c *Client) GetMachine(machineID string) (*VmDetails, error) {
//J
	var resp VmResponse
	filter := NewFilter()
	filter.Set("extended", "true")
	req := request{
		method:     client.GET,
		url:        fmt.Sprintf("%s/%s/", "vm", machineID),
		filter:		filter,
		resp:		&resp,
	}
	if _, err := c.sendRequest(req); err != nil {
		return nil, errors.Newf2(err, resp.Detail, "failed to get machine \"%s\"", machineID)
	}
	return &resp.Result, nil
}

func (c *Client) GetMachineState(machineID string) (*string, error) {
//J
	var resp VmResponse
	req := request{
		method:     client.GET,
		url:        fmt.Sprintf("%s/%s/status/", "vm", machineID),
		resp:		&resp,
	}
	if _, err := c.sendRequest(req); err != nil {
		return nil, errors.Newf2(err, resp.Detail, "failed to get machine \"%s\"", machineID)
	}
	return &resp.Result.Status, nil
}

func isTransientState(state string) bool {
	return strings.HasSuffix(state, "-")
}

func (c *Client) GetMachineNics(machineId string) ([]VmNicDefinition, error) {
//J
	var resp VmNicsResponse
	req := request{
		method: client.GET,
		url:    makeURL("vm", machineId, "define", "nic"),
		resp:   &resp,
	}
	if _, err := c.sendRequest(req); err != nil {
		return nil, errors.Newf2(err, resp.Detail, "failed to get nic info for machine \"%s\"", machineId)
	}
	return resp.Result, nil
}

func (c *Client) GetMachineDisks(machineId string) ([]VmDiskDefinition, error) {
//J
	var resp VmDisksResponse
	req := request{
		method: client.GET,
		url:    makeURL("vm", machineId, "define", "disk"),
		resp:   &resp,
	}
	if _, err := c.sendRequest(req); err != nil {
		return nil, errors.Newf2(err, resp.Detail, "failed to get disk info for machine \"%s\"", machineId)
	}
	return resp.Result, nil
}

// CreateMachine creates a new machine definition with the options specified.
func (c *Client) CreateMachineDefinition(opts MachineDefinition) (*MachineDefinition, error) {
//J
	var resp CreateMachineResponse
	req := request{
		method:         client.POST,
		//url:            fmt.Sprintf("vm/%s/define/", opts.Name),
		url:            makeURL("vm", opts.Name, "define"),
		reqValue:       &opts,
		resp:           &resp,
		expectedStatus: http.StatusCreated,
	}
	if _, err := c.sendRequest(req); err != nil {
		return nil, errors.Newf(err, "failed to create machine with name: %s", opts.Name)
	}
	return &resp.Result, nil
}

func (c *Client) AddMachineNicDefinition(machineID string, opts VmNicDefinition) (*VmNicDefinition, error) {
    errStr := "failed to create nic definition for machine: %s"
    nics, nicErr := c.GetMachineNics(machineID)
	if nicErr != nil {
		return nil, errors.Newf(nicErr, errStr, machineID)
	}
    nicCount := len(nics)

	var resp VmNicResponse

	req := request{
		method:         client.POST,
        // nics are counted from 1, we want to define the first unused (+1)
        url:            makeURL("vm", machineID, "define", "nic", fmt.Sprintf("%d", nicCount+1)),
		reqValue:       &opts,
		resp:           &resp,
		expectedStatus: http.StatusCreated,
	}
	if _, err := c.sendRequest(req); err != nil {
		return nil, errors.Newf(err, errStr, machineID)
	}
	return &resp.Result, nil
}

func (c *Client) AddMachineDiskDefinition(machineID string, opts VmDiskDefinition) (*VmDiskDefinition, error) {
    errStr := "failed to create disk definition for machine: %s"
    disks, diskErr := c.GetMachineDisks(machineID)
	if diskErr != nil {
		return nil, errors.Newf(diskErr, errStr, machineID)
	}
    diskCount := len(disks)

	var resp VmDiskResponse

	req := request{
		method:         client.POST,
        // disk are counted from 1, we want to define the first unused (+1)
        url:            makeURL("vm", machineID, "define", "disk", fmt.Sprintf("%d", diskCount+1)),
		reqValue:       &opts,
		resp:           &resp,
		expectedStatus: http.StatusCreated,
	}
	if _, err := c.sendRequest(req); err != nil {
		return nil, errors.Newf(err, errStr, machineID)
	}
	return &resp.Result, nil
}

// DeployMachine actualy deploys VM on a compute node
func (c *Client) DeployMachine(machineID string) (error) {
	var resp DcResponse
	req := request{
		method:         client.POST,
		url:            makeURL("vm", machineID),
		resp:           &resp,
		expectedStatuses: []int{http.StatusCreated, http.StatusOK},
	}
	if _, err := c.sendRequest(req); err != nil {
		return errors.Newf2(err, resp.Detail, "failed to deploy machine \"%s\"", machineID)
	}

	taskDetail, err := c.WaitForTaskStatus(resp.Task_id, "SUCCESS", VmDeployTimeout, req.expectedStatuses)
	if err != nil {
		return errors.Newf2(err, taskDetail.Message, "failed to deploy machine \"%s\"", machineID)
	}

	return nil
}

// ApplyMachineChanges actualy deploys VM on a compute node
func (c *Client) ApplyMachineChanges(machineID string) (error) {
    errMsg :=  "failed to apply machine settings \"%s\""
	var resp DcResponse
	req := request{
		method:         client.PUT,
		url:            makeURL("vm", machineID),
		resp:           &resp,
		expectedStatuses: []int{http.StatusCreated, http.StatusOK},
	}
	if _, err := c.sendRequest(req); err != nil {
		return errors.Newf2(err, resp.Detail, errMsg, machineID)
	}

	taskDetail, err := c.WaitForTaskStatus(resp.Task_id, "SUCCESS", VmDeleteTimeout, req.expectedStatuses)
	if err != nil {
		return errors.Newf2(err, taskDetail.Message, errMsg, machineID)
	}

	return nil
}

// CreateMachine creates a new machine with the options specified.
func (c *Client) CreateMachine(definition CreateMachineOpts) (*MachineDefinition, error) {
//J
	machine, err := c.CreateMachineDefinition(definition.Vm)
	if err != nil {
		if machine != nil {
			c.DeleteMachineDefinition(machine.Uuid)
		}
		return nil, err
	}

	for _, diskDef := range definition.Disks {
		if _, err := c.AddMachineDiskDefinition(machine.Uuid, diskDef); err != nil {
			c.DeleteMachineDefinition(machine.Uuid)
			return nil, err
		}
	}

	for _, nicDef := range definition.Nics {
		if _, err := c.AddMachineNicDefinition(machine.Uuid, nicDef); err != nil {
			c.DeleteMachineDefinition(machine.Uuid)
			return nil, err
		}
	}

	if err := c.DeployMachine(machine.Uuid); err != nil {
		c.DeleteMachineDefinition(machine.Uuid)
		return nil, err
	}

	return machine, nil
}

// StopMachine stops a running machine.
// See API docs: http://apidocs.joyent.com/cloudapi/#StopMachine
func (c *Client) StopMachine(machineID string, force bool) error {
//J
	if state, err := c.GetMachineState(machineID); err == nil && *state == "stopped" {
		return nil
	}
	var resp DcResponse
	var opts ReqData
	opts.Force = force
	req := request{
		method:         client.PUT,
		url:            fmt.Sprintf("vm/%s/status/stop/", machineID),
		expectedStatuses: []int{http.StatusCreated, http.StatusOK},
		reqValue:		&opts,
		resp:           &resp,
	}
	if _, err := c.sendRequest(req); err != nil {
		return errors.Newf2(err, resp.Detail, "failed to stop machine with id: %s", machineID)
	}

	_, err := c.WaitForTaskStatus(resp.Task_id, "SUCCESS", VmDeleteTimeout, req.expectedStatuses)
	if err != nil {
		return errors.Newf(err, "failed to stop machine \"%s\"", machineID)
	}

	return nil
}

// StartMachine starts a stopped machine.
// See API docs: http://apidocs.joyent.com/cloudapi/#StartMachine
func (c *Client) StartMachine(machineID string) error {
	var resp DcResponse
	req := request{
		method:         client.POST,
		url:            fmt.Sprintf("%s/%s?action=%s", apiMachines, machineID, actionStart),
		expectedStatus: http.StatusAccepted,
		resp:           &resp,
	}
	if _, err := c.sendRequest(req); err != nil {
		return errors.Newf2(err, resp.Detail, "failed to start machine with id: %s", machineID)
	}

	_, err := c.WaitForTaskStatus(resp.Task_id, "SUCCESS", VmDeleteTimeout, req.expectedStatuses)
	if err != nil {
		return errors.Newf(err, "failed to stop machine \"%s\"", machineID)
	}

	return nil
}

// RebootMachine reboots (stop followed by a start) a machine.
// See API docs: http://apidocs.joyent.com/cloudapi/#RebootMachine
func (c *Client) RebootMachine(machineID string) error {
	req := request{
		method:         client.POST,
		url:            fmt.Sprintf("%s/%s?action=%s", apiMachines, machineID, actionReboot),
		expectedStatus: http.StatusAccepted,
	}
	if _, err := c.sendRequest(req); err != nil {
		return errors.Newf(err, "failed to reboot machine with id: %s", machineID)
	}
	return nil
}

// ResizeMachine allows you to resize a SmartMachine. Virtual machines can also
// be resized, but only resizing virtual machines to a higher capacity package
// is supported.
// See API docs: http://apidocs.joyent.com/cloudapi/#ResizeMachine
func (c *Client) ResizeMachine(machineID, packageName string) error {
	req := request{
		method:         client.POST,
		url:            fmt.Sprintf("%s/%s?action=%s&package=%s", apiMachines, machineID, actionResize, packageName),
		expectedStatus: http.StatusAccepted,
	}
	if _, err := c.sendRequest(req); err != nil {
		return errors.Newf(err, "failed to resize machine with id: %s", machineID)
	}
	return nil
}

// RenameMachine renames an existing machine.
// See API docs: http://apidocs.joyent.com/cloudapi/#RenameMachine
func (c *Client) RenameMachine(machineID, machineName string) error {
	req := request{
		method:         client.POST,
		url:            fmt.Sprintf("%s/%s?action=%s&name=%s", apiMachines, machineID, actionRename, machineName),
		expectedStatus: http.StatusAccepted,
	}
	if _, err := c.sendRequest(req); err != nil {
		return errors.Newf(err, "failed to rename machine with id: %s", machineID)
	}
	return nil
}

// Delete machine data. Leaves the machine definition.
// Use DeleteMachineDefinition() for complete removal.
func (c *Client) DestroyMachine(machineID string) error {
//J
	var resp DcResponse
	req := request{
		method:         client.DELETE,
		url:            makeURL("vm", machineID),
		expectedStatuses: []int{http.StatusCreated, http.StatusOK}, // Pending, OK
		resp:           &resp,
	}

	if _, err := c.sendRequest(req); err != nil {
		return errors.Newf2(err, resp.Detail, "failed to delete machine \"%s\"", machineID)
	}

	_, err := c.WaitForTaskStatus(resp.Task_id, "SUCCESS", VmDeleteTimeout, req.expectedStatuses)
	if err != nil {
		return errors.Newf(err, "failed to delete machine \"%s\"", machineID)
	}

	return nil
}

// Deletes VM definition in DB. Machine must be in "notcreated" state.
func (c *Client) DeleteMachineDefinition(machineID string) error {
//J
	var resp DcResponse
	req := request{
		method:         client.DELETE,
		url:            makeURL("vm", machineID, "define"),
		expectedStatuses: []int{http.StatusOK},
		resp:           &resp,
	}

	if _, err := c.sendRequest(req); err != nil {
		return errors.Newf2(err, resp.Detail, "failed to delete machine \"%s\"", machineID)
	}

	return nil
}

func (c *Client) DeleteMachine(machineID string, force bool) error {
	errMsg := "failed to delete machine: " + machineID

	var status string
	for timeout := VmDeployTimeout; timeout > 0; timeout-=1 {
		vmStatus, err := c.GetMachineState(machineID)
		if err != nil {
			return errors.Newf(err, errMsg)
		}
		status = *vmStatus

		if
		status == "deploying" || status == "notready" ||
		status == "starting" || status == "stopping" ||
		strings.HasSuffix(status, "-") {
			// transient state, wait for finish
			time.Sleep(TaskQuerySleepTime * time.Second)
			continue
		} else {
			break
		}
	}

	stop := false
	destroy := false
	del := false

	if status == "running" {
		stop = true
		destroy = true
		del = true
	} else if status == "stopped" {
		destroy = true
		del = true
	} else if status == "notcreated" {
		del = true
	} else {
		return fmt.Errorf("Cannot delete machine \"%s\": invalid machine state", machineID)
	}

	if stop == true {
		if err := c.StopMachine(machineID, force); err != nil {
			return errors.Newf(err, errMsg)
		}
	}
	if destroy == true {
		if err := c.DestroyMachine(machineID); err != nil {
			return errors.Newf(err, errMsg)
		}
	}
	if del == true {
		if err := c.DeleteMachineDefinition(machineID); err != nil {
			return errors.Newf(err, errMsg)
		}
	}

	return nil
}

// MachineAudit provides a list of machine's accomplished actions, (sorted from
// latest to older one).
// See API docs: http://apidocs.joyent.com/cloudapi/#MachineAudit
func (c *Client) MachineAudit(machineID string) ([]AuditAction, error) {
	var resp []AuditAction
	req := request{
		method: client.GET,
		url:    makeURL(apiMachines, machineID, apiAudit),
		resp:   &resp,
	}
	if _, err := c.sendRequest(req); err != nil {
		return nil, errors.Newf(err, "failed to get actions for machine with id %s", machineID)
	}
	return resp, nil
}
