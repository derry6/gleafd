package server

import (
	"encoding/json"
	"net/http"
	"net/http/pprof"
	"strconv"

	"github.com/derry6/gleafd/pkg/log"
	"github.com/julienschmidt/httprouter"
)

type HttpResponse struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

func NewHttpHandler(svc Service, logger log.Logger) http.Handler {
	r := httprouter.New()

	r.Handle("GET", "/api/v1/segments/:biztag", makeGetSegmentsHandle(svc, logger))
	r.Handle("GET", "/api/v1/snowflakes/:biztag", makeGetSnowflakesHandle(svc, logger))

	r.HandlerFunc("GET", "/api/v1/health",
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			status, err := svc.HealthCheck(r.Context(), "")
			if err != nil {
				logger.Errorw("HealthCheck", "err", err)
				encodeHttpError(w, err)
				return
			}
			httpRsp := &HttpResponse{Code: 0, Msg: "Serving", Data: status}
			if err := json.NewEncoder(w).Encode(httpRsp); err != nil {
				logger.Errorw("Healthcheck", "err", err)
			}
		})

	r.HandlerFunc("GET", "/debug/pprof/", pprof.Index)
	r.HandlerFunc("GET", "/debug/pprof/cmdline", pprof.Cmdline)
	r.HandlerFunc("GET", "/debug/pprof/profile", pprof.Profile)
	r.HandlerFunc("GET", "/debug/pprof/symbol", pprof.Symbol)
	r.HandlerFunc("GET", "/debug/pprof/trace", pprof.Trace)

	// Index for /debug/pprof
	r.Handler("GET", "/debug/pprof/goroutine", pprof.Handler("goroutine"))
	r.Handler("GET", "/debug/pprof/heap", pprof.Handler("heap"))
	r.Handler("GET", "/debug/pprof/threadcreate", pprof.Handler("threadcreate"))
	r.Handler("GET", "/debug/pprof/block", pprof.Handler("block"))

	return r
}

func makeGetSegmentsHandle(svc Service, logger log.Logger) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		// Decode Request
		biztag := params.ByName("biztag")
		count, err := getFormValueInt(r, "count", 1)
		if err != nil {
			// logger
			//logger.Errorw("GetSegment", "biztag", biztag, "count", count, "err", err)
			encodeHttpError(w, err)
			return
		}
		ids, err := svc.GetSegments(r.Context(), biztag, count)
		if err != nil {
			//logger.Errorw("GetSegment", "biztag", biztag, "count", count, "err", err)
			encodeHttpError(w, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		httpRsp := &HttpResponse{Code: 0, Msg: "Ok", Data: ids}
		if err = json.NewEncoder(w).Encode(httpRsp); err != nil {
			logger.Errorw("GetSegment", "biztag", biztag, "count", count, "err", err)
		}
	}
}

func makeGetSnowflakesHandle(svc Service, logger log.Logger) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		// Decode Request
		biztag := params.ByName("biztag")
		count, err := getFormValueInt(r, "count", 1)
		if err != nil {
			// logger
			//logger.Errorw("GetSnowflake", "biztag", biztag, "count", count, "err", err)
			encodeHttpError(w, err)
			return
		}
		ids, err := svc.GetSnowflakes(r.Context(), biztag, count)
		if err != nil {
			//logger.Errorw("GetSnowflake", "biztag", biztag, "count", count, "err", err)
			encodeHttpError(w, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		httpRsp := &HttpResponse{Code: 0, Msg: "Ok", Data: ids}
		if err = json.NewEncoder(w).Encode(httpRsp); err != nil {
			logger.Errorw("GetSegment", "biztag", biztag, "count", count, "err", err)
		}
	}
}

func encodeHttpError(w http.ResponseWriter, err error) error {
	httpRsp := &HttpResponse{Code: 0, Msg: "Ok"}
	if err != nil {
		// TODO:
		httpRsp.Code = 400
		httpRsp.Msg = err.Error()
	}
	return json.NewEncoder(w).Encode(httpRsp)
}

func getFormValueInt(r *http.Request, name string, defv int) (int, error) {
	if s := r.FormValue(name); len(s) == 0 {
		return defv, nil
	} else if v, err := strconv.ParseInt(s, 10, 64); err != nil {
		return 0, err
	} else {
		return int(v), nil
	}
}
