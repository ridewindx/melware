package melware

import (
	"database/sql"
	"github.com/spf13/viper"
	"github.com/ridewindx/mel"
)

func DBConfig(mel *mel.Mel) *viper.Viper {
	config := GetConfig(mel)
	dbConfig := config.Sub("db")

	dbConfig.SetDefault("conn_max_lifetime", 0)
	dbConfig.SetDefault("mac_idle_conns", 0)
	dbConfig.SetDefault("max_open_conns", 0)

	return dbConfig
}

func DB(mel *mel.Mel) *sql.DB {
	dbConfig := DBConfig(mel)
	db, err := sql.Open(dbConfig.GetString("driver"), dbConfig.GetString("data_source"))
	if err != nil {
		panic(err)
	}
	db.SetConnMaxLifetime(dbConfig.GetInt64("conn_max_lifetime"))
	return db
}
