package main

import (
	"flag"
	"github.com/AgieAja/go-migration-database/migrates"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
)

func main() {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	log.Logger = log.Output(
		zerolog.ConsoleWriter{
			Out:     os.Stderr,
			NoColor: false,
		},
	)

	err := godotenv.Load("config/.env")
	if err != nil {
		log.Error().Msg("Failed read configuration database")
		return
	}

	pathMigration := os.Getenv("PATH_MIGRATION")
	migrationDir := flag.String("migration-dir", pathMigration, "migration directory")

	upMigration := flag.Bool("up", false, "Up migration flag")
	downMigration := flag.Bool("down", false, "Down migration flag")
	versionMigration := flag.Bool("version", false, "Version of migration flag")

	newMigrationFile := flag.Bool("create", false, "Create new migration file")
	newMigrationFileName := flag.String("filename", "", "New migration file name")
	flag.Parse()

	if *newMigrationFile {
		if *newMigrationFileName == "" {
			log.Error().Msg("please specify migration file name with --filename")
			migrates.ShowHelp()
			return
		}

		//create new migration file
		err := migrates.CreateNewMigrationFile(*migrationDir, *newMigrationFileName)
		if err != nil {
			log.Error().Msg("failed to create migration file " + err.Error())
		}

		return
	}

	//check if at least up or down flag is specified
	if !(*upMigration || *downMigration || *versionMigration) {
		log.Error().Msg("please specify --up or --down for migration")
		migrates.ShowHelp()
		return
	}

	//check migration direction up/down
	if *upMigration && *downMigration {
		log.Warn().Msg("use --up or --down at once only")
		migrates.ShowHelp()
		return
	}

	//setting db config
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASS")
	dbName := os.Getenv("DB_NAME")
	dbDriver := os.Getenv("DB_DRIVER")
	migrationConf, errMigrationConf := migrates.NewMigrationConfig(*migrationDir, dbHost, dbPort, dbUser, dbPass, dbName, dbDriver)
	if errMigrationConf != nil {
		log.Error().Msg(errMigrationConf.Error())
		return
	}
	defer func() {
		errConnClose := migrationConf.Db.Close()
		if errConnClose != nil {
			log.Error().Msg("errConnClose : " + errConnClose.Error())
		}
	}()

	if *upMigration {
		err = migrates.MigrateUp(migrationConf)
		if err != nil {
			log.Error().Msg(err.Error())
			return
		}
	} else if *downMigration {
		err = migrates.MigrateDown(migrationConf)
		if err != nil {
			log.Error().Msg(err.Error())
			return
		}
	} else if *versionMigration {
		err = migrates.PrintMigrationVersion(migrationConf)
		if err != nil {
			log.Error().Msg(err.Error())
			return
		}
	}
}
