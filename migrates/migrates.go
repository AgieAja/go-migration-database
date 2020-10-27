package migrates

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/rs/zerolog/log"
	"io/ioutil"
	"strconv"
	"time"
)

type MigrationConfig struct {
	//db is the database instance
	Db *sql.DB

	//Dialect is the type of Database (mysql)
	Dialect string

	//MigrationDir is the location of migration folder (default to db/migrations)
	MigrationDir string
}

func NewMigrationConfig(migrationDir, dbHost, dbPort, dbUser, dbPass, dbName, dbDriver string) (*MigrationConfig, error) {
	//dbHost := os.Getenv("DB_HOST")
	//dbPort := os.Getenv("DB_PORT")
	//dbUser := os.Getenv("DB_USER")
	//dbPass := os.Getenv("DB_PASS")
	//dbName := os.Getenv("DB_NAME")
	log.Info().Msg("db Host : " + dbHost)
	log.Info().Msg("db Port : " + dbPort)
	log.Info().Msg("db User : " + dbUser)
	log.Info().Msg("db password : " + dbPass)
	log.Info().Msg("db name : " + dbName)
	log.Info().Msg("db driver : " + dbDriver)

	migrationConf := MigrationConfig{MigrationDir: migrationDir}
	dbDSN := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&multiStatements=true", dbUser, dbPass, dbHost, dbPort, dbName)
	//dbDriver := os.Getenv("DB_DRIVER")

	var err error
	//open db connection based on driver
	switch dbDriver {
	case "mysql":
		migrationConf.Dialect = dbDriver
		migrationConf.Db, err = sql.Open(dbDriver, dbDSN)
		if err != nil {
			return nil, errors.New("Migrate.NewMigrationConfig err : " + err.Error())
		}
	default:
		return nil, errors.New("error db driver is not found (currently mysql supported only)")
	}

	return &migrationConf, nil
}

//migrate up will migrate the database to the latest version
func MigrateUp(config *MigrationConfig) error {
	log.Info().Msg("Migrating up database ...")
	driver, errDriver := mysql.WithInstance(config.Db, &mysql.Config{})
	if errDriver != nil {
		return errors.New("Migrate.MigrateUp errDriver : " + errDriver.Error())
	}

	migrateDatabase, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", config.MigrationDir),
		config.Dialect, driver)
	if err != nil {
		return errors.New("Migrate.MigrateUp err : " + err.Error())
	}

	errUp := migrateDatabase.Up()
	if errUp != nil {
		return errors.New("Migrate.MigrateUp errUp : " + errUp.Error())
	}

	log.Info().Msg("Migration done ...")

	//get latest version
	version, dirty, errVersion := migrateDatabase.Version()
	//ignore error in this line. Skip the version check
	if errVersion != nil {
		return errors.New("Migrate.MigrateUp errVersion : " + errVersion.Error())
	}

	if dirty {
		log.Info().Msg("dirty migration. Please clean up database")
	}

	msgLatestVersion := fmt.Sprintf("latest version is %d", version)
	log.Info().Msg(msgLatestVersion)
	return nil
}

//CreateNewMigrationFile will create new migration file to the specified page
func CreateNewMigrationFile(filePath, fileName string) error {

	//get current time
	currentTime := time.Now()

	currentYear := currentTime.Year()
	currentMonth := fmt.Sprintf("%02d", int(currentTime.Month()))
	currentDay := fmt.Sprintf("%02d", currentTime.Day())

	currentHour := fmt.Sprintf("%02d", currentTime.Hour())
	currentMinute := fmt.Sprintf("%02d", currentTime.Minute())
	currentSec := fmt.Sprintf("%02d", currentTime.Second())

	currentTimeFilePrefix := strconv.Itoa(currentYear) + currentMonth + currentDay +
		currentHour + currentMinute + currentSec

	fmt.Println(currentHour + currentMinute + currentSec)
	migrationUpFullPath := filePath + "/" +
		currentTimeFilePrefix + "_" + fileName + ".up.sql"
	err := ioutil.WriteFile(migrationUpFullPath, []byte(""), 0644)
	if err != nil {
		return errors.New("Migrate.CreateNewMigrationFile err : " + err.Error())
	}

	migrationDownFullPath := filePath + "/" +
		currentTimeFilePrefix + "_" + fileName + ".down.sql"
	err = ioutil.WriteFile(migrationDownFullPath, []byte(""), 0644)
	if err != nil {
		return errors.New("Migrate.CreateNewMigrationFile err : " + err.Error())
	}

	return nil
}

//migrate down will migrate the database to -1 of the version
func MigrateDown(config *MigrationConfig) error {
	log.Info().Msg("Migrating down database ...")
	driver, errDriver := mysql.WithInstance(config.Db, &mysql.Config{})
	if errDriver != nil {
		return errDriver
	}

	migrationDatabase, errMigrate := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", config.MigrationDir),
		config.Dialect, driver)
	if errMigrate != nil {
		return errors.New("Migrate.MigrateDown errMigrate : " + errMigrate.Error())
	}

	err := migrationDatabase.Down()
	if err != nil {
		return errors.New("Migrate.MigrateDown err : " + err.Error())
	}

	log.Info().Msg("Migration done ...")

	//get latest version
	version, dirty, errVersion := migrationDatabase.Version()
	if errVersion != nil {
		//ignore error in this line. Skip the version check
		return errors.New("Migrate.MigrateDown errVersion : " + errVersion.Error())
	}

	if dirty {
		log.Warn().Msg("dirty migration. Please clean up database")
	}

	msgLatestVersion := fmt.Sprintf("latest version is %d", version)
	log.Info().Msg(msgLatestVersion)
	return nil
}

func PrintMigrationVersion(config *MigrationConfig) error {
	driver, errDriver := mysql.WithInstance(config.Db, &mysql.Config{})
	if errDriver != nil {
		return errors.New("Migrate.PrintMigrationVersion errDriver : " + errDriver.Error())
	}

	migrationDatabase, errMigrate := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", config.MigrationDir),
		config.Dialect, driver)
	if errMigrate != nil {
		return errors.New("Migrate.PrintMigrationVersion errMigrate : " + errMigrate.Error())
	}
	version, dirty, errVersion := migrationDatabase.Version()
	if errVersion != nil {
		return errors.New("Migrate.PrintMigrationVersion errVersion : " + errVersion.Error())
	}

	msgDirty := fmt.Sprintf("Database migration version %d. Is dirty %v \n", version, dirty)
	log.Warn().Msg(msgDirty)

	return nil
}

func ShowHelp() {
	helpMessage := `
Database migration tool. Version 1.0

usage : 

migrate OPTIONS COMMAND

Options : 
 -config-path 		DB Configuration path in yml (default: db/dbconf.yml)
 -env 			Migration environment based on config path (default: development)
 -migration-dir		Migration directory files (default: db/migrations)

Command :
--up 		Up migration
--down		Down migration
--version   see migrations version


Creating migration file : 

--create --migration-dir [YOUR MIGRATION DIR] --filename [YOUR MIGRATION NAME]

--migration-dir 			Path of file (default: db/migrations)
--filename				Name of migration file  

`
	fmt.Print(helpMessage)
}
