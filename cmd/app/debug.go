//go:build debug
// +build debug

package main

import (
	"expvar"
	"fmt"
	"net/http"
	"net/http/pprof"
	_ "net/http/pprof"
	"runtime"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/pkg/profile"
)

const webDebugEnabled = true

type stopper interface {
	Stop()
}

func initDebugger() stopper {
	return profile.Start(
		profile.NoShutdownHook,
		profile.Quiet,
		profile.CPUProfile,
		profile.MemProfile,
		profile.MemProfileAllocs,
		profile.MemProfileHeap,
		profile.MemProfileRate(runtime.MemProfileRate),
		profile.MutexProfile,
		profile.BlockProfile,
		profile.TraceProfile,
		profile.ThreadcreationProfile,
		profile.GoroutineProfile,
		profile.ProfilePath("./pprof"),
	)
}

func stopDebugger(profiler stopper) {
	profiler.Stop()
}

func ProfilerHandler() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.NoCache)
	r.Use(middleware.RealIP)
	r.Use(custom_middleware.RequestID)
	r.Use(middleware.Recoverer)
	r.Use(middleware.StripSlashes)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, r.RequestURI+"/pprof/", http.StatusMovedPermanently)
	})
	r.HandleFunc("/pprof", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, r.RequestURI+"/", http.StatusMovedPermanently)
	})

	r.HandleFunc("/pprof/*", pprof.Index)
	r.HandleFunc("/pprof/index", pprof.Index)
	r.HandleFunc("/pprof/cmdline", pprof.Cmdline)
	r.HandleFunc("/pprof/profile", pprof.Profile)
	r.HandleFunc("/pprof/symbol", pprof.Symbol)
	r.HandleFunc("/pprof/trace", pprof.Trace)

	r.HandleFunc("/pprof/allocs", pprof.Handler("allocs").ServeHTTP)
	r.HandleFunc("/pprof/block", pprof.Handler("block").ServeHTTP)
	r.HandleFunc("/pprof/goroutine", pprof.Handler("goroutine").ServeHTTP)
	r.HandleFunc("/pprof/heap", pprof.Handler("heap").ServeHTTP)
	r.HandleFunc("/pprof/mutex", pprof.Handler("mutex").ServeHTTP)
	r.HandleFunc("/pprof/threadcreate", pprof.Handler("threadcreate").ServeHTTP)

	r.Mount("/charts", debugcharts.PublicMux)

	r.HandleFunc("/vars", expVars)

	return r
}

// Replicated from expvar.go as not public.
func expVars(w http.ResponseWriter, r *http.Request) {
	first := true
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "{\n")
	expvar.Do(func(kv expvar.KeyValue) {
		if !first {
			fmt.Fprintf(w, ",\n")
		}
		first = false
		fmt.Fprintf(w, "%q: %s", kv.Key, kv.Value)
	})
	fmt.Fprintf(w, "\n}\n")
}
