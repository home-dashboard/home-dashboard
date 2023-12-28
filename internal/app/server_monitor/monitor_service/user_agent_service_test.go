package monitor_service

import "testing"

func TestFetchUserAgent(t *testing.T) {
	ua, err := FetchUserAgent()
	if err != nil {
		t.Error(err)
	}

	t.Log("ua: ", len(*ua))
}
