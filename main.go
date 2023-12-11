package main

import (
	"fmt"
	"log"
	"os"

	"boilerplate/database"
	"boilerplate/router"
	"boilerplate/utils"
	"github.com/joho/godotenv"
	cli "github.com/urfave/cli/v2"
)

var version string = "development"

func initializer() {
	if _, err := os.Stat(".env"); err == nil {
		log.Println("Loading the config from .env file")
		err = godotenv.Load(".env")

		if err != nil {
			log.Println("Error loading .env config file")
		}
		log.Println("Successfully loaded the config file")
	}
	// ********************************************************************* \\
	// defer database.CloseDb() missing
	// ********************************************************************* \\
	database.ConnectDb()
}

func main() {

	var v bool
	app := &cli.App{
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:        "debug",
				Aliases:     []string{"d"},
				Value:       false,
				Usage:       "Run esgscore in DEBUG mode",
				Destination: &utils.Debug,
			},
			&cli.BoolFlag{
				Name:        "version",
				Aliases:     []string{"v"},
				Value:       false,
				Usage:       "Print version of current binary",
				Destination: &v,
			},
		},
		Action: func(cCtx *cli.Context) error {
			if v {
				printVersion()
				os.Exit(0)
			} else {
				initializer()
				router.ClientRoutes()
			}
			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func printVersion() {
	fmt.Printf("Current Version: %s\n", version)
}
