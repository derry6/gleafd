package tests

import (
	"encoding/json"
	"net/http"
	"testing"
)

func BenchmarkGetSegment(b *testing.B) {
	type response struct {
		Code int     `json:"code"`
		Msg  string  `json:"msg"`
		Data []int64 `json:"data"`
	}
	var last int64
	for i := 0; i < b.N; i++ {
		httpRsp, err := http.Get("http://127.0.0.1:9060/api/v1/segments/example?count=1")
		if err != nil {
			b.Fatal(err)
		}
		defer httpRsp.Body.Close()
		if httpRsp.StatusCode != http.StatusOK {
			b.Fatalf("statusCode = %d, want = 200", httpRsp.StatusCode)
		}
		var rsp response
		if err = json.NewDecoder(httpRsp.Body).Decode(&rsp); err != nil {
			b.Fatalf("Unknown response body: %v", err)
		}
		if rsp.Code != 0 {
			b.Fatalf("response code = %d, want = 0", rsp.Code)
		}
		if len(rsp.Data) != 1 {
			b.Fatalf("response error, data = %v", rsp.Data)
		}
		if rsp.Data[0] <= last {
			b.Fatalf("resposne error, data[0] = %d, must be greater than %d", rsp.Data[0], last)
		}
		last = rsp.Data[0]
	}
}
