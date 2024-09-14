package booty

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/mroth/weightedrand"
	"log"
	"log/slog"
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

type Image struct {
	FileName string
	Rarity   Rarity
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

	bootyFolder := fmt.Sprintf("%s/%s", os.Getenv("ASSETS_FOLDER"), "booty")

	// Make sure that the booty folder exists.
	if !util.DirExists(bootyFolder) {
		err := os.Mkdir(bootyFolder, 0777)
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
		BootyFolder:  bootyFolder,
		EventManager: e,
		Store:        store,
	}
}

func (bc *BootyCommand) Run(s *discordgo.Session, m *discordgo.MessageCreate) error {

	s.
	// Upload the file to Discord
	randFile, err := bc.getRandomFile()
	if err != nil {
		slog.Error("could not find files: %v", util.ErrAttr(err))
		return errors.New("could not find files")
	}

	embed, msg, err := bc.sendMessage(s, randFile, m.ChannelID)
	if err != nil {
		slog.Error("could send message: %v", util.ErrAttr(err))
		return errors.New("could send message")
	}

	err = bc.addReaction(s, m.ChannelID, msg.ID)
	if err != nil {
		return err
	}

	hash := embed.File.Name[:strings.Index(embed.File.Name, ".")]
	imageID, err := bc.Store.GetBootyImageId(hash)
	if err != nil {
		imageID, err = bc.Store.SaveBootyImage(randFile.FileName, randFile.Rarity.Name, embed.File.ContentType, hash)
		if err != nil {
			slog.Error("could not save booty image: %v", util.ErrAttr(err))
			return errors.New("could not save booty image")
		}
	} else {
		bc.Store.IncrementPostCount(imageID)
		bc.Store.UpdateRarity(imageID, randFile.Rarity.Name)
	}

	_, err = bc.Store.SaveBootyMessage(msg.ID, msg.ChannelID, m.GuildID, imageID)
	if err != nil {
		return err
	}

	return nil
}

// CalculateRarity determines the rarity level of an image based on its popularity.
func (bc *BootyCommand) calculateWeight(fileName string) weightedrand.Choice {

	filePath := fmt.Sprintf("%s/%s", bc.BootyFolder, fileName)
	h := md5.New()
	h.Write([]byte(filePath))
	hash := hex.EncodeToString(h.Sum(nil))

	image, err := bc.Store.GetBootyImage(hash)
	if err != nil {
		return weightedrand.NewChoice(Image{FileName: fileName, Rarity: RarityTypes["common"]}, 1)
	}
	if image.PostCount == 0 {
		return weightedrand.NewChoice(Image{FileName: fileName, Rarity: RarityTypes["common"]}, 1)
	}

	rarity := RarityTypes["common"]
	weight := uint((float64(image.Likes)-float64(image.Dislikes))/float64(image.PostCount)) + 1

	switch {
	case weight > 2 && weight <= 4:
		rarity = RarityTypes["uncommon"]
	case weight > 4 && weight <= 6:
		rarity = RarityTypes["rare"]
	case weight > 6 && weight <= 8:
		rarity = RarityTypes["epic"]
	case weight > 8:
		rarity = RarityTypes["legendary"]
	default:
		rarity = RarityTypes["common"]
	}

	return weightedrand.NewChoice(&Image{FileName: fileName, Rarity: rarity}, weight)
}

func handleReaction(msg *event.MessageReactionInteraction, s *storage.Storage, removed bool) {
	if msg.Name != LikeReaction && msg.Name != DislikeReaction {
		return
	}

	if msg.Name == LikeReaction {
		if !removed {
			_, err := s.AddBootyLike(msg.MessageID)
			if err != nil {
				return
			}
			return
		}

		_, err := s.RemoveBootyLike(msg.MessageID)
		if err != nil {
			return
		}
	}

	if msg.Name == DislikeReaction {
		if !removed {
			_, err := s.AddBootyDislike(msg.MessageID)
			if err != nil {
				return
			}
			return
		}

		_, err := s.RemoveBootyDislike(msg.MessageID)
		if err != nil {
			return
		}
	}
}

func (bc *BootyCommand) createEmbed(f *os.File, rarity Rarity) (*discordgo.MessageSend, error) {

	mimeType, err := bc.getMimeType(f.Name())
	if err != nil {
		return nil, err
	}

	h := md5.New()
	h.Write([]byte(f.Name()))
	hash := fmt.Sprintf("%s%s", hex.EncodeToString(h.Sum(nil)), mimeType.Ext)

	embed := &discordgo.MessageEmbed{
		Title:       rarity.Name,
		Description: "Click image to enlarge.",
		Author:      &discordgo.MessageEmbedAuthor{},
		Color:       rarity.Value,
		Image: &discordgo.MessageEmbedImage{
			URL: fmt.Sprintf("attachment://%s", hash),
		},
		Timestamp: time.Now().Format(time.RFC3339),
		Footer: &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("Vote by clicking the %s or %s icons", LikeReaction, DislikeReaction),
		},
	}

	msg := discordgo.MessageSend{
		Embed: embed,
		File: &discordgo.File{
			Name:        hash,
			ContentType: mimeType.Type,
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

func (bc *BootyCommand) getRandomFile() (*Image, error) {
	bootyFiles, err := os.ReadDir(bc.BootyFolder)
	if err != nil {
		return nil, err
	}

	var fileList []weightedrand.Choice
	for _, file := range bootyFiles {
		if !file.IsDir() {

			filepath := fmt.Sprintf("%s/%s", bc.BootyFolder, file.Name())
			v, err := bc.getMimeType(filepath)
			if err != nil {
				continue
			}

			if v.Ext == "" {
				continue
			}

			choice := bc.calculateWeight(file.Name())
			fileList = append(fileList, choice)
		}
	}

	chooser, err := weightedrand.NewChooser(fileList...)
	if err != nil {
		log.Printf("could not choose image: %v", chooser)
		return nil, err
	}
	pick := chooser.Pick().(*Image)
	return pick, nil
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

func (bc *BootyCommand) sendMessage(s *discordgo.Session, image *Image, channelID string) (*discordgo.MessageSend, *discordgo.Message, error) {
	f, err := os.Open(fmt.Sprintf("%s/%s", bc.BootyFolder, image.FileName))
	if err != nil {
		slog.Error("could not open file", util.ErrAttr(err))
		return nil, nil, errors.New("could not open file")
	}
	defer f.Close()

	embed, err := bc.createEmbed(f, image.Rarity)
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
