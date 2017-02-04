package melware

import (
	"database/sql"
    "time"
	"github.com/spf13/viper"
	"github.com/ridewindx/mel"
)

func DBConfig(mel *mel.Mel) *viper.Viper {
	config := GetConfig(mel)
	dbConfig := config.Sub("db")

	dbConfig.SetDefault("conn_max_lifetime", 0)
	dbConfig.SetDefault("max_idle_conns", 0)
	dbConfig.SetDefault("max_open_conns", 0)

	return dbConfig
}

func DB(mel *mel.Mel) *sql.DB {
	dbConfig := DBConfig(mel)
	db, err := sql.Open(dbConfig.GetString("driver"), dbConfig.GetString("data_source"))
	if err != nil {
		panic(err)
	}
	db.SetConnMaxLifetime(time.Duration(dbConfig.GetInt64("conn_max_lifetime")))
	db.SetMaxIdleConns(dbConfig.GetInt("max_idle_conns"))
	db.SetMaxOpenConns(dbConfig.GetInt("max_open_conns"))

    // Send a ping to make sure the database connection is alive.
	err = db.Ping()
	if err != nil {
		db.Close()
		panic(err)
	}

	return db
}
