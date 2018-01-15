package url

import (
	"net/url"
)

func Join(base string, relative string) (string, error) {
	u, err := url.Parse(relative)
	if err != nil {
		return "", err
	}
	b, err := url.Parse(base)
	if err != nil {
		return "", err
	}
	return b.ResolveReference(u).String(), nil
}
