package oauth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	authorizeURL = "https://github.com/login/oauth/authorize"
	tokenURL     = "https://github.com/login/oauth/access_token"
	userURL      = "https://api.github.com/user"
	emailsURL    = "https://api.github.com/user/emails"
	scopes       = "read:user user:email"
	apiVersion   = "2022-11-28"
)

var (
	ErrCodeExchangeFailed = errors.New("oauth: github code exchange failed")
	ErrNoVerifiedEmail    = errors.New("oauth: no verified email available from github")
)

type GitHubConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

type GitHubClient struct {
	cfg        GitHubConfig
	httpClient *http.Client
	tokenURL   string
	userURL    string
	emailsURL  string
}

func NewGitHubClient(cfg GitHubConfig) *GitHubClient {
	return &GitHubClient{
		cfg:        cfg,
		httpClient: &http.Client{Timeout: 10 * time.Second},
		tokenURL:   tokenURL,
		userURL:    userURL,
		emailsURL:  emailsURL,
	}
}

type GitHubUser struct {
	ID    string
	Login string
	Email string
	Name  string
}

func (c *GitHubClient) AuthorizeURL(state string) string {
	q := url.Values{
		"client_id":    {c.cfg.ClientID},
		"redirect_uri": {c.cfg.RedirectURL},
		"scope":        {scopes},
		"state":        {state},
	}
	return authorizeURL + "?" + q.Encode()
}

type githubTokenResponse struct {
	AccessToken      string `json:"access_token"`
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

func (c *GitHubClient) Exchange(ctx context.Context, code string) (string, error) {
	form := url.Values{
		"client_id":     {c.cfg.ClientID},
		"client_secret": {c.cfg.ClientSecret},
		"code":          {code},
		"redirect_uri":  {c.cfg.RedirectURL},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var tokenResp githubTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", err
	}

	if tokenResp.Error != "" || tokenResp.AccessToken == "" {
		return "", fmt.Errorf("%w: %s %s", ErrCodeExchangeFailed, tokenResp.Error, tokenResp.ErrorDescription)
	}

	return tokenResp.AccessToken, nil
}

type githubUserResponse struct {
	ID    int64   `json:"id"`
	Login string  `json:"login"`
	Email *string `json:"email"`
	Name  string  `json:"name"`
}

type githubEmailResponse struct {
	Email    string `json:"email"`
	Primary  bool   `json:"primary"`
	Verified bool   `json:"verified"`
}

func (c *GitHubClient) FetchUser(ctx context.Context, githubAccessToken string) (GitHubUser, error) {
	var profile githubUserResponse
	if err := c.getJSON(ctx, c.userURL, githubAccessToken, &profile); err != nil {
		return GitHubUser{}, err
	}

	email := c.verifiedPrimaryEmail(ctx, githubAccessToken)
	if email == "" && profile.Email != nil && *profile.Email != "" {
		email = *profile.Email
	}
	if email == "" {
		return GitHubUser{}, ErrNoVerifiedEmail
	}

	return GitHubUser{
		ID:    strconv.FormatInt(profile.ID, 10),
		Login: profile.Login,
		Email: email,
		Name:  profile.Name,
	}, nil
}

func (c *GitHubClient) verifiedPrimaryEmail(ctx context.Context, githubAccessToken string) string {
	var emails []githubEmailResponse
	if err := c.getJSON(ctx, c.emailsURL, githubAccessToken, &emails); err != nil {
		return ""
	}
	for _, e := range emails {
		if e.Primary && e.Verified {
			return e.Email
		}
	}
	return ""
}

func (c *GitHubClient) getJSON(ctx context.Context, rawURL, githubAccessToken string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+githubAccessToken)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", apiVersion)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("oauth: github request to %s failed with status %d", rawURL, resp.StatusCode)
	}

	return json.NewDecoder(resp.Body).Decode(out)
}
