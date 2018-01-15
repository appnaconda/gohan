package database

import (
	"context"
	"database/sql"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/GoogleCloudPlatform/cloudsql-proxy/proxy/proxy"
	"github.com/go-sql-driver/mysql"
	goauth "golang.org/x/oauth2/google"
	//"github.com/ExpansiveWorlds/instrumentedsql"
	//"github.com/ExpansiveWorlds/instrumentedsql/google"
)

const SQLScope = "https://www.googleapis.com/auth/sqlservice.admin"

func init() {
	mysql.RegisterDial("cloudsql", proxy.Dial)
}

// New creates a new database connection using a connection string (DB_CONN_STR).
// If the connection string contains the string '@cloudsql(' (e.g. username:password@cloudsql(test-instance-name)/dbname)
// a cloud sql proxy will be initialized using the account service token file (DB_TOKEN_FILE).
func New(ctx context.Context) (*sql.DB, error) {

	conn := os.Getenv("DB_CONN_STR")

	if conn == "" {
		return nil, fmt.Errorf("failed creating the db connection. The DB_CONN_STR env variable was not found")
	}

	if strings.Contains(conn, "@cloudsql(") {
		tokenFile := os.Getenv("DB_TOKEN_FILE")
		if tokenFile == "" {
			return nil, fmt.Errorf("failed creating the db connection. trying to connect using cloudsql proxy and the token file (DB_TOKEN_FILE)  was not provided.")
		}

		if err := initProxy(ctx, tokenFile); err != nil {
			return nil, err
		}
	}

	//sql.Register("instrumented-mysql", instrumentedsql.WrapDriver(mysql.MySQLDriver{}, instrumentedsql.WithTracer(google.NewTracer())))
	db, err := sql.Open("mysql", conn)
	if err != nil {
		return nil, err
	}

	if err := db.PingContext(ctx); err != nil {
		return nil, err
	}

	dbMaxOpenConns, err := strconv.Atoi(os.Getenv("DB_MAX_OPEN_CONNS"))
	if err != nil || dbMaxOpenConns == 0 {
		// We should always use connection limits, if not limit was specified,
		// 3 will be the default value
		dbMaxOpenConns = 3
	}

	dbMaxIdleConns, err := strconv.Atoi(os.Getenv("DB_MAX_IDLE_CONNS"))
	if err != nil {
		dbMaxIdleConns = 3
	}

	db.SetMaxOpenConns(dbMaxOpenConns)
	db.SetMaxIdleConns(dbMaxIdleConns)

	return db, nil
}

func initProxy(ctx context.Context, tokenFile string) error {
	// initialize proxy
	all, err := ioutil.ReadFile(tokenFile)
	if err != nil {
		return fmt.Errorf("invalid json file %q: %v", tokenFile, err)
	}

	cfg, err := goauth.JWTConfigFromJSON(all, SQLScope)
	if err != nil {
		return fmt.Errorf("invalid json file %q: %v", tokenFile, err)
	}

	client := cfg.Client(ctx)

	if client != nil {
		// Initializing the proxy
		proxy.Init(client, nil, nil)
	}

	return nil
}
