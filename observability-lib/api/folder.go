package api

import (
	"fmt"

	"github.com/go-resty/resty/v2"
)

type Folder struct {
	ID    uint   `json:"id"`
	UID   string `json:"uid"`
	Title string `json:"title"`
}

// FindOrCreateFolder returns the folder by its name or creates it if it doesn't exist
func (c *Client) FindOrCreateFolder(name string) (*Folder, error) {
	folder, err := c.GetFolderByTitle(name)
	if err != nil {
		return nil, fmt.Errorf("could not find or create folder: %w", err)
	}
	if folder == nil {
		folder, _, err = c.PostFolder(name)
		if err != nil {
			return nil, fmt.Errorf("could not find create folder: %w", err)
		}
	}

	return folder, nil
}

// GetFolderByTitle Get a folder by title
func (c *Client) GetFolderByTitle(title string) (*Folder, error) {
	folders, _, err := c.GetFolders()
	if err != nil {
		return nil, err
	}
	for _, folder := range folders {
		if folder.Title == title {
			return &folder, nil
		}
	}

	return nil, nil
}

type GetAllFoldersResponse []Folder

// GetFolders Get all folders
func (c *Client) GetFolders() (GetAllFoldersResponse, *resty.Response, error) {
	var grafanaResp GetAllFoldersResponse

	resp, err := c.resty.R().
		SetHeader("Accept", "application/json").
		SetResult(&grafanaResp).
		SetQueryParam("limit", "100").
		Get("/api/folders")

	if err != nil {
		return GetAllFoldersResponse{}, resp, fmt.Errorf("error making API request: %w", err)
	}

	statusCode := resp.StatusCode()
	if statusCode != 200 {
		return GetAllFoldersResponse{}, resp, fmt.Errorf("error fetching folders, received unexpected status code %d: %s", statusCode, resp.String())
	}
	return grafanaResp, resp, nil
}

// PostFolder Create a new folder
func (c *Client) PostFolder(name string) (*Folder, *resty.Response, error) {
	var grafanaResp Folder

	resp, err := c.resty.R().
		SetHeader("Content-Type", "application/json").
		SetBody(struct {
			Title string `json:"title"`
		}{
			Title: name,
		}).
		SetResult(&grafanaResp).
		Post("/api/folders")

	if err != nil {
		return nil, resp, fmt.Errorf("error making API request: %w", err)
	}

	statusCode := resp.StatusCode()
	if statusCode != 200 {
		return nil, resp, fmt.Errorf("error creating folder, received unexpected status code %d: %s", statusCode, resp.String())
	}
	return &grafanaResp, resp, nil
}
