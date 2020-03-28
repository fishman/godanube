package cloudapi

import (
	"net/http"
	"fmt"
	"time"

	"github.com/erigones/godanube/client"
	"github.com/erigones/godanube/errors"
)

const (
	TaskQuerySleepTime = 2	// sec
	VmSnapTimeout	= 30 / TaskQuerySleepTime
)

type TaskInfo struct {
	Message		string
	Returncode	int
	Detail		string
}
type TaskResponse struct {
	DcResponse
	Result	TaskInfo
}

// queries for executed task status
func (c *Client) GetTaskInfo(taskId string) (*TaskResponse, error) {
	var resp TaskResponse
	req := request{
		method:     client.GET,
		expectedStatuses: []int{http.StatusCreated, http.StatusOK},
		url:        fmt.Sprintf("%s/%s/status/", "task", taskId),
		resp:		&resp,
	}
	if _, err := c.sendRequest(req); err != nil {
		return nil, errors.Newf2(err, resp.Detail, "failed to get info for task \"%s\"", taskId)
	}
	return &resp, nil
}

func (c *Client) GetRunningTasks() ([]string, error) {
	var resp []string
	req := request{
		method:     client.GET,
		expectedStatuses: []int{http.StatusOK},
		url:        "task",
		resp:		&resp,
	}
	if _, err := c.sendRequest(req); err != nil {
		return nil, errors.Newf(err, "failed to get running tasks")
	}
	return resp, nil
}

func (c *Client) WaitForTaskStatus(taskId, targetStatus string, timeoutSec uint, validHTTPStatuses []int) (*TaskInfo, error) {
	var resp TaskResponse
	req := request{
		method:     client.GET,
		expectedStatuses: validHTTPStatuses,
		url:        fmt.Sprintf("%s/%s/status/", "task", taskId),
		resp:		&resp,
	}
	for  {
		_, err := c.sendRequest(req)
		if err != nil {
			return &resp.Result, err
		} else if resp.Status == targetStatus {
			return &resp.Result, nil
		} else if resp.Status == "FAILED" {
				return &resp.Result, errors.Newf(nil, "Task \"%s\" has failed", taskId)
		} else {
			time.Sleep(TaskQuerySleepTime * time.Second)
			if(timeoutSec <= 0) {
				return &resp.Result, errors.Newf(nil, "Timed out waiting for task \"%s\"", taskId)
			} else {
				timeoutSec-=1
			}
		}
	}
}

func (c *Client) CancelTask(taskId string, force bool) (*TaskResponse, error) {
	var resp TaskResponse
	var opts ReqData
	opts.Force = force
	req := request{
		method:     client.PUT,
		expectedStatuses: []int{http.StatusCreated, http.StatusOK, http.StatusGone},
		url:        fmt.Sprintf("%s/%s/cancel/", "task", taskId),
		resp:		&resp,
		reqValue:	&opts,
	}
	if _, err := c.sendRequest(req); err != nil {
		return nil, errors.Newf2(err, resp.Detail, "failed to get cancel task \"%s\"", taskId)
	}

	/*
	_, err := c.WaitForTaskStatus(resp.Task_id, "REVOKED", 10, req.expectedStatuses)
	if err != nil {
		return nil, errors.Newf2(err, resp.Detail, "failed to cancel task \"%s\"", taskId)
	}
	*/

	return &resp, nil
}




type CreateSnapshotOpts struct {
	ReqData
	MachineID	string	`json:"-"`
	SnapName	string	`json:"-"`
	Disk_id		int		`json:"disk_id,omitempty"`
	Note		string	`json:"note,omitempty"`
	FsFreeze	bool	`json:"fs_freeze,omitempty"`
}

//TODO prelozit
// Input: *CreateSnapshotOpts
// Output: *TaskInfo
func (c *Client) CreateSnap(opts *CreateSnapshotOpts) (*TaskInfo, error) {
	/*
	var opts CreateSnapshotOpts
	if note != "" {
		opts.Note = note
	}
	//opts.Dc = "main"
	*/
	var resp DcResponse
	req := request{
		method:     client.POST,
		expectedStatuses: []int{http.StatusCreated, http.StatusOK},
		url:        fmt.Sprintf("vm/%s/snapshot/%s/", opts.MachineID, opts.SnapName),
		reqValue:   opts,
		resp:		&resp,
	}
	if _, err := c.sendRequest(req); err != nil {
		return nil, errors.Newf2(err, resp.Detail, "failed to create snapshot \"%s\" for \"%s\"", opts.SnapName, opts.MachineID)
	}

	taskResult, err := c.WaitForTaskStatus(resp.Task_id, "SUCCESS", VmSnapTimeout, req.expectedStatuses)
	if err != nil {
		return nil, errors.Newf2(err, resp.Detail, "failed to create snapshot \"%s\" for \"%s\"", opts.SnapName, opts.MachineID)
	}

	return taskResult, nil
}

