package binding

import (
	"net/http"

	"github.com/go-playground/form/v4"
)

var formDecoder = form.NewDecoder()

type queryBinding struct{}

func (queryBinding) Name() string {
	return "URL-QUERY"
}

func (queryBinding) Bind(req *http.Request, obj any) error {
	params := req.URL.Query()

	err := formDecoder.Decode(obj, params)
	if err != nil {
		return &bindingError{msg: err.Error()}
	}

	if err := validate(obj); err != nil {
		return &bindingError{msg: err.Error()}
	}
	return nil
}
