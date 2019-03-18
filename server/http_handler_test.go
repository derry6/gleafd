package server

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/derry6/gleafd/pkg/log"
)

func TestGetFormValueInt(t *testing.T) {
	defv := 1
	var tests = map[string]int{
		"":      1,
		"count": 100,
		"args":  -102,
		"keys":  0,
		"abc":   1234,
	}
	for arg, want := range tests {
		url := "http://localhost:8080/snowflakes/msgs"
		if len(arg) > 0 {
			url += "?" + fmt.Sprintf("%s=%d", arg, want)
		}
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			t.Fatal(err)
		}
		v, err := getFormValueInt(req, arg, defv)
		if err != nil {
			t.Errorf("url = %v, err = %v", url, err)
		}
		if v != want {
			t.Errorf("url = %v, v = %v, want = %v", url, v, want)
		}
	}
}

type fakeSegmentService struct {
}

func (s *fakeSegmentService) GetSegments(ctx context.Context, biztag string, count int) (ids []int64, err error) {
	if count == 0 {
		count = 1
	}
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < count; i++ {
		ids = append(ids, rand.Int63())
	}
	return ids, nil
}
func (s *fakeSegmentService) GetSnowflakes(ctx context.Context, biztag string, count int) (ids []int64, err error) {
	if count == 0 {
		count = 1
	}
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < count; i++ {
		ids = append(ids, rand.Int63())
	}
	return ids, nil
}
func (s *fakeSegmentService) HealthCheck(ctx context.Context, name string) (status int, err error) {
	return 1, nil
}

func (s *fakeSegmentService) Close() error {
	return nil
}

func newFakeServer() *httptest.Server {
	r := NewHttpHandler(&fakeSegmentService{}, log.DefaultLogger)
	return httptest.NewServer(r)
}

func doTestHttpHandler(t *testing.T, uri string, rsp interface{}) {
	httpServer := newFakeServer()
	defer httpServer.Close()
	httpRsp, err := http.Get(httpServer.URL + uri)
	if err != nil {
		t.Fatal(err)
	}
	defer httpRsp.Body.Close()

	if httpRsp.StatusCode != http.StatusOK {
		t.Errorf("status code = %d, want = 200", httpRsp.StatusCode)
	}
	if ct := httpRsp.Header.Get("Content-Type"); ct != "application/json" {
		t.Errorf("content type is %v, want application/json", ct)
	}
	if err = json.NewDecoder(httpRsp.Body).Decode(rsp); err != nil {
		t.Errorf("decode response: %v", err)
	}
}

func TestSnowflakeHttpHandler(t *testing.T) {
	type GetSnowflakesResponse struct {
		Code int     `json:"code"`
		Msg  string  `json:"msg"`
		Data []int64 `json:"data"`
	}
	var getRsp GetSnowflakesResponse
	doTestHttpHandler(t, "/api/v1/snowflakes/msgs?count=10", &getRsp)

	if getRsp.Code != 0 {
		t.Errorf("code = %d, want = 0", getRsp.Code)
	}
	if len(getRsp.Data) != 10 {
		t.Errorf("len of ids is %d, want 10", len(getRsp.Data))
	}
}

func TestSegmentHttpHandler(t *testing.T) {
	type GetSnowflakesResponse struct {
		Code int     `json:"code"`
		Msg  string  `json:"msg"`
		Data []int64 `json:"data"`
	}
	var getRsp GetSnowflakesResponse
	doTestHttpHandler(t, "/api/v1/segments/msgs?count=10", &getRsp)

	if getRsp.Code != 0 {
		t.Errorf("code = %d, want = 0", getRsp.Code)
	}
	if len(getRsp.Data) != 10 {
		t.Errorf("len of ids is %d, want 10", len(getRsp.Data))
	}
}
