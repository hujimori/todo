package sampleclient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
)

type Client struct {
	BaseURL    *url.URL
	HTTPClient *http.Client
	Token      string
	Logger     *log.Logger
}

func NewClient(rawBaseURL, token string, logger *log.Logger) (*Client, error) {
	baserURL, err := url.Parse(rawBaseURL)
	if err != nil {
		return nil, err
	}

	if logger == nil {
		logger = log.New(os.Stderr, "[LOG]", log.LstdFlags)
	}

	return &Client{
		BaseURL:    baserURL,
		HTTPClient: http.DefaultClient,
		Token:      token,
		Logger:     logger,
	}, nil
}

func (c *Client) GetUserItems(ctx context.Context, userID string, page, perPage int) ([]*Item, error) {
	reqURL := *c.BaseURL

	reqURL.Path = path.Join(reqURL.Path, "users", userID, "items")

	q := reqURL.Query()
	q.Add("page", strconv.Itoa(page))
	q.Add("per_page", strconv.Itoa(perPage))
	reqURL.RawQuery = q.Encode()

	req, err := http.NewRequest(http.MethodGet, reqURL.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "qiita-go-client")

	req = req.WithContext(ctx)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		var items []*Item
		if err := json.Unmarshal(bodyBytes, &items); err != nil {
			return nil, err
		}
		return items, nil
	case http.StatusBadRequest:
		return nil, errors.New("bad request. some parameters may be invalid")
	case http.StatusNotFound:
		return nil, fmt.Errorf("not found. user with id '%s' may not exist", userID)
	default:
		return nil, errors.New("unexpected error")
	}

	return nil, nil
}

type Item struct {
	ID         string `json:"id"`
	Title      string `json:"title"`
	LikesCount int    `json:"likes_count"`
}