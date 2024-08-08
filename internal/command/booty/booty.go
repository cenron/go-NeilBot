package booty

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/cenron/neil-bot-go/pkg/event"
	"github.com/cenron/neil-bot-go/pkg/storage"
	"github.com/cenron/neil-bot-go/pkg/util"
)

const (
	LikeReaction    = "👍"
	DislikeReaction = "👎"
)

var MimeToExt = map[string]string{
	"image/png":  ".png",
	"image/bmp":  ".bmp",
	"image/gif":  ".gif",
	"image/jpeg": ".jpeg",
	"image/webp": ".webp",
}

var RarityTypes = map[string]Rarity{
	"common": {
		Name:  "Common",
		Value: 0xDEDEDE,
	},
	"uncommon": {
		Name:  "Uncommon",
		Value: 0x1eff00,
	},
	"rare": {
		Name:  "Rare",
		Value: 0x0070dd,
	},
	"epic": {
		Name:  "Epic",
		Value: 0xa335ee,
	},
	"legendary": {
		Name:  "Legendary",
		Value: 0xff8000,
	},
}

type MimeType struct {
	Type string
	Ext  string
}

type Rarity struct {
	Name  string
	Value int
}

type BootyCommand struct {
	MimeToExt    map[string]string
	BootyFolder  string
	RarityTypes  map[string]Rarity
	EventManager *event.EventManager
	Store        *storage.Storage
}

func NewBootyCommand(e *event.EventManager, store *storage.Storage) *BootyCommand {

	util.LoadEnv()

	bootyfolder := fmt.Sprintf("%s/%s", os.Getenv("ASSETS_FOLDER"), "booty")

	// Make sure that the booty folder exists.
	if !util.DirExists(bootyfolder) {
		err := os.Mkdir(bootyfolder, 0777)
		if err != nil && !os.IsExist(err) {
			log.Fatal(err)
		}
	}

	// Register our event handlers
	e.Register(event.ADD_REACTION, func(msg interface{}) {
		if msgreaction, ok := msg.(event.MessageReactionInteraction); ok {
			handleReaction(&msgreaction, store, false)
		}
	})
	e.Register(event.REMOVE_REACTION, func(msg interface{}) {
		if msgreaction, ok := msg.(event.MessageReactionInteraction); ok {
			handleReaction(&msgreaction, store, true)
		}
	})

	return &BootyCommand{
		MimeToExt:    MimeToExt,
		RarityTypes:  RarityTypes,
		BootyFolder:  bootyfolder,
		EventManager: e,
		Store:        store,
	}
}

func (bc *BootyCommand) Run(s *discordgo.Session, m *discordgo.MessageCreate) error {

	// Upload the file to Discord
	randfile, err := bc.getRandomFile()
	if err != nil {
		slog.Error("could not find files: %v", util.ErrAttr(err))
		return errors.New("could not find files")
	}

	embed, msg, err := bc.sendMessage(s, randfile, m.ChannelID)
	if err != nil {
		slog.Error("could send message: %v", util.ErrAttr(err))
		return errors.New("could send message")
	}

	err = bc.addReaction(s, m.ChannelID, msg.ID)
	if err != nil {
		return err
	}

	imageID, err := bc.Store.SaveBootyImage(randfile, embed.File.ContentType, embed.File.Name[:strings.Index(embed.File.Name, ".")])
	if err != nil {
		slog.Error("could not save booty image: %v", util.ErrAttr(err))
		return errors.New("could not save booty image")
	}

	_, err = bc.Store.SaveBootyMessage(msg.ID, msg.ChannelID, m.GuildID, imageID)
	if err != nil {
		return err
	}

	return nil
}

func handleReaction(msg *event.MessageReactionInteraction, s *storage.Storage, removed bool) {
	if msg.Name != LikeReaction && msg.Name != DislikeReaction {
		return
	}

	if !removed {
		fmt.Printf("Add reaction: %+v\n", msg)
		_, err := s.AddBootyLike(msg.MessageID)
		if err != nil {
			return
		}
		return
	}

	fmt.Printf("Removed reaction: %+v\n", msg)

}

func (bc *BootyCommand) createEmbed(f *os.File, color int) (*discordgo.MessageSend, error) {

	mimietype, err := bc.getMimeType(f.Name())
	if err != nil {
		return nil, err
	}

	hasher := md5.New()
	hasher.Write([]byte(f.Name()))
	hash := fmt.Sprintf("%s%s", hex.EncodeToString(hasher.Sum(nil)), mimietype.Ext)

	embed := &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{},
		Color:  color,
		Image: &discordgo.MessageEmbedImage{
			URL: fmt.Sprintf("attachment://%s", hash),
		},
		Timestamp: time.Now().Format(time.RFC3339),
		Footer:    &discordgo.MessageEmbedFooter{},
	}

	msg := discordgo.MessageSend{
		Embed: embed,
		File: &discordgo.File{
			Name:        hash,
			ContentType: mimietype.Type,
			Reader:      f,
		},
	}

	return &msg, nil
}

func (bc *BootyCommand) getMimeType(filepath string) (mt *MimeType, err error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	buffer := make([]byte, 512)
	file.Read(buffer)

	mimetype := http.DetectContentType(buffer)

	return &MimeType{
		Type: mimetype,
		Ext:  bc.MimeToExt[mimetype],
	}, nil
}

func (bc *BootyCommand) getRandomFile() (string, error) {
	bootyfiles, err := os.ReadDir(bc.BootyFolder)
	if err != nil {
		return "", err
	}

	var filelist []string
	for _, file := range bootyfiles {
		if !file.IsDir() {
			filelist = append(filelist, file.Name())
		}
	}

	return filelist[rand.Intn(len(filelist))], nil
}

func (bc *BootyCommand) addReaction(s *discordgo.Session, channelID, messageID string) error {

	err := s.MessageReactionAdd(channelID, messageID, LikeReaction)
	if err != nil {
		return err
	}

	err = s.MessageReactionAdd(channelID, messageID, DislikeReaction)
	if err != nil {
		return err
	}

	return nil
}

func (bc *BootyCommand) sendMessage(s *discordgo.Session, file string, channelID string) (*discordgo.MessageSend, *discordgo.Message, error) {
	f, err := os.Open(fmt.Sprintf("%s/%s", bc.BootyFolder, file))
	if err != nil {
		slog.Error("could not open file", util.ErrAttr(err))
		return nil, nil, errors.New("could not open file")
	}
	defer f.Close()

	embed, err := bc.createEmbed(f, RarityTypes["common"].Value)
	if err != nil {
		slog.Error("could not create embed: %v", util.ErrAttr(err))
		return nil, nil, errors.New("could not create embed")
	}

	msg, err := s.ChannelMessageSendComplex(channelID, embed)
	if err != nil {
		slog.Error("could not send message: %s", util.ErrAttr(err))
		return nil, nil, errors.New("could not send message")
	}

	return embed, msg, nil
}
