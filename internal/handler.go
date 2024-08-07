package internal

import (
	"crypto/md5"
	"encoding/hex"
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/cenron/neil-bot-go/internal/command/booty"
	"github.com/cenron/neil-bot-go/pkg/event"
)

var EventManager *event.EventManager = event.NewEventManager()

type CommandInterface interface {
	Run(s *discordgo.Session, m *discordgo.MessageCreate) error
}

// Our list of commands we support.
var CommandMap = map[string]CommandInterface{
	"booty": booty.NewBootyCommand(EventManager),
}

func InitHandlers(s *discordgo.Session) {

	s.AddHandler(handleMessageCreate)

	// Reaction added
	s.AddHandler(handleAddReaction)

	// Reaction removed
	s.AddHandler(handleRemoveReaction)
}

func handleMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	cmd_array := strings.Split(m.Content, " ")
	if len(cmd_array) < 2 {
		return
	}

	if cmd_array[0] != "neil" {
		return
	}

	handler, ok := CommandMap[cmd_array[1]]
	if !ok {
		return
	}

	handler.Run(s, m)
}

func handleAddReaction(s *discordgo.Session, r *discordgo.MessageReactionAdd) {
	if s.State.User.ID == r.UserID {
		return
	}

	hasher := md5.New()
	hasher.Write([]byte(r.Emoji.Name))
	log.Printf("Got reaction: msg: %s - emoji: %s", r.MessageID, hex.EncodeToString(hasher.Sum(nil)))

	EventManager.Emit(event.ADD_REACTION, hex.EncodeToString(hasher.Sum(nil)))

}

func handleRemoveReaction(s *discordgo.Session, r *discordgo.MessageReactionRemove) {
	if s.State.User.ID == r.UserID {
		return
	}

	hasher := md5.New()
	hasher.Write([]byte(r.Emoji.Name))
	log.Printf("Reaction removed: msg: %s - emoji: %s", r.MessageID, hex.EncodeToString(hasher.Sum(nil)))

	EventManager.Emit(event.REMOVE_REACTION, hex.EncodeToString(hasher.Sum(nil)))

}
