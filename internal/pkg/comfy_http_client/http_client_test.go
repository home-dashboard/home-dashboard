package comfy_http_client

import (
	"errors"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"
)

func createClient(t *testing.T) *Client {
	t.Helper()

	client, err := New("http://jsonplaceholder.typicode.com", http.Header{
		"Authorization": []string{"Barer ..."},
	}, time.Second*30)

	if err != nil {
		t.Error(err)
	}

	return client
}

func TestNew(t *testing.T) {
	client := createClient(t)

	t.Run("absolute path", func(t *testing.T) {
		if req, err := client.Get("/posts"); err != nil {
			t.Error(err)
		} else if req.URL.EscapedPath() != "/posts" {
			t.Error("path should equal \"/posts\"")
		}
	})

	t.Run("relative path", func(t *testing.T) {
		if req, err := client.Get("posts"); err != nil {
			t.Error(err)
		} else if req.URL.EscapedPath() != "/posts" {
			t.Error("path should equal \"/posts\"")
		}
	})

	t.Run("http(s) url", func(t *testing.T) {
		if req, err := client.Get("http://jsonplaceholder.typicode.com/posts"); err != nil {
			t.Error(err)
		} else if req.URL.EscapedPath() != "/posts" {
			t.Error("path should equal \"/posts\"")
		}
	})
}

func TestClient_Get(t *testing.T) {
	client := createClient(t)

	t.Run("no params", func(t *testing.T) {
		if req, err := client.Get("/posts"); err != nil {
			t.Error(err)
		} else if res, err := client.Send(req); err != nil {
			t.Error(err)
		} else if _, err := client.ReadAsString(res); err != nil {
			t.Error(err)
		}
	})

	t.Run("with params", func(t *testing.T) {
		if req, err := client.Get("/posts"); err != nil {
			t.Error(err)
		} else {
			client.AppendQueryParams(req, url.Values{"userId": []string{"1"}})
			if !req.URL.Query().Has("userId") {
				t.Error(errors.New("query param userId is empty"))
			}

			if res, err := client.Send(req); err != nil {
				t.Error(err)
			} else if _, err := client.ReadAsString(res); err != nil {
				t.Error(err)
			}
		}
	})
}

func TestClient_Post(t *testing.T) {
	client := createClient(t)

	t.Run("no body", func(t *testing.T) {
		if req, err := client.Post("/posts", strings.NewReader("")); err != nil {
			t.Error(err)
		} else if res, err := client.Send(req); err != nil {
			t.Error(err)
		} else if _, err := client.ReadAsString(res); err != nil {
			t.Error(err)
		}
	})

	t.Run("with body", func(t *testing.T) {
		if req, err := client.Post("/posts", strings.NewReader("{\"title\":\"foo\",\"body\":\"bar\",\"userId\":1}")); err != nil {
			t.Error(err)
		} else if res, err := client.Send(req); err != nil {
			t.Error(err)
		} else if _, err := client.ReadAsString(res); err != nil {
			t.Error(err)
		}
	})
}

func TestClient_Put(t *testing.T) {
	client := createClient(t)

	if req, err := client.Put("/posts/1", strings.NewReader("")); err != nil {
		t.Error(err)
	} else if res, err := client.Send(req); err != nil {
		t.Error(err)
	} else if _, err := client.ReadAsString(res); err != nil {
		t.Error(err)
	}
}

func TestClient_Delete(t *testing.T) {
	client := createClient(t)

	if req, err := client.Delete("/posts/1"); err != nil {
		t.Error(err)
	} else if res, err := client.Send(req); err != nil {
		t.Error(err)
	} else if _, err := client.ReadAsString(res); err != nil {
		t.Error(err)
	}
}
