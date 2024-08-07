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
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/cenron/neil-bot-go/pkg/event"
	"github.com/cenron/neil-bot-go/pkg/util"
)

var MimeToExt = map[string]string{
	"image/png":  ".png",
	"image/bmp":  ".bmp",
	"image/gif":  ".gif",
	"image/jpeg": ".jpeg",
	"image/webp": ".webp",
}

var RarityTypes = []Rarity{
	{
		Name:  "Common",
		Value: 0xDEDEDE,
	},
	{
		Name:  "Uncommon",
		Value: 0x1eff00,
	},
	{
		Name:  "Rare",
		Value: 0x0070dd,
	},
	{
		Name:  "Epic",
		Value: 0xa335ee,
	},
	{
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
	RarityTypes  []Rarity
	EventManager *event.EventManager
}

func NewBootyCommand(e *event.EventManager) *BootyCommand {

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
			handleAddReaction(&msgreaction)
		}
	})
	e.Register(event.REMOVE_REACTION, func(msg interface{}) {
		if msgreaction, ok := msg.(event.MessageReactionInteraction); ok {
			handleRemoveReaction(&msgreaction)
		}
	})

	return &BootyCommand{
		MimeToExt:    MimeToExt,
		RarityTypes:  RarityTypes,
		BootyFolder:  bootyfolder,
		EventManager: e,
	}
}

func (bc *BootyCommand) Run(s *discordgo.Session, m *discordgo.MessageCreate) error {

	// Upload the file to Discord
	randfile, err := bc.getRandomFile()
	if err != nil {
		slog.Error("could not find files: %v", util.ErrAttr(err))
		return errors.New("could not find files")
	}

	file, err := os.Open(fmt.Sprintf("%s/%s", bc.BootyFolder, randfile))
	if err != nil {
		slog.Error("could not open file", util.ErrAttr(err))
		return errors.New("could not open file")
	}
	defer file.Close()

	embed, err := bc.createEmbed(file, 0xBF40BF)
	if err != nil {
		slog.Error("could not create embed: %v", util.ErrAttr(err))
		return errors.New("could not create embed")
	}

	msg, err := s.ChannelMessageSendComplex(m.ChannelID, embed)
	if err != nil {
		slog.Error("could not send message: %s", util.ErrAttr(err))
		return errors.New("could not send message")
	}

	s.MessageReactionAdd(m.ChannelID, msg.ID, "👍")
	s.MessageReactionAdd(m.ChannelID, msg.ID, "👎")

	return nil
}

func handleAddReaction(msg *event.MessageReactionInteraction) {
	fmt.Printf("Add reaction: %+v\n", msg)
}

func handleRemoveReaction(msg *event.MessageReactionInteraction) {
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
