package configuration

import (
	"context"
	"database/sql"
	"log"
	"time"
)

type storeImpl struct {
	credential Configuration
}

func NewStoreImpl(credential Configuration) *storeImpl {
	return &storeImpl{credential: credential}
}

func (d *storeImpl) initDatabase(configBaseKey string) (*sql.DB, error) {
	dbHost := d.credential.GetString(configBaseKey + ".host")
	dbPort := d.credential.GetString(configBaseKey + ".port")
	dbUser := d.credential.GetString(configBaseKey + ".user")
	dbPass := d.credential.GetString(configBaseKey + ".pass")
	dbName := d.credential.GetString(configBaseKey + ".name")
	sourceName := dbUser + ":" + dbPass + "@tcp(" + dbHost + ":" + dbPort + ")/" + dbName
	db, err := sql.Open("mysql", sourceName)

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

func (d *storeImpl) InitDBMaster() (*sql.DB, error) {
	return d.initDatabase("database.master")
}

func (d *storeImpl) InitDbAuditTrail() (*sql.DB, error) {
	return d.initDatabase("database.audittrail")
}

func (d *storeImpl) InitDBReplica() (*sql.DB, error) {
	return d.initDatabase("database.replica")
}
