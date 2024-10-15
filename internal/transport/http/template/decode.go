package template

import (
	"context"
	"encoding/json"
	"github.com/valyala/bytebufferpool"
	"google.golang.org/grpc/codes"
	"io"
	"net/http"
	"node71.otclick.ru/backend/template/pkg/error_templates"
)

func decodeCreateRequest(_ context.Context, r *http.Request) (interface{}, error) {
	request := &pbTemplate.CreateRequest{}
	buf := bytebufferpool.Get()
	defer bytebufferpool.Put(buf)

	_, err := io.Copy(buf, r.Body) // buf.ReadFrom(r.Body)
	if err != nil {
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	err = json.Unmarshal(buf.Bytes(), &request)
	if err != nil {
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	return request, nil
}
