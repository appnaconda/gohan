package gohan

import (
	"database/sql"
	"fmt"

	"context"

	"net/http"

	"github.com/appnaconda/gohan/logger"
)

type ServiceContext struct {
	Logger               logger.Logger
	Context              context.Context
	HttpClient           *http.Client
	db                   *sql.DB
	LoggedUserIdentifier string
}

func (sc *ServiceContext) GetDB() (*sql.DB, error) {
	if sc.db == nil {
		return nil, fmt.Errorf("no databse connection was found")
	}

	return sc.db, nil
}
