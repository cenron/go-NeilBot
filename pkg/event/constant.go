package event

const (
	ADD_REACTION    = "add_reaction"
	REMOVE_REACTION = "remove_reaction"
)

type MessageReactionInteraction struct {
	Hash      string
	Name      string
	UserID    string
	MessageID string
	ChannelID string
	GuildID   string
}
