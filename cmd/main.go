package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/bwmarrin/discordgo"
	"github.com/cenron/neil-bot-go/internal"
	"github.com/cenron/neil-bot-go/pkg/db"
	"github.com/cenron/neil-bot-go/pkg/util"
)

func main() {

	util.LoadEnv()

	storage := db.NewStorage(fmt.Sprintf("%s/%s", os.Getenv("ASSETS_FOLDER"), os.Getenv("DB_FILE")))
	// Used for testing DELETE
	db.CreateTable(storage.DB)

	sess, err := discordgo.New(fmt.Sprintf("Bot %s", os.Getenv("DISCORD_KEY")))
	if err != nil {
		log.Fatal(err)
	}

	// Initialize our handlers
	internal.InitHandlers(sess)

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
