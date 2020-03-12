package github

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/go-kit/kit/endpoint"
	kithttp "github.com/go-kit/kit/transport/http"
)

type Config struct {
	Project string
	Owner   string
	Token   string
}

func (c *Config) Validate() error {
	if c.Owner == "" {
		return errors.New("invalid owner")
	}
	if c.Token == "" {
		return errors.New("invalid token")
	}
	if c.Project == "" {
		return errors.New("invalid project")
	}
	return nil
}

type Client struct {
	f       endpoint.Endpoint
	owner   string
	project string
}

func New(cfg *Config) (*Client, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	tgt, err := url.Parse("https://api.github.com")
	if err != nil {
		return nil, err
	}
	options := []kithttp.ClientOption{
		kithttp.ClientBefore(
			kithttp.SetRequestHeader("Authorization", "Bearer "+cfg.Token),
		),
	}
	c := &Client{
		f: kithttp.NewClient(
			http.MethodPut,
			tgt,
			encoder,
			decoder,
			options...).Endpoint(),
		owner:   cfg.Owner,
		project: cfg.Project,
	}
	return c, nil
}

type fileRequest struct {
	project string
	owner   string
	path    string
	data    file
}

type file struct {
	Message   string `json:"message"`
	Committer struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	} `json:"committer"`
	Content string `json:"content"`
}

func encoder(_ context.Context, r *http.Request, request interface{}) error {
	req := request.(fileRequest)
	r.URL.Path = "/repos/" + req.owner + "/" + req.project + "/contents/" + req.path
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(req.data); err != nil {
		return err
	}
	r.Body = ioutil.NopCloser(&buf)
	r.Header.Set("Content-Type", "application/json")
	return nil
}

type fileResponse struct {
	message string
}

func decoder(_ context.Context, r *http.Response) (response interface{}, err error) {
	if r.StatusCode < 200 || r.StatusCode > 299 {
		f := fileResponse{}
		err := json.NewDecoder(r.Body).Decode(&f)
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("failed response: %v", f.message)
	}
	return nil, nil
}

func (c *Client) CreateFile(ctx context.Context, path, author, email, message, content string) error {
	s := base64.StdEncoding.EncodeToString([]byte(content))
	_, err := c.f(ctx, fileRequest{
		project: c.project,
		owner:   c.owner,
		path:    path,
		data: file{
			Message: message,
			Committer: struct {
				Name  string `json:"name"`
				Email string `json:"email"`
			}{
				Name:  author,
				Email: email,
			},
			Content: s,
		},
	})
	if err != nil {
		return err
	}
	return nil
}
