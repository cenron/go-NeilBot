package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/bwmarrin/discordgo"
	"github.com/cenron/neil-bot-go/internal"
	"github.com/cenron/neil-bot-go/pkg/storage"
	"github.com/cenron/neil-bot-go/pkg/util"
	"github.com/jmoiron/sqlx"
)

func main() {

	util.LoadEnv()

	db := storage.NewDB(fmt.Sprintf("%s/%s", os.Getenv("ASSETS_FOLDER"), os.Getenv("DB_FILE")))
	defer func(db *sqlx.DB) {
		err := db.Close()
		if err != nil {
			os.Exit(1)
		}
	}(db)

	storage.RunMigration(db)
	store := storage.NewStorage(db)

	sess, err := discordgo.New(fmt.Sprintf("Bot %s", os.Getenv("DISCORD_KEY")))
	if err != nil {
		log.Fatal(err)
	}

	// Initialize our handlers
	internal.InitHandlers(sess, store)

	sess.Identify.Intents = discordgo.IntentsAllWithoutPrivileged
	err = sess.Open()
	if err != nil {
		log.Fatal(err)
	}

	defer sess.Close()

	fmt.Println(("the  bot is online!"))

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	log.Println("Press Ctrl+C to exit")
	<-stop
}
