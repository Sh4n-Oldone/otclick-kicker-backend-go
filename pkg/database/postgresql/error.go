package postgresql

import (
	"database/sql"
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"

	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/error_templates"
)

// DecodeDatabaseError is a function that decodes some reason database errors into human-understandable error messages.
func DecodeDatabaseError(err error) error {
	if errors.Unwrap(err) != nil {
		err = errors.Unwrap(err)
	}
	pgConn, ok := err.(*pgconn.PgError)
	if ok {
		switch pgConn.Code {
		default:
			// Unknown database error
			return error_templates.New(pgConn.Message, errors.New(pgConn.Message), codes.Internal, http.StatusInternalServerError)
		}
	}
	switch err {
	case pgx.ErrNoRows, sql.ErrNoRows:
		// No rows in such desired arguments
		return error_templates.New(err.Error(), err, codes.NotFound, http.StatusNotFound)
	default:
		return error_templates.New(err.Error(), err, codes.Internal, http.StatusInternalServerError)
	}
}
