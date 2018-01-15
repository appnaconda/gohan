package url

import (
	"testing"
)

func TestJoinUrl(t *testing.T) {
	tt := []struct {
		baseUrl     string
		relativeUrl string
		expect      string
	}{
		{baseUrl: "http://example.com", relativeUrl: "helloworld", expect: "http://example.com/helloworld"},
		{baseUrl: "http://example.com", relativeUrl: "/helloworld/", expect: "http://example.com/helloworld/"},
		{baseUrl: "http://example.com/", relativeUrl: "/helloworld/", expect: "http://example.com/helloworld/"},
		{baseUrl: "http://example.com//", relativeUrl: "helloworld/there", expect: "http://example.com/helloworld/there"},
		{baseUrl: "http://example.com//", relativeUrl: "user/get?email=test@test.com", expect: "http://example.com/user/get?email=test@test.com"},
	}

	for _, tc := range tt {
		result, err := Join(tc.baseUrl, tc.relativeUrl)
		if err != nil {
			t.Errorf("expecting not error; got %s", err)
		}

		if result != tc.expect {
			t.Errorf("expecting %s; got %s", tc.expect, result)
		}
	}
}
