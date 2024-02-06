package comfy_http_client

import (
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Client struct {
	BaseUrl string
	Header  http.Header
	client  *http.Client
}

// New 创建一个 Client, 可以通过 baseUrl 设置基本路径, 使用 header 设置默认请求头, 使用 timeout 设置超时时间
func New(baseUrl string, header http.Header, timeout time.Duration) (*Client, error) {
	baseUrl = strings.Trim(baseUrl, "/")

	client := &Client{
		BaseUrl: baseUrl,
		Header:  header,
		client: &http.Client{
			Timeout: timeout,
		},
	}

	return client, nil
}

// Get 创建 Get 请求, path 执行一下三种路径:
//   - 相对路径: 会自动拼接在 New 中传入的 baseUrl 后.
//   - 绝对路径: 会自动拼接在 New 中传入的 baseUrl 后, 并且忽略 baseUrl 本身存在的 path 部分(如果有的话).
//   - 完整的 http(s) url: 直接使用该 url, 忽略 New 中传入的 baseUrl.
func (c *Client) Get(path string) (*http.Request, error) {
	path, _ = c.joinIfRelativePath(path)

	req, err := http.NewRequest("GET", path, nil)
	c.AppendHeader(req, c.Header)

	return req, err
}

// Post 创建 Post 请求, path 同 Get 中的 path.
func (c *Client) Post(path string, body io.Reader) (*http.Request, error) {
	path, _ = c.joinIfRelativePath(path)

	req, err := http.NewRequest("POST", path, body)
	c.AppendHeader(req, c.Header)

	return req, err
}

// Put 创建 Put 请求, path 同 Get 中的 path.
func (c *Client) Put(path string, body io.Reader) (*http.Request, error) {
	path, _ = c.joinIfRelativePath(path)

	req, err := http.NewRequest("PUT", path, body)
	c.AppendHeader(req, c.Header)

	return req, err
}

// Delete 创建 Delete 请求, path 同 Get 中的 path.
func (c *Client) Delete(path string) (*http.Request, error) {
	path, _ = c.joinIfRelativePath(path)

	req, err := http.NewRequest("DELETE", path, nil)
	c.AppendHeader(req, c.Header)

	return req, err
}

// AppendQueryParams 设置 [url.URL.RawQuery] 属性
func (c *Client) AppendQueryParams(request *http.Request, queryParams url.Values) {
	request.URL.RawQuery = queryParams.Encode()
}

// AppendHeader 会将 header 添加到 [*http.Request.Header] 中. 不会修改原来已存在的 header
func (c *Client) AppendHeader(request *http.Request, header http.Header) {
	for key, values := range request.Header.Clone() {
		for _, value := range values {
			header.Add(key, value)
		}
	}

	request.Header = header
}

// Send 发送请求
func (c *Client) Send(request *http.Request) (*http.Response, error) {
	res, err := c.client.Do(request)
	return res, err
}

// ReadAsString 将响应结果读取为字符串
func (c *Client) ReadAsString(response *http.Response) (string, error) {
	builder := strings.Builder{}
	if _, err := io.Copy(&builder, response.Body); err != nil {
		return "", err
	} else {
		return builder.String(), nil
	}
}

const (
	httpPrefix  = "http://"
	httpsPrefix = "https://"
)

func (c *Client) joinIfRelativePath(path string) (string, error) {
	loweredPath := strings.ToLower(path)

	if strings.HasPrefix(loweredPath, httpPrefix) || strings.HasPrefix(loweredPath, httpsPrefix) {
		return path, nil
	}

	baseUrl, err := url.ParseRequestURI(c.BaseUrl)
	if err != nil {
		return path, err
	}

	if len(path) <= 0 {
		baseUrl.Path = ""
	} else if path[0] == 47 { // 绝对路径
		baseUrl.Path = path
	} else {
		baseUrl = baseUrl.JoinPath(path)
	}

	return baseUrl.String(), nil
}
