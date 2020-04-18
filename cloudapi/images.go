package cloudapi

import (
	"net/http"
	"time"

	"github.com/erigones/godanube/client"
	"github.com/erigones/godanube/errors"
)

const (
	imageDeleteTimeout = 120
	imageImportTimeout = 1800
)

// Image represent the software packages that will be available on newly provisioned machines
type Image struct {
	GenericDcEntity          // contains name, alias, uuid, owner, etc.
	Version         string   `json:"version,omitempty"` // Image version
	Type            string   `json:"type,omitempty"`    // Image type, one of 'smartmachine' or 'virtualmachine'
	Ostype          int      `json:"ostype,omitempty"`  // Underlying operating system
	Size            int      `json:"size,omitempty"`    // Starting disk size
	Resize          bool     `json:"resize,omitempty"`  // Resizable root disk at deploy
	Deploy          bool     `json:"deploy,omitempty"`  //
	Tags            []string `json:"tags,omitempty"`    // An array of associated img tags
	Status          int      `json:"status,omitempty"`  // Current image state. One of 'ok' or 'pending'.
	Dcs             []string `json:"-"`                 // vDC list where the image is attached

	/* only for ListImagesInVdc() */
	DcBound bool `json:"dc_bound,omitempty"` // Whether the image is dedicated to one vDC

	/* only for GetRemoteImageInfo() */
	Manifest	interface{}	`json:"manifest,omitempty"`

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

type ImageRepo struct {
	Url        string    `json:"url,omitempty"`
	Name       string    `json:"name,omitempty""`
	ImageCount int       `json:"image_count,omitempty"`
	LastUpdate time.Time `json:"last_update,omitempty"`
	Error      string    `json:"error,omitempty"`
}
/*** STRUCTS FOR IMAGE-SPECIFIC DC RESPONSES ***/
type ImageResponse struct {
	DcResponse
	Result Image `json:"result"`
}

type ImageResponseFull struct {
	DcResponse
	Result []Image `json:"result"`
}

type ImageRepoResponse struct {
	DcResponse
	Result ImageRepo `json:"result"`
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

// ListImages provides a list of image names available in the Danube Cloud.
// This call needs SuperAdmin rights. With Admin rights use ListAttachedImages()
func (c *Client) ListImages() ([]string, error) {
	//J
	var resp ResponseList
	//filter := NewFilter()
	//filter.Set("extended", "true")
	req := request{
		method: client.GET,
		url:    "image",
		//filter: filter,
		expectedStatuses: []int{http.StatusOK},
		resp:             &resp,
	}
	if _, err := c.sendRequest(req); err != nil {
		return nil, errors.Newf2(err, resp.Detail, "failed to get list of images")
	}
	return resp.Result, nil
}

func (c *Client) ListAttachedImages() ([]Image, error) {
	//J
	var resp ImageResponseFull
	filter := NewFilter()
	filter.Set("full", "true")
	req := request{
		method:           client.GET,
		url:              makeURL("dc", c.client.GetVirtDC(), "image"),
		filter:           filter,
		expectedStatuses: []int{http.StatusOK},
		resp:             &resp,
	}
	if _, err := c.sendRequest(req); err != nil {
		return nil, errors.Newf2(err, resp.Detail, "failed to get list of images")
	}
	return resp.Result, nil
}

// GetImage returns the image details.
// This call needs SuperAdmin rights. With Admin rights use GetAttachedImage()
func (c *Client) GetImage(imageName string) (*Image, error) {
	//J
	var resp ImageResponse
	req := request{
		method:           client.GET,
		url:              makeURL("image", imageName),
		expectedStatuses: []int{http.StatusOK},
		resp:             &resp,
	}
	if _, err := c.sendRequest(req); err != nil {
		return nil, errors.Newf2(err, resp.Detail, "failed to get image info for \"%s\"", imageName)
	}
	return &resp.Result, nil
}

// GetAttachedImage returns the details of the image that is attached in the active virtual datacenter.
func (c *Client) GetAttachedImage(imageName string) (*Image, error) {
	//J
	var resp ImageResponse
	req := request{
		method:           client.GET,
		url:              makeURL("dc", c.client.GetVirtDC(), "image", imageName),
		expectedStatuses: []int{http.StatusOK},
		resp:             &resp,
	}
	if _, err := c.sendRequest(req); err != nil {
		return nil, errors.Newf2(err, resp.Detail, "failed to get image info for \"%s\"", imageName)
	}
	return &resp.Result, nil
}

// GetRemoteImageInfo returns the details of the remote image that is present in the specified repo
func (c *Client) GetRemoteImageInfo(imageUuid, repoName string) (*Image, error) {
	//J
	var resp ImageResponse
	req := request{
		method:           client.GET,
		url:              makeURL("imagestore", repoName, "image", imageUuid),
		expectedStatuses: []int{http.StatusOK},
		resp:             &resp,
	}
	if _, err := c.sendRequest(req); err != nil {
		return nil, errors.Newf2(err, resp.Detail, "failed to get image info for \"%s\"", imageUuid)
	}
	return &resp.Result, nil
}

// DeleteImage Delete the image specified by name.
func (c *Client) DeleteImage(imageName string) error {
	//J
	var resp ImageResponse
	req := request{
		method:           client.DELETE,
		url:              makeURL("image", imageName),
		expectedStatuses: []int{http.StatusCreated, http.StatusOK},
		resp:             &resp,
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

func (c *Client) ListImgRepos() ([]string, error) {
	//J
	var resp ResponseList
	req := request{
		method:           client.GET,
		url:              "imagestore",
		expectedStatuses: []int{http.StatusOK},
		resp:             &resp,
	}
	if _, err := c.sendRequest(req); err != nil {
		return nil, errors.Newf2(err, resp.Detail, "failed to get list of configured repos")
	}
	return resp.Result, nil
}

func (c *Client) ListRemoteImages(repoName string) ([]string, error) {
	//J
	var resp ResponseList
	req := request{
		method:           client.GET,
		url:              makeURL("imagestore", repoName, "image"),
		expectedStatuses: []int{http.StatusOK},
		resp:             &resp,
	}
	if _, err := c.sendRequest(req); err != nil {
		return nil, errors.Newf2(err, resp.Detail, "failed to get list of remote images")
	}
	return resp.Result, nil
}

func (c *Client) GetImgRepo(repoName string) (*ImageRepo, error) {
	//J
	var resp ImageRepoResponse
	req := request{
		method:           client.GET,
		url:              makeURL("imagestore", repoName),
		expectedStatuses: []int{http.StatusOK},
		resp:             &resp,
	}
	if _, err := c.sendRequest(req); err != nil {
		return nil, errors.Newf2(err, resp.Detail, "failed to get repository info")
	}
	return &resp.Result, nil
}

func (c *Client) RefreshImgRepo(repoName string) error {
	//J
	var resp DcResponse
	req := request{
		method:           client.PUT,
		url:              makeURL("imagestore", repoName),
		expectedStatuses: []int{http.StatusOK},
		resp:             &resp,
	}
	if _, err := c.sendRequest(req); err != nil {
		return errors.Newf2(err, resp.Detail, "failed to refresh repository")
	}
	return nil
}

// https://docs.danubecloud.org/api-reference/api/image_base.html#post--image-(name)
func (c *Client) ImportImage(remoteImageUuid, newImageName, repoName string) error {
	//J
	var resp DcResponse

	type importRequest struct {
		ReqData
		GenericDcEntity
	}

	// if we want to support all options of the API call, we need a separate struct
	var opts importRequest
	opts.Name = newImageName // TODO check for empty
	opts.Access = AccessPublic

	req := request{
		method:           client.POST,
		url:              makeURL("imagestore", repoName, "image", remoteImageUuid),
		reqValue:         &opts,
		expectedStatuses: []int{http.StatusOK, http.StatusCreated},
		resp:             &resp,
	}
	if _, err := c.sendRequest(req); err != nil {
		return errors.Newf2(err, resp.Detail, "failed to import image \"%s\"", remoteImageUuid)
	}

	_, err := c.WaitForTaskStatus(resp.Task_id, "SUCCESS", imageImportTimeout, req.expectedStatuses)
	if err != nil {
		return errors.Newf(err, "failed to stop machineimport image \"%s\"", remoteImageUuid)
	}

	return nil
}
