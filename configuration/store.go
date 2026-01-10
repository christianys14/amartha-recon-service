package configuration

import (
	"context"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"github.com/jmoiron/sqlx"
)

type storeImpl struct {
	credential Configuration
}

func NewStoreImpl(credential Configuration) *storeImpl {
	return &storeImpl{credential: credential}
}

func (d *storeImpl) initDatabase(configBaseKey string) (*sqlx.DB, error) {
	dbHost := d.credential.GetString(configBaseKey + ".host")
	dbPort := d.credential.GetString(configBaseKey + ".port")
	dbUser := d.credential.GetString(configBaseKey + ".user")
	dbPass := d.credential.GetString(configBaseKey + ".pass")
	dbName := d.credential.GetString(configBaseKey + ".name")
	sourceName := dbUser + ":" + dbPass + "@tcp(" + dbHost + ":" + dbPort + ")/" + dbName
	db, err := sqlx.Open("mysql", sourceName)

	if err != nil {
		log.Println("error when init database, because ", err)
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	if err = db.PingContext(ctx); err != nil {
		log.Println("error when connect to database", err.Error())
		return nil, err
	}

	return db, nil
}

func (d *storeImpl) InitDBMaster() (*sqlx.DB, error) {
	return d.initDatabase("database.master")
}

func (d *storeImpl) InitDbAuditTrail() (*sqlx.DB, error) {
	return d.initDatabase("database.audittrail")
}

func (d *storeImpl) InitDBReplica() (*sqlx.DB, error) {
	return d.initDatabase("database.replica")
}
