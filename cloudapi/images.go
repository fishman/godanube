package cloudapi

import (
	"net/http"

	"github.com/erigones/godanube/client"
	"github.com/erigones/godanube/errors"
)

const (
    imageDeleteTimeout  = 120
    imageImportTimeout  = 1800
)

// Image represent the software packages that will be available on newly provisioned machines
type Image struct {
    GenericDcEntity         // contains name, alias, uuid, owner, etc.
    Version      string    `json:"version,omitempty"`   // Image version
    Type         string    `json:"type,omitempty"`      // Image type, one of 'smartmachine' or 'virtualmachine'
    Ostype       int       `json:"ostype,omitempty"`    // Underlying operating system
    Size         int       `json:"size,omitempty"`      // Starting disk size
    Resize       bool      `json:"resize,omitempty"`    // Resizable root disk at deploy?
    Deploy       bool      `json:"deploy,omitempty"`    // 
    Tags         []string  `json:"tags,omitempty"`      // An array of associated img tags
    Status       int       `json:"status,omitempty"`    // Current image state. One of 'ok' or 'pending'.
    Dcs          []string  `json:"-"`      // vDC list where the image is attached

    /*DELME
	Requirements map[string]interface{} // Minimum requirements for provisioning a machine with this image, e.g. 'password' indicates that a password must be provided
	Homepage     string                 // URL for a web page including detailed information for this image (new in API version 7.0)
	PublishedAt  string                 `json:"published_at"` // Time this image has been made publicly available (new in API version 7.0)
	Public       bool                   // Indicates if the image is publicly available (new in API version 7.1)
	Tags         map[string]string      // A map of key/value pairs that allows clients to categorize images by any given criteria (new in API version 7.1)
	EULA         string                 // URL of the End User License Agreement (EULA) for the image (new in API version 7.1)
	ACL          []string               // An array of account UUIDs given access to a private image. The field is only relevant to private images (new in API version 7.1)
    */
}

/*** STRUCTS FOR IMAGE-SPECIFIC DC RESPONSES ***/
type ImageResponse struct {
	DcResponse
	Result Image	`json:"result"`
}

/* DELME
// ExportImageOpts represent the option that can be specified
// when exporting an image.
type ExportImageOpts struct {
	MantaPath string `json:"manta_path"` // The Manta path prefix to use when exporting the image
}

// MantaLocation represent the properties that allow a user
// to retrieve the image file and manifest from Manta
type MantaLocation struct {
	MantaURL     string `json:"manta_url"`     // Manta datacenter URL
	ImagePath    string `json:"image_path"`    // Path to the image
	ManifestPath string `json:"manifest_path"` // Path to the image manifest
}*/

// CreateImageFromMachineOpts represent the option that can be specified
// when creating a new image from an existing machine.
type CreateImageFromMachineOpts struct {
	Machine     string            `json:"machine"`               // The machine UUID from which the image is to be created
	Name        string            `json:"name"`                  // Image name
	Version     string            `json:"version"`               // Image version
	Description string            `json:"description,omitempty"` // Image description
	Homepage    string            `json:"homepage,omitempty"`    // URL for a web page including detailed information for this image
	EULA        string            `json:"eula,omitempty"`        // URL of the End User License Agreement (EULA) for the image
	ACL         []string          `json:"acl,omitempty"`         // An array of account UUIDs given access to a private image. The field is only relevant to private images
	Tags        map[string]string `json:"tags,omitempty"`        // A map of key/value pairs that allows clients to categorize images by any given criteria
}

// ListImages provides a list of image names available in the datacenter.
func (c *Client) ListImages() ([]string, error) {
//J
	var resp ResponseList
	//filter := NewFilter()
	//filter.Set("extended", "true")
	req := request{
		method: client.GET,
		url:    "image",
		//filter: filter,
		resp:   &resp,
	}
	if _, err := c.sendRequest(req); err != nil {
		return nil, errors.Newf2(err, resp.Detail, "failed to get list of images")
	}
	return resp.Result, nil
}

// GetImage returns the image specified by imageId.
// See API docs: http://apidocs.joyent.com/cloudapi/#GetImage
func (c *Client) GetImage(imageName string) (*Image, error) {
//J
	var resp ImageResponse
	req := request{
		method: client.GET,
		url:    makeURL("image", imageName),
		resp:   &resp,
	}
	if _, err := c.sendRequest(req); err != nil {
		return nil, errors.Newf2(err, resp.Detail, "failed to get image info for \"%s\"", imageName)
	}
	return &resp.Result, nil
}

// DeleteImage (Beta) Delete the image specified by imageId. Must be image owner to do so.
// See API docs: http://apidocs.joyent.com/cloudapi/#DeleteImage
func (c *Client) DeleteImage(imageName string) error {
//J
	var resp ImageResponse
	req := request{
		method:         client.DELETE,
		url:            makeURL("image", imageName),
		expectedStatuses: []int{http.StatusCreated, http.StatusOK},
		resp:   &resp,
	}
	if _, err := c.sendRequest(req); err != nil {
		return errors.Newf2(err, resp.Detail, "failed to delete image \"%s\"", imageName)
	}

	_, err := c.WaitForTaskStatus(resp.Task_id, "SUCCESS", imageDeleteTimeout, req.expectedStatuses)
	if err != nil {
		return errors.Newf(err, "failed to delete image \"%s\"", imageName)
	}

	return nil
}

/*
func (c *Client) CreateImageFromMachine(opts CreateImageFromMachineOpts) (*Image, error) {
	var resp Image
	req := request{
		method:         client.POST,
		url:            apiImages,
		reqValue:       opts,
		resp:           &resp,
		expectedStatus: http.StatusCreated,
	}
	if _, err := c.sendRequest(req); err != nil {
		return nil, errors.Newf(err, "failed to create image from machine %s", opts.Machine)
	}
	return &resp, nil
}
*/
