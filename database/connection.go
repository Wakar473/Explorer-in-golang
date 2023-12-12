package database

import (
	"context"
	"database/sql"
	"log"
	"os"
	// "time"

	"boilerplate/constant"

	_ "github.com/go-sql-driver/mysql"
)
var db *sql.DB

// type manager struct {
// 	connection *sql.DB
// 	ctx        context.Context
// 	cancel     context.CancelFunc
// }

var Mgr Manager

type Manager interface {
	Insert(interface{}, string) error
}

func ConnectDb() {
	
	var err error
	uri := os.Getenv("DB_HOST")
	if uri == "" {
		uri = constant.MYSQLDBUri
	}
	
	db, err = sql.Open(constant.Dialect, ""+constant.UserName+":@tcp("+constant.MYSQLDBUri+")/"+constant.Database+
	"")
	// log.Printf("========== db ===================", db)
	if err != nil {
		log.Fatalf("Failed to connect to the MySQL database: %v", err)
	}
	// if err != nil {
	// 	ConnectDb()
	// }
	log.Printf("Successfully connected to the database at %s", uri)

	
}
	

	// if utils.IsDevelopment() {
	// 	logger.SugaredLogger.Infof("Successfully connected to the database at %s", uri)
	// } else {
	// 	logger.SugaredLogger.Info("Successfully connected to the database")
	// }


  func Close(client *sql.DB, ctx context.Context,
	cancel context.CancelFunc) {

	// CancelFunc to cancel to context
	defer cancel()

	// client provides a method to close
	// a mongoDB connection.
	defer func() {

		// client.Disconnect method also has deadline.
		// returns error if any,
		
	}()
}
