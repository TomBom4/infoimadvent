package cmd

import (
	"log"
	"os"

	"github.com/hoffx/infoimadvent/config"
	"github.com/hoffx/infoimadvent/storage"
	"github.com/urfave/cli"
	macaron "gopkg.in/macaron.v1"
)

var Reset = cli.Command{
	Name:   "reset",
	Usage:  "resets database and session-storage",
	Action: reset,
}

func reset(ctx *cli.Context) {
	if config.Config.DB.Name == "" {
		config.Load(ctx.GlobalString("config"))
	}
	if !uStorer.Active || !qStorer.Active {
		initStorer()
	}
	err := uStorer.ResetDB()
	if err != nil {
		log.Println(err)
	}
	err = qStorer.ResetDB()
	if err != nil {
		log.Println(err)
	}
	err = os.RemoveAll(config.Config.Sessioner.StoragePath)
	if err != nil {
		log.Println(err)
	}

	if ctx.Args().First() == "web" {
		runWeb(ctx)
	}
}

func initStorer() {
	var err error
	uStorer, err = storage.NewUserStorer(config.Config.DB.Name, config.Config.DB.User, config.Config.DB.Password, macaron.Env == macaron.DEV)
	if err != nil {
		log.Fatal(err)
	}
	qStorer, err = storage.NewQuestStorer(config.Config.DB.Name, config.Config.DB.User, config.Config.DB.Password, macaron.Env == macaron.DEV)
	if err != nil {
		log.Fatal(err)
	}
}
