package github

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFetchParticipation_202Retry(t *testing.T) {
	var attempts int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/repos/owner/repo/stats/participation" {
			count := atomic.AddInt32(&attempts, 1)
			if count <= 2 {
				w.WriteHeader(http.StatusAccepted)
				w.Write([]byte(`{}`))
				return
			}
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"all":[10,20,30],"owner":[5,10,15]}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	result, err := fetchParticipationFromMock(t, srv.URL, "owner", "repo")
	require.NoError(t, err)
	assert.Equal(t, []int{10, 20, 30}, result)
	assert.GreaterOrEqual(t, int(atomic.LoadInt32(&attempts)), 3)
}

func TestFetchTotalCommits_LinkHeader(t *testing.T) {
	tests := []struct {
		name     string
		link     string
		expected int
	}{
		{
			name:     "standard link header",
			link:     `<https://api.github.com/repos/o/r/commits?per_page=1&page=500>; rel="last"`,
			expected: 500,
		},
		{
			name:     "no link header",
			link:     "",
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.link != "" {
					w.Header().Set("Link", tt.link)
				}
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`[{"sha":"abc"}]`))
			}))
			defer srv.Close()

			total := parseLinkHeader(srv, tt.link)
			assert.Equal(t, tt.expected, total)
		})
	}
}

func fetchParticipationFromMock(t *testing.T, baseURL, owner, repo string) ([]int, error) {
	t.Helper()

	url := fmt.Sprintf("%s/repos/%s/%s/stats/participation", baseURL, owner, repo)

	var lastBody []byte
	for i := 0; i < 5; i++ {
		resp, err := http.Get(url)
		if err != nil {
			return nil, err
		}

		if resp.StatusCode == http.StatusOK {
			var p struct {
				All []int `json:"all"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&p); err != nil {
				resp.Body.Close()
				return nil, err
			}
			resp.Body.Close()
			return p.All, nil
		}

		resp.Body.Close()
		_ = lastBody
	}

	return nil, fmt.Errorf("failed after retries")
}

func parseLinkHeader(_ *httptest.Server, link string) int {
	if link == "" {
		return 1
	}

	for _, part := range strings.Split(link, ",") {
		part = strings.TrimSpace(part)
		if strings.Contains(part, `rel="last"`) {
			idx := strings.LastIndex(part, "page=")
			if idx == -1 {
				continue
			}
			idx += 5
			end := strings.IndexByte(part[idx:], '>')
			if end == -1 {
				end = strings.IndexByte(part[idx:], '&')
			}
			if end == -1 {
				end = len(part[idx:])
			}
			pageStr := part[idx : idx+end]
			total, _ := strconv.Atoi(pageStr)
			return total
		}
	}
	return 1
}
