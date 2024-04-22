package engine

import (
	"net/http"
	"runtime/debug"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func newRouter(version string, enableTrailingSlash bool, enableCORS bool) *chi.Mux {
	router := chi.NewRouter()
	router.Use(middleware.RealIP)  // should be first middleware
	router.Use(middleware.Logger)  // log to console
	router.Use(problemRecoverer)   // catch panics and turn into 500s
	router.Use(middleware.GetHead) // support HEAD requests https://docs.ogc.org/is/17-069r4/17-069r4.html#_http_1_1
	if enableTrailingSlash {
		router.Use(middleware.StripSlashes)
	}
	if enableCORS {
		router.Use(cors.Handler(cors.Options{
			AllowedOrigins:   []string{"*"},
			AllowedMethods:   []string{http.MethodGet, http.MethodHead, http.MethodOptions},
			AllowedHeaders:   []string{HeaderRequestedWith},
			ExposedHeaders:   []string{HeaderContentCrs, HeaderLink},
			AllowCredentials: false,
			MaxAge:           int((time.Hour * 24).Seconds()),
		}))
	}
	// some GIS clients don't sent proper CORS preflight requests, still respond with OK for any OPTIONS request
	router.Use(optionsFallback)
	// add semver header, implements https://gitdocumentatie.logius.nl/publicatie/api/adr/#api-57
	router.Use(middleware.SetHeader(HeaderAPIVersion, version))
	router.Use(middleware.Compress(5, CompressibleMediaTypes...)) // enable gzip responses
	return router
}

func optionsFallback(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// Custom middleware.Recoverer adapted from Chi (https://github.com/go-chi/chi/blob/master/middleware/recoverer.go)
// to return RFC-7807 Problem messages.
func problemRecoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rvr := recover(); rvr != nil {
				if rvr == http.ErrAbortHandler { //nolint:errorlint // already so in Chi
					// we don't recover http.ErrAbortHandler so the response
					// to the client is aborted, this should not be logged
					panic(rvr)
				}

				logEntry := middleware.GetLogEntry(r)
				if logEntry != nil {
					logEntry.Panic(rvr, debug.Stack())
				} else {
					middleware.PrintPrettyStack(rvr)
				}

				if r.Header.Get("Connection") != "Upgrade" {
					RenderProblem(ProblemServerError, w)
				}
			}
		}()
		next.ServeHTTP(w, r)
	})
}
