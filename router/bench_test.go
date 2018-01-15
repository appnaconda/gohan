package router

import (
	"github.com/julienschmidt/httprouter"
	"net/http"
	"net/http/httptest"
	"testing"
)

func benchRequest(b *testing.B, router http.Handler, r *http.Request) {
	w := httptest.NewRecorder()
	u := r.URL
	rq := u.RawQuery
	r.RequestURI = u.RequestURI()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		u.RawQuery = rq
		router.ServeHTTP(w, r)

		if w.Code != http.StatusOK {
			b.Errorf("benchRequest - Expecting 200; Got %d", w.Code)
		}
	}
}

//// Gohan Router without Param
//func BenchmarkGohan_Static(b *testing.B) {
//	router := New()
//	defer func() {
//		if r := recover(); r != nil {
//			b.Errorf("should not panic")
//		}
//	}()
//
//	tt := []struct {
//		PathColon string
//		PathBrace string
//	}{
//		{"/users", "/users/{id}"},
//	}
//
//	for _, tc := range tt {
//		router.GET(tc.PathColon, func(w http.ResponseWriter, req *http.Request) {})
//	}
//
//	r, _ := http.NewRequest("GET", "/users", nil)
//	benchRequest(b, router, r)
//}
//
//// HttpRouter Router without Param
//func BenchmarkHttpRouter_Static(b *testing.B) {
//
//	router := httprouter.New()
//	defer func() {
//		if r := recover(); r != nil {
//			b.Errorf("should not panic")
//		}
//	}()
//
//	tt := []struct {
//		PathColon string
//		PathBrace string
//	}{
//		{"/users", "/users/{id}"},
//	}
//
//	for _, tc := range tt {
//		router.GET(tc.PathColon, func(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {})
//	}
//
//	r, _ := http.NewRequest("GET", "/users", nil)
//	benchRequest(b, router, r)
//}
//
//// Default HttpServeMux
//func BenchmarkHttpServeMux_Static(b *testing.B) {
//
//	router := http.NewServeMux()
//	defer func() {
//		if r := recover(); r != nil {
//			b.Errorf("should not panic")
//		}
//	}()
//
//	tt := []struct {
//		PathColon string
//		PathBrace string
//	}{
//		{"/users", "/users/{id}"},
//	}
//
//	for _, tc := range tt {
//		router.HandleFunc(tc.PathColon, func(w http.ResponseWriter, req *http.Request) {})
//	}
//
//	r, _ := http.NewRequest("GET", "/users", nil)
//	benchRequest(b, router, r)
//}

// Gohan Router with Param
func BenchmarkGohan_Param(b *testing.B) {
	router := New()
	defer func() {
		if r := recover(); r != nil {
			b.Errorf("should not panic")
		}
	}()

	tt := []struct {
		PathColon string
		PathBrace string
	}{
		{"/users/:id", "/users/{id}"},
	}

	for _, tc := range tt {
		router.GET(tc.PathColon, func(w http.ResponseWriter, req *http.Request) {})
	}

	r, _ := http.NewRequest("GET", "/users/12345", nil)
	benchRequest(b, router, r)
}

// HttpRouter Router without Param
func BenchmarkHttpRouter_Param(b *testing.B) {

	router := httprouter.New()
	defer func() {
		if r := recover(); r != nil {
			b.Errorf("should not panic")
		}
	}()

	tt := []struct {
		PathColon string
		PathBrace string
	}{
		{"/users/:id", "/users/{id}"},
	}

	for _, tc := range tt {
		router.GET(tc.PathColon, func(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {})
	}

	r, _ := http.NewRequest("GET", "/users/654321", nil)
	benchRequest(b, router, r)
}

//// Default HttpServeMux
//func BenchmarkHttpServeMux_Param(b *testing.B) {
//
//}
