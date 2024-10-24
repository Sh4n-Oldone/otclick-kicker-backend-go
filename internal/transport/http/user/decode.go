package user

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/valyala/bytebufferpool"
	"google.golang.org/grpc/codes"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/entity"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/error_templates"
)

func decodeLoginRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	request := &entity.LoginUserRequest{}

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

func decodeCreateRequest(_ context.Context, r *http.Request) (interface{}, error) {
	request := &entity.CreateUserRequest{}

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

func decodeChangePasswordRequest(_ context.Context, r *http.Request) (interface{}, error) {
	request := &entity.ChangePasswordRequest{}

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

func decodeCheckAuthRequest(_ context.Context, r *http.Request) (interface{}, error) {
	return nil, nil
}