// This package contains helper functions for the http.request
package request

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/gorilla/schema"
)

// Unmarshal request into an object based on the request's content-type
func Unmarshal(r *http.Request, i interface{}) error {
	contentType := r.Header.Get("Content-Type")
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}

	switch {
	case strings.HasPrefix(contentType, "application/json"):
		if err := json.Unmarshal(body, &i); err != nil {
			return err
		}

	case strings.HasPrefix(contentType, "application/xml"), strings.HasPrefix(contentType, "text/xml"):
		if err := xml.Unmarshal(body, &i); err != nil {
			return err
		}

	case strings.HasPrefix(contentType, "application/x-www-form-urlencoded"), strings.HasPrefix(contentType, "multipart/form-data"):
		decoder := schema.NewDecoder()
		err := r.ParseForm()
		if err != nil {
			return err
		}
		err = decoder.Decode(&i, r.PostForm)
		if err != nil {
			return err
		}
	default:
		return errors.New(fmt.Sprintf("Binding unsupported media type %s", contentType))
	}

	return nil
}

func PathParam(r *http.Request, name string) string {
	return r.FormValue(":" + name)
}
