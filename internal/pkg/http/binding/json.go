package binding

import (
	"net/http"

	"github.com/m1ll3r1337/catalog-service/internal/pkg/http/httph"
)

type jsonBinding struct{}

func (jsonBinding) Name() string {
	return "JSON"
}

func (jsonBinding) Bind(req *http.Request, obj any) error {
	if req == nil {
		return &bindingError{
			msg: "invalid request",
		}
	}
	if req.Body == nil {
		return &bindingError{
			msg: "invalid request",
		}
	}

	if err := httph.DecodeJSON(req, obj); err != nil {
		return &bindingError{msg: err.Error()}
	}
	if err := validate(obj); err != nil {
		return &bindingError{msg: err.Error()}
	}

	return nil
}
