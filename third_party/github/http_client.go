package github

import (
	"context"
	"github.com/google/go-github/v50/github"
	"github.com/google/go-querystring/query"
	"github.com/shurcooL/githubv4"
	"github.com/siaikin/home-dashboard/internal/pkg/configuration"
	"golang.org/x/oauth2"
	"net/url"
	"reflect"
)

// GitHub 的个人访问令牌.
// 来源: https://docs.github.com/authentication/keeping-your-account-and-data-secure/creating-a-personal-access-token
//var personalAccessToken string

type client struct {
	*github.Client
}

var httpClient *client

var graphqlClient *githubv4.Client

// 初始化 Github 的 http client 和 graphql client
func httpClientInitial() error {
	config := configuration.Get().ServerMonitor.ThirdParty.GitHub

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: config.PersonalAccessToken},
	)
	tokenClient := oauth2.NewClient(context.Background(), ts)

	// GitHub 的 http client
	httpClient = &client{github.NewClient(tokenClient)}
	// GitHub 的 graphql client
	graphqlClient = githubv4.NewClient(tokenClient)

	return nil
}

// addOptions adds the parameters in opts as URL query parameters to s. opts
// must be a struct whose fields may contain "url" tags.
func addOptions(s string, opts interface{}) (string, error) {
	v := reflect.ValueOf(opts)
	if v.Kind() == reflect.Ptr && v.IsNil() {
		return s, nil
	}

	u, err := url.Parse(s)
	if err != nil {
		return s, err
	}

	qs, err := query.Values(opts)
	if err != nil {
		return s, err
	}

	u.RawQuery = qs.Encode()
	return u.String(), nil
}

var lastModified string

// ListNotificationsByLastModified 通过 Last-Modified 请求头获取通知列表.
func (c *client) ListNotificationsByLastModified(ctx context.Context, opts *github.NotificationListOptions) ([]*github.Notification, *github.Response, error) {
	u := "notifications"
	u, err := addOptions(u, opts)
	if err != nil {
		return nil, nil, err
	}

	req, err := c.NewRequest("GET", u, nil)
	if err != nil {
		return nil, nil, err
	}

	// 设置 If-Modified-Since 请求头.
	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/If-Modified-Since
	// https://docs.github.com/en/rest/activity/notifications?apiVersion=2022-11-28#about-github-notifications
	if len(lastModified) > 0 {
		req.Header.Add("If-Modified-Since", lastModified)
	}

	var notifications []*github.Notification
	resp, err := c.Do(ctx, req, &notifications)
	if err != nil {
		return nil, resp, err
	}

	// 获取 Last-Modified 响应头.
	lastModified = resp.Header.Get("Last-Modified")

	return notifications, resp, nil
}

// ResetListNotificationsLastModified 重置 ListNotifications 的 Last-Modified 响应头.
func (c *client) ResetListNotificationsLastModified() {
	lastModified = ""
}
