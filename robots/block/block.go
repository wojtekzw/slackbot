package robots

import (
	"fmt"
	"strings"

	"github.com/wojtekzw/slackbot/robots"
	"github.com/wojtekzw/slackbot/rtm"
)

type bot struct {
	config *rtm.BlockConfig
}

// TODO - nie używać zmiennej globalne rtm.Config - tylko jak ?
func init() {
	r := &bot{config: rtm.Config}

	robots.RegisterRobot("block", r)
}

func (r bot) Run(p *robots.Payload) (slashCommandImmediateReturn string) {
	// go r.DeferredAction(p)
	// log.Printf("Block robot run: %s\n", slashCommandImmediateReturn)
	return r.blockCommand(p)
}

func (r bot) DeferredAction(p *robots.Payload) {
	response := &robots.IncomingWebhook{
		Domain:      p.TeamDomain,
		Channel:     p.ChannelID,
		Username:    "Block Bot",
		Text:        "Error - unsed func",
		IconEmoji:   ":ghost:",
		UnfurlLinks: true,
		Parse:       robots.ParseStyleFull,
	}

	response.Send()
}

func (r bot) Description() (description string) {
	return "Block write to a channel\n\tUsage: /block add|remove|list|refresh|help"
}

func (r bot) listCommand() string {
	adminsStr := strings.Join(r.config.AdminNames(), ",")
	allowedUsersStr := strings.Join(r.config.AllowedUsersNames(), ", ")

	return fmt.Sprintf("Blocked channel: %s\nAdmins: %s\nAllowed users: %s\n", r.config.Channel.Name, adminsStr, allowedUsersStr)
}

func (r bot) blockCommand(p *robots.Payload) (result string) {
	inText := strings.ToLower(strings.TrimSpace(p.Text))
	outText := ""

	switch inText {
	case "add":
		outText = inText + ": Add command"
	case "remove":
		outText = inText + ": Remove command"
	case "list":
		outText = r.listCommand()
	case "refresh":
		outText = inText + ": Refresh command"
	case "help":
		outText = r.Description()

	default:
		outText = inText + ": Unknown command"
	}
	return outText
}
