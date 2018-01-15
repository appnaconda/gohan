package response

import (
	"encoding/json"
	"encoding/xml"
	"net/http"
)

func Ok(w http.ResponseWriter) error {
	return Blob(w, []byte(""), http.StatusOK)
}

func NoContent(w http.ResponseWriter, code int) error {
	return Blob(w, []byte(""), code)
}

func String(w http.ResponseWriter, s string, code int) error {
	return StringBlob(w, []byte(s), code)
}
func StringBlob(w http.ResponseWriter, b []byte, code int) error {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	return Blob(w, b, code)
}

func HTML(w http.ResponseWriter, s string, code int) error {
	return HTMLBlob(w, []byte(s), code)
}

func HTMLBlob(w http.ResponseWriter, b []byte, code int) error {
	w.Header().Add("Content-Type", "text/html; charset=utf-8")
	return Blob(w, b, code)
}

func JSON(w http.ResponseWriter, data interface{}, code int) error {
	b, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return JSONBlob(w, b, code)
}

func JSONPretty(w http.ResponseWriter, data interface{}, indent string, code int) error {
	b, err := json.MarshalIndent(data, "", indent)
	if err != nil {
		return err
	}
	return JSONBlob(w, b, code)
}

func JSONBlob(w http.ResponseWriter, b []byte, code int) error {
	w.Header().Add("Content-Type", "application/json")
	return Blob(w, b, code)
}

func XML(w http.ResponseWriter, data interface{}, code int) error {
	b, err := xml.Marshal(data)
	if err != nil {
		return err
	}
	return XMLBlob(w, b, code)
}

func XMLPretty(w http.ResponseWriter, data interface{}, indent string, code int) error {
	b, err := xml.MarshalIndent(data, "", indent)
	if err != nil {
		return err
	}
	return XMLBlob(w, b, code)
}

func XMLBlob(w http.ResponseWriter, b []byte, code int) error {
	w.Header().Add("Content-Type", "application/xml; charset=UTF-8")
	return Blob(w, b, code)
}

func Blob(w http.ResponseWriter, body []byte, code int) error {
	w.WriteHeader(code)
	_, err := w.Write(body)
	return err
}
