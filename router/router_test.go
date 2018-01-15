package router

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// Test route duplication
func TestDuplication(t *testing.T) {
	router := New()
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Duplicate - Expecting: panic; Got: nil")
		}
	}()

	router.GET("/users", func(w http.ResponseWriter, req *http.Request) {})
	router.GET("/users", func(w http.ResponseWriter, req *http.Request) {})
}

// Test route duplication 2
func TestDuplication2(t *testing.T) {
	router := New()
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Duplicate 2 - Expecting: panic; Got: nil")
		}
	}()

	router.GET("/users", func(w http.ResponseWriter, req *http.Request) {})
	router.GET("/users/", func(w http.ResponseWriter, req *http.Request) {})
}

// Test conflicts
func TestConflicts(t *testing.T) {
	router := New()
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Conflicts - Expecting: panic; Got: nil")
		}
	}()

	router.GET("/users/:id/list", func(w http.ResponseWriter, req *http.Request) {})
	router.GET("/users/email/list", func(w http.ResponseWriter, req *http.Request) {})
}

// Test conflicts 2
func TestConflicts2(t *testing.T) {
	router := New()
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Conflicts - Expecting: panic; Got: nil")
		}
	}()

	router.GET("/users/email/list", func(w http.ResponseWriter, req *http.Request) {})
	router.GET("/users/:id/list", func(w http.ResponseWriter, req *http.Request) {})
}

// Test - url should begin with slash
func TestUrlShouldBeginWithSlash(t *testing.T) {
	router := New()
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Should begin with / - Expecting: panic; Got: nil")
		}
	}()

	router.GET("users", func(w http.ResponseWriter, req *http.Request) {})
}

// Test - url should not be blank
func TestUrlShouldNotBeBlank(t *testing.T) {
	router := New()
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("should be blank - Expecting: panic; Got: nil")
		}
	}()

	router.GET("", func(w http.ResponseWriter, req *http.Request) {})
}

// Test - url should not be blank
func TestHandlerShouldNotBeNil(t *testing.T) {
	router := New()
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("should be blank - Expecting: panic; Got: nil")
		}
	}()

	router.GET("/", nil)
}

// Should not panic
func TestShouldNotPanic(t *testing.T) {
	router := New()
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Should Not Panic - Expecting: nil; Got: %+v", r)
		}
	}()

	tt := []struct {
		Method string
		Url    string
	}{
		{"GET", "/"},
		{"GET", "/users"},
		{"GET", "/user/list"},
		{"GET", "/users/:id"},
		{"GET", "/users/:id/emails"},
		{"GET", "/users/:id/delete"},
		{"POST", "/users"},
		{"DELETE", "/users"},
		{"GET", "/roles"},
		{"GET", "/role/list"},
		{"GET", "/roles/:id"},
	}

	for _, tc := range tt {
		router.Handle(tc.Method, tc.Url, func(w http.ResponseWriter, req *http.Request) {})
	}
}

// Testing Handler
func TestHandler(t *testing.T) {
	router := New()
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Should Not Panic - Expecting: nil; Got: %+v", r)
		}
	}()

	tt := []struct {
		Method         string
		Pattern        string
		Handler        HandlerFunc
		Url            string
		ExtectedResult string
	}{
		{
			Method:  "GET",
			Pattern: "/",
			Handler: func(w http.ResponseWriter, req *http.Request) {
				w.Write([]byte("1"))
			},
			Url:            "/",
			ExtectedResult: "1",
		},
		{
			Method:  "GET",
			Pattern: "/users",
			Handler: func(w http.ResponseWriter, req *http.Request) {
				w.Write([]byte("hello"))
			},
			Url:            "/users",
			ExtectedResult: "hello",
		},
		{
			Method:  "POST",
			Pattern: "/users",
			Handler: func(w http.ResponseWriter, req *http.Request) {
				w.Write([]byte("hi"))
			},
			Url:            "/users",
			ExtectedResult: "hi",
		},
		{
			Method:  "GET",
			Pattern: "/users/:id",
			Handler: func(w http.ResponseWriter, req *http.Request) {
				w.Write([]byte("2"))
			},
			Url:            "/users/1",
			ExtectedResult: "2",
		},
		{
			Method:  "DELETE",
			Pattern: "/users/:id",
			Handler: func(w http.ResponseWriter, req *http.Request) {
				w.Write([]byte("deleted"))
			},
			Url:            "/users/1",
			ExtectedResult: "deleted",
		},
	}

	// Adding all handlers
	for _, tc := range tt {
		router.Handle(tc.Method, tc.Pattern, tc.Handler)
	}

	for _, tc := range tt {
		entry, _ := router.find(tc.Method, tc.Url, false)

		res := httptest.NewRecorder()

		entry.Handler(res, nil)

		if res.Body.String() != tc.ExtectedResult {
			t.Errorf("Expecting: %s; Got: %s", tc.ExtectedResult, res.Body.String())
		}
	}
}

// Testing Handler Params
func TestHandlerParams(t *testing.T) {
	router := New()
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Should Not Panic - Expecting: nil; Got: %+v", r)
		}
	}()

	router.GET("/users", func(w http.ResponseWriter, req *http.Request) {})
	router.POST("/users/:id", func(w http.ResponseWriter, req *http.Request) {})
	router.DELETE("/users/:username", func(w http.ResponseWriter, req *http.Request) {})
	router.GET("/users/:id/email", func(w http.ResponseWriter, req *http.Request) {})
	router.PUT("/users/:uuid", func(w http.ResponseWriter, req *http.Request) {})
	router.PATCH("/users/:uuid/update", func(w http.ResponseWriter, req *http.Request) {})
	router.OPTIONS("/users", func(w http.ResponseWriter, req *http.Request) {})
	router.HEAD("/users/:param1/:param2/nonparam/:param3", func(w http.ResponseWriter, req *http.Request) {})

	res := httptest.NewRecorder()

	// No params are expected
	entry, params := router.find("GET", "/users", false)
	entry.Handler(res, nil)

	if len(params) > 0 {
		t.Errorf("Expecting no param; got: %+v", params)
	}

	// Params
	entry, params = router.find("POST", "/users/1", false)
	entry.Handler(res, nil)

	if len(params) <= 0 {
		t.Errorf("Expecting 1 param; got: %d [%+v]", len(params), params)
	}

	value, ok := params[":id"]
	if !ok {
		t.Errorf("Expecting param ':id'; got: nil")
	}

	if value != "1" {
		t.Errorf("Param ID - Expecting %s; got: %s", "1", value)
	}

	// Params
	entry, params = router.find("DELETE", "/users/123456", false)
	entry.Handler(res, nil)

	if len(params) <= 0 {
		t.Errorf("Expecting 1 param; got: %d [%+v]", len(params), params)
	}

	value, ok = params[":username"]
	if !ok {
		t.Errorf("Expecting param ':username'; got: nil")
	}

	if value != "123456" {
		t.Errorf("Param USERNAME - Expecting %s; got: %s", "1", value)
	}

	// Params
	entry, params = router.find("GET", "/users/123456789/email", false)
	entry.Handler(res, nil)

	if len(params) <= 0 {
		t.Errorf("Expecting 1 param; got: %d [%+v]", len(params), params)
	}

	value, ok = params[":id"]
	if !ok {
		t.Errorf("Expecting param ':id'; got: nil")
	}

	if value != "123456789" {
		t.Errorf("Param ID - Expecting %s; got: %s", "1", value)
	}

	// Params
	entry, params = router.find("PATCH", "/users/123e4567-e89b-12d3-a456-426655440000/update", false)
	entry.Handler(res, nil)

	if len(params) <= 0 {
		t.Errorf("Expecting 1 param; got: %d [%+v]", len(params), params)
	}

	value, ok = params[":uuid"]
	if !ok {
		t.Errorf("Expecting param ':uuid'; got: nil")
	}

	if value != "123e4567-e89b-12d3-a456-426655440000" {
		t.Errorf("Param UUID - Expecting %s; got: %s", "1", value)
	}

	// No params are expected
	entry, params = router.find("OPTIONS", "/users", false)
	entry.Handler(res, nil)

	if len(params) > 0 {
		t.Errorf("Expecting no param; got: %+v", params)
	}

	// Params
	entry, params = router.find("HEAD", "/users/hello/hi/nonparam/hola", false)
	entry.Handler(res, nil)

	if len(params) < 3 {
		t.Errorf("Expecting 3 param; got: %d [%+v]", len(params), params)
	}

	value, ok = params[":param1"]
	if !ok {
		t.Errorf("Expecting param ':param1'; got: nil")
	}

	if value != "hello" {
		t.Errorf("Param PARAM1 - Expecting %s; got: %s", "hello", value)
	}

	value, ok = params[":param2"]
	if !ok {
		t.Errorf("Expecting param ':param2'; got: nil")
	}

	if value != "hi" {
		t.Errorf("Param PARAM2 - Expecting %s; got: %s", "hi", value)
	}

	value, ok = params[":param3"]
	if !ok {
		t.Errorf("Expecting param ':param3'; got: nil")
	}

	if value != "hola" {
		t.Errorf("Param PARAM3 - Expecting %s; got: %s", "hola", value)
	}
}

// Test ServeHttp
func handlerTest(w http.ResponseWriter, req *http.Request) {
	w.Write([]byte("OK"))
}
func handlerTest2(w http.ResponseWriter, req *http.Request) {
	w.Write([]byte(req.Form.Get(":uuid")))
}
func TestHServeHttp(t *testing.T) {
	router := New()
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Should Not Panic - Expecting: nil; Got: %+v", r)
		}
	}()

	router.GET("/users", handlerTest)
	router.POST("/users/:id", handlerTest)
	router.DELETE("/users/:username", handlerTest)
	router.GET("/users/:id/email", handlerTest)
	router.PUT("/users/:uuid", handlerTest2)
	router.PATCH("/users/:uuid/update", handlerTest)
	router.OPTIONS("/users", handlerTest)
	router.HEAD("/users/:param1/:param2/nonparam/:param3", handlerTest)

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("POST", "/users/123456", nil)
	router.ServeHTTP(w, r)

	if w.Body.String() != "OK" {
		t.Errorf("ServeHTTP - Expecting ok; Got %s", w.Body.String())
	}

	// Not found
	w = httptest.NewRecorder()
	r, _ = http.NewRequest("GET", "/users/123456", nil)
	router.ServeHTTP(w, r)

	if w.Code != http.StatusNotFound {
		t.Errorf("ServeHTTP - Expecting 404; Got %d", w.Code)
	}

	// Testing param
	w = httptest.NewRecorder()
	r, _ = http.NewRequest("PUT", "/users/12345-67890", nil)
	router.ServeHTTP(w, r)

	if w.Body.String() != "12345-67890" {
		t.Errorf("ServeHTTP - Expecting 12345-67890; Got %s", w.Body.String())
	}

}
