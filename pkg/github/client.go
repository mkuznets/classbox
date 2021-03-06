package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/pkg/errors"
	"golang.org/x/oauth2"
)

const (
	baseUrl = "https://api.github.com"
)

type Client struct {
	http  *http.Client
	token *oauth2.Token
}

func New(token *oauth2.Token) *Client {
	return &Client{
		http:  &http.Client{},
		token: token,
	}
}

type ErrorResponse struct {
	Response *http.Response
	Message  string `json:"message"`
}

func (e *ErrorResponse) Error() string {
	return e.Message
}

func (e *ErrorResponse) NotFound() bool {
	return e.Response.StatusCode == http.StatusNotFound
}

func checkResponse(r *http.Response) error {
	if c := r.StatusCode; 200 <= c && c <= 299 {
		return nil
	}
	e := &ErrorResponse{Response: r}
	data, err := ioutil.ReadAll(r.Body)
	if err == nil && data != nil {
		//noinspection GoUnhandledErrorResult
		if err := json.Unmarshal(data, &e); err != nil {
			return errors.WithMessage(err, "could not parse error response")
		}
	}
	return e
}

func (c *Client) Request(ctx context.Context, method string, path string, body []byte, acceptHeader string) ([]byte, error) {

	url := fmt.Sprintf(baseUrl + path)
	buf := bytes.NewBuffer(body)

	req, err := http.NewRequestWithContext(ctx, method, url, buf)
	if err != nil {
		return nil, errors.WithMessage(err, "could not create Request")
	}
	c.token.SetAuthHeader(req)

	if acceptHeader != "" {
		req.Header.Set("Accept", acceptHeader)
	} else {
		req.Header.Set("Accept", "application/vnd.github.v3+json")
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, errors.WithMessage(err, "could not send Request")
	}

	//noinspection GoUnhandledErrorResult
	defer resp.Body.Close()

	if err := checkResponse(resp); err != nil {
		// return nil, errors.WithMessagef(err, "HTTP error on %s %s", method, path)
		return nil, err
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.WithMessage(err, "could not read the response")
	}
	return data, nil
}

func (c *Client) RevokeOAuth(ctx context.Context, clientID, clientSecret string) error {

	url := fmt.Sprintf("%s/applications/%s/grant", baseUrl, clientID)
	buf := bytes.NewBuffer([]byte(fmt.Sprintf(`{"access_token":"%s"}`, c.token.AccessToken)))

	req, err := http.NewRequestWithContext(ctx, "DELETE", url, buf)
	if err != nil {
		return errors.Wrap(err, "could not create Request")
	}
	req.Header.Set("Accept", "application/vnd.github.doctor-strange-preview+json")
	req.SetBasicAuth(clientID, clientSecret)

	resp, err := c.http.Do(req)
	if err != nil {
		return errors.Wrap(err, "could not send Request")
	}

	//noinspection GoUnhandledErrorResult
	defer resp.Body.Close()

	err = checkResponse(resp)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) User(ctx context.Context) (*User, error) {
	data, err := c.Request(ctx, "GET", "/user", nil, "")
	if err != nil {
		return nil, err
	}

	user := User{}
	err = json.Unmarshal(data, &user)
	if err != nil {
		return &user, errors.Wrap(err, "could not decode response")
	}

	return &user, nil
}

func (c *Client) InstallationByLogin(ctx context.Context, login string) (*Installation, error) {
	path := fmt.Sprintf("/users/%s/installation", login)
	data, err := c.Request(ctx, "GET", path, nil, "application/vnd.github.machine-man-preview+json")
	if err != nil {
		return nil, err
	}

	inst := Installation{}
	err = json.Unmarshal(data, &inst)
	if err != nil {
		return &inst, errors.Wrap(err, "could not decode response")
	}

	return &inst, nil
}

func (c *Client) InstallationByID(ctx context.Context, instID uint64) (*Installation, error) {
	path := fmt.Sprintf("/app/installations/%d", instID)
	data, err := c.Request(ctx, "GET", path, nil, "application/vnd.github.machine-man-preview+json")
	if err != nil {
		return nil, err
	}

	inst := Installation{}
	err = json.Unmarshal(data, &inst)
	if err != nil {
		return &inst, errors.Wrap(err, "could not decode response")
	}

	return &inst, nil
}

func (c *Client) Uninstall(ctx context.Context, instID int) error {
	path := fmt.Sprintf("/app/installations/%d", instID)
	_, err := c.Request(ctx, "DELETE", path, nil, "application/vnd.github.gambit-preview+json")
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) AuthAsInstallation(ctx context.Context, instID int) error {
	path := fmt.Sprintf("/app/installations/%d/access_tokens", instID)
	data, err := c.Request(ctx, "POST", path, nil, "application/vnd.github.machine-man-preview+json")
	if err != nil {
		return err
	}

	var accessToken AccessToken

	if err := json.Unmarshal(data, &accessToken); err != nil {
		return errors.Wrap(err, "could not decode response")
	}
	if accessToken.Token == "" {
		return errors.New("installation access token is missing in the response")
	}

	c.token = &oauth2.Token{
		AccessToken: accessToken.Token,
		TokenType:   "token",
	}
	return nil
}

func (c *Client) InstallationRepos(ctx context.Context) ([]*Repo, error) {
	data, err := c.Request(ctx, "GET", "/installation/repositories", nil, "application/vnd.github.machine-man-preview+json")
	if err != nil {
		return nil, err
	}
	var resp struct {
		Repos []*Repo `json:"repositories"`
	}
	err = json.Unmarshal(data, &resp)
	if err != nil {
		return nil, errors.Wrap(err, "could not decode response")
	}

	return resp.Repos, nil
}

func (c *Client) ReposByInstID(ctx context.Context, instID uint64) ([]*Repo, error) {
	path := fmt.Sprintf("/user/installations/%d/repositories", instID)
	data, err := c.Request(ctx, "GET", path, nil, "application/vnd.github.machine-man-preview+json")
	if err != nil {
		return nil, err
	}
	var resp struct {
		Repos []*Repo `json:"repositories"`
	}
	err = json.Unmarshal(data, &resp)
	if err != nil {
		return nil, errors.Wrap(err, "could not decode response")
	}
	return resp.Repos, nil
}

func (c *Client) Repo(ctx context.Context, owner, name string) (*Repo, error) {
	path := fmt.Sprintf("/repos/%s/%s", owner, name)
	data, err := c.Request(ctx, "GET", path, nil, "")
	if err != nil {
		return nil, err
	}

	repo := Repo{}
	err = json.Unmarshal(data, &repo)
	if err != nil {
		return &repo, errors.Wrap(err, "could not decode response")
	}

	return &repo, nil
}

func (c *Client) CreateRepoFromTemplate(ctx context.Context, src, dst string, isPrivate bool) (*Repo, error) {
	path := fmt.Sprintf("/repos/%s/generate", src)
	jsonStr := []byte(fmt.Sprintf(`{"name":"%s","private": %v}`, dst, isPrivate))
	data, err := c.Request(
		ctx, "POST", path, jsonStr,
		"application/vnd.github.baptiste-preview+json")
	if err != nil {
		return nil, err
	}

	repo := Repo{}
	err = json.Unmarshal(data, &repo)
	if err != nil {
		return &repo, errors.Wrap(err, "could not decode response")
	}

	return &repo, nil
}

func (c *Client) CreateCheckRun(ctx context.Context, login, repo string, checkRun *CheckRun) (*CheckRun, error) {
	body, err := json.Marshal(&checkRun)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	path := fmt.Sprintf("/repos/%s/%s/check-runs", login, repo)
	data, err := c.Request(
		ctx, "POST", path, body,
		"application/vnd.github.antiope-preview+json")
	if err != nil {
		return nil, err
	}
	var cr CheckRun
	err = json.Unmarshal(data, &cr)
	if err != nil {
		return nil, errors.Wrap(err, "could not decode response")
	}
	return &cr, nil
}

func (c *Client) UpdateCheckRun(ctx context.Context, login, repo string, checkRun *CheckRun) error {
	cr := *checkRun
	id := cr.ID
	cr.ID = 0
	body, err := json.Marshal(&cr)
	if err != nil {
		return errors.WithStack(err)
	}

	path := fmt.Sprintf("/repos/%s/%s/check-runs/%d", login, repo, id)
	_, err = c.Request(ctx, "PATCH", path, body, "application/vnd.github.antiope-preview+json")
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (c *Client) Archive(ctx context.Context, login, repo, commit string) ([]byte, error) {
	path := fmt.Sprintf("/repos/%s/%s/zipball/%s", login, repo, commit)
	data, err := c.Request(ctx, "GET", path, nil, "")
	if err != nil {
		return nil, err
	}
	return data, nil
}
