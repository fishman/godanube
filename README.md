# godanube

`godanube` is a Go client for Danube Cloud

<!-- markdown-toc start - Don't edit this section. Run M-x markdown-toc-generate-toc again -->
**Table of Contents**

- [godanube](#godanube)
    - [Usage](#usage)
    - [Developing](#developing)
    - [License](#license)

<!-- markdown-toc end -->

## Usage

To create a client `*cloudapi.Client` you'll need:

- API key for Danube Cloud account (GUI: Profile -> API Keys)
- URL for the Danube Cloud installation (e.g. `https://1.2.3.4/api/` or `https://console.danube.cloud/api/`)
- Name of the [virtual datacenter](https://docs.danubecloud.org/user-guide/gui/datacenters/datacenters.html) (default is `main`). You can switch the virtual datacenter also later on.

Now you can initialize a client with the following:

```go
package main

import (
	"io/ioutil"
	"log"
	"os"

	"github.com/erigones/gocommon/client"
	"github.com/erigones/godanube/cloudapi"
	"github.com/erigones/gosign/auth"
)

func getclient(apiKey string, virtDc string, endpointUrl string) (*cloudapi.Client, error) {
	userAuth, err := auth.NewAuth("", "", apiKey)
	if err != nil {
		return nil, err
	}

	creds := &auth.Credentials{
		UserAuthentication: userAuth,
		ApiEndpoint:        auth.Endpoint{URL: endpointUrl},
	}

	return cloudapi.New(client.NewClient(
		creds.ApiEndpoint.URL,
		virtDc,
		cloudapi.DefaultAPIVersion,
		creds,
		log.New(os.Stderr, "", log.LstdFlags),
	)), nil
}

func main() {
	var c *cloudapi.Client
	baseUrl := "https://10.100.10.162:443/api/"
	c, _ = myclient("5996d65ce6963eb722ae667647517613", "main", baseUrl)

	resp, err := c.ListMachines()

	if(err != nil) {
		fmt.Println("error:" + err.Error())
		return
	}
	if len(resp) > 0 {
		fmt.Printf("The first machine is %s', resp[0])
	}
}

```

## Contributing

Report bugs and request features using [GitHub Issues](https://github.com/erigones/godanube/issues), or contribute code via a [GitHub Pull Request](https://github.com/erigones/godanube/pulls). Changes will be code reviewed before merging.


## Developing

This library assumes a Go development environment setup based on [How to Write Go Code](https://golang.org/doc/code.html). Your GOPATH environment variable should be pointed at your workspace directory.

You can now use `go get github.com/erigones/godanube` to install the repository to the correct location, but if you are intending on contributing back a change you may want to consider cloning the repository via git yourself. 

You can have create a `go.mod` file inside your project that will redirect the github repositories to your local filesytem. You don't need to have your project inside a `GOPATH`.

For example your directory tree might look like:

```
~/devel/
|_ godanube/		// cloned godanube repo
|_ myproject/		// your project
   |_ main.go
   |_ go.mod
```

And this is an example of `go.mod` file for this layout:
```
module main
require github.com/erigones/godanube v0.0.0
require github.com/erigones/godanube/client v0.0.0
require github.com/erigones/godanube/errors v0.0.0
require github.com/erigones/godanube/jpc v0.0.0
require github.com/erigones/godanube/testing v0.0.0
require github.com/erigones/godanube/cloudapi v0.0.0
require github.com/erigones/godanube/localservices v0.0.0
require github.com/erigones/godanube/localservices/hook v0.0.0
require github.com/erigones/godanube/auth v0.0.0
replace github.com/erigones/godanube/client => ../godanube/client
replace github.com/erigones/godanube/errors => ../godanube/errors
replace github.com/erigones/godanube/jpc => ../godanube/jpc
replace github.com/erigones/godanube/testing => ../godanube/testing
replace github.com/erigones/godanube/cloudapi => ../godanube/cloudapi
replace github.com/erigones/godanube/localservices => ../godanube/localservices
replace github.com/erigones/godanube/localservices/hook => ../godanube/localservices/hook
replace github.com/erigones/godanube => ../godanube
replace github.com/erigones/godanube/auth => ../godanube/auth
```

### Build the Library

```
cd ${GOPATH}/src/github.com/erigones/godanube
go build ./...
```

## License

godanube is licensed under the Mozilla Public License Version 2.0, a copy of which
is available at [LICENSE](LICENSE)
