package rtm

import (
	"fmt"
	"os"
	"strings"

	"log"

	"github.com/nlopes/slack"
	"github.com/wojtekzw/slackbot/utils"
)

var (
	Config = &BlockConfig{}
)

type NameID struct {
	Name string
	ID   string
}
type BlockConfig struct {
	Channel      NameID
	Owner        NameID
	Admins       []NameID
	AllowedUsers []NameID
	DeletedMsg   string
	users        utils.GlobalUsers
	groups       utils.GlobalChannels
}

// FIXME: Trim && ToLower all env strings
// walidacja danych zewnętrznych - co robic w razie błedów
func (b *BlockConfig) ReadBlockChannelConfig(api *slack.Client) {

	b.Channel.Name = os.Getenv("SLACKBOT_BLOCK_CHANNEL_NAME")

	b.Owner.Name = os.Getenv("SLACKBOT_OWNER_NAME")

	defAdminName := os.Getenv("SLACKBOT_ADMIN_NAME")
	for _, elem := range strings.Fields(defAdminName) {
		b.Admins = append(b.Admins, NameID{Name: elem})
	}

	defUserName := os.Getenv("SLACKBOT_ALLOWED_USER_NAME")
	for _, elem := range strings.Fields(defUserName) {
		b.AllowedUsers = append(b.AllowedUsers, NameID{Name: elem})
	}

	b.DeletedMsg = os.Getenv("SLACKBOT_DELETED_MSG")

	log.Printf("Admins: %v len %v cap %v\n", b.Admins, len(b.Admins), cap(b.Admins))
	log.Printf("Allowed: %v  len %v cap %v\n", b.AllowedUsers, len(b.AllowedUsers), cap(b.AllowedUsers))

	b.convertIDsToNames(api)

	log.Printf("Admins 2: %v len %v cap %v\n", b.Admins, len(b.Admins), cap(b.Admins))
	log.Printf("Allowed 2: %v  len %v cap %v\n", b.AllowedUsers, len(b.AllowedUsers), cap(b.AllowedUsers))

}

func (b *BlockConfig) convertIDsToNames(api *slack.Client) {

	// TODO - refresh users and channels every 1h !!!! or on demand (refresh command)

	b.users.GetUsers(api)

	b.groups.GetChannels(api)

	// convert names to ID's
	b.Channel.ID = b.groups.NameToID(b.Channel.Name[1:])

	b.Owner.ID = b.users.NameToID(b.Owner.Name[1:])

	for i := range b.Admins {
		b.Admins[i].ID = b.users.NameToID(b.Admins[i].Name[1:])
	}
	for i := range b.AllowedUsers {
		b.AllowedUsers[i].ID = b.users.NameToID(b.AllowedUsers[i].Name[1:])
	}

}

func (b *BlockConfig) IsOwner(id string) bool {
	return b.Owner.ID == id
}

func (b *BlockConfig) IsAdmin(id string) bool {
	for i := range b.Admins {
		if b.Admins[i].ID == id {
			return true
		}
	}
	return false
}

func (b *BlockConfig) IsAllowedUser(id string) bool {
	for i := range b.AllowedUsers {
		if b.AllowedUsers[i].ID == id {
			return true
		}
	}
	return false
}

func (b *BlockConfig) IsBlockedChannel(id string) bool {
	log.Printf("Config blocked: %s (%s), mgs channel: %s (%s)\n", b.Channel.ID, b.groups.IDToName(b.Channel.ID), id, b.groups.IDToName(id))
	return b.Channel.ID == id
}

func (b *BlockConfig) IsAllowedWrite(id string) bool {
	log.Printf("isOwner: %t, isAdmin: %t, isAllowedUser: %t\n", b.IsOwner(id), b.IsAdmin(id), b.IsAllowedUser(id))
	return b.IsOwner(id) || b.IsAdmin(id) || b.IsAllowedUser(id)
}

func (b *BlockConfig) IsNotAllowedWrite(id string) bool {
	return !b.IsAllowedWrite(id)
}
func (b *BlockConfig) AdminNames() []string {
	var admins []string
	for i := range b.Admins {
		admins = append(admins, b.Admins[i].Name)
	}
	return admins
}

func (b *BlockConfig) AllowedUsersNames() []string {
	var users []string
	for i := range b.AllowedUsers {
		users = append(users, b.AllowedUsers[i].Name)
	}
	return users
}

// RunRTM - listen to RTM events and remove messages
func RunRTM() {
	apiToken := os.Getenv("SLACKBOT_API_TOKEN")

	if len(apiToken) == 0 {
		log.Printf("SLACKBOT_API_TOKEN not set. RTM module not run")
		return
	}

	slackbotRTMDebug := false
	if os.Getenv("SLACKBOT_RTM_DEBUG") == "true" {
		slackbotRTMDebug = true
	}

	api := slack.New(apiToken)
	// api.SetDebug(true)

	// FIXME: Global variable Config
	Config.ReadBlockChannelConfig(api)

	rtm := api.NewRTM()
	go rtm.ManageConnection()

Loop:
	for {
		select {
		case msg := <-rtm.IncomingEvents:

			// Print pretty events
			if slackbotRTMDebug {
				log.Printf("Event Received: %s\n", utils.StructPrettyPrint(msg))
			}

			switch ev := msg.Data.(type) {
			case *slack.HelloEvent:
				// Ignore hello

			case *slack.ConnectedEvent:
				// rtm.SendMessage(rtm.NewOutgoingMessage("Hello world", "C0FCTCZNK"))

			case *slack.AckMessage:
				// log.Println("Ack:", ev.Info)

			case *slack.MessageEvent:
				if (len(ev.SubType) == 0 || ev.SubType == "bot_message") && !ev.Hidden {
					// Only empty subtype (ordinary message) or "bot_message" are passed through
					// to be deleted
					// If Hidden - do nothing - because users don't see it

					if Config.IsBlockedChannel(ev.Channel) && Config.IsNotAllowedWrite(ev.User) {
						if slackbotRTMDebug || true {
							log.Printf("Message to delete: %s\n", utils.StructPrettyPrint(msg))
						}
						schan, ts, err := api.DeleteMessage(ev.Channel, ev.Timestamp)
						if err != nil {
							log.Printf("Error deleting message. Chan: %s, ts: %s, err: %v\n", schan, ts, err)
						} else {
							if ev.SubType == "bot_message" {
								// Don't answer to bot
								break
							}

							replyUser, _ := api.GetUserInfo(ev.User)
							replyUserAtName := "@" + replyUser.Name

							if slackbotRTMDebug {
								log.Printf("Event Received: %s\n", utils.StructPrettyPrint(replyUser))
							}

							params := slack.PostMessageParameters{AsUser: false, Username: "Block Bot"}
							attachment := slack.Attachment{
								Pretext: "Deleted message is below:",
								Text:    ev.Text,
							}

							params.Attachments = []slack.Attachment{attachment}
							// channelID, timestamp, err := api.PostMessage(replyToChannel,
							api.PostMessage(replyUserAtName,
								fmt.Sprintf("Your message was deleted from '%s' channel. You are not allowed to publish there. %s", Config.Channel.Name, Config.DeletedMsg), params)
						}
					}
				}
			case *slack.PresenceChangeEvent:
				Config.users.SetPresenceByID(ev.User, ev.Presence)
				// log.Printf("Presence Change: User: %s %s\n", users.IDToName(ev.User), users.GetPresenceByID(ev.User))

			case *slack.LatencyReport:
				// log.Printf("Current latency: %v\n", ev.Value)

			case *slack.RTMError:
				// log.Printf("Error: %s\n", ev.Error())

			case *slack.InvalidAuthEvent:
				log.Printf("Invalid credentials")
				break Loop

			default:

				// Ignore other events..
				// log.Printf("Unexpected: [%v] %v\n", ev, msg.Data)
			}
		}
	}
}
