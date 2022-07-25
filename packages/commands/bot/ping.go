package commands

import (
	"fmt"
	"strings"
	"time"

	db "github.com/astralservices/bots/supabase"
	"github.com/astralservices/bots/utils"
	"github.com/bwmarrin/discordgo"
	"github.com/zekroTJA/shireikan"
)

// Ping is a command responding with a ping
// message in the commands channel.
type Ping struct {
}

// GetInvoke returns the command invokes.
func (c *Ping) GetInvokes() []string {
	return []string{"ping", "p"}
}

// GetDescription returns the commands description.
func (c *Ping) GetDescription() string {
	return "Retrieves the gateway and API latency"
}

// GetHelp returns the commands help text.
func (c *Ping) GetHelp() string {
	return "`ping` - ping"
}

// GetGroup returns the commands group.
func (c *Ping) GetGroup() string {
	return utils.CategoryBot
}

// GetDomainName returns the commands domain name.
func (c *Ping) GetDomainName() string {
	return "internal.bot.ping"
}

// GetSubPermissionRules returns the commands sub
// permissions array.
func (c *Ping) GetSubPermissionRules() []shireikan.SubPermission {
	return nil
}

// IsExecutableInDMChannels returns whether
// the command is executable in DM channels.
func (c *Ping) IsExecutableInDMChannels() bool {
	return false
}

// Exec is the commands execution handler.
func (c *Ping) Exec(ctx shireikan.Context) error {
	bot := ctx.GetObject("bot").(utils.IBot)

	start := time.Now()
	m, err := utils.ReplyWithEmbed(ctx, utils.GenerateEmbed(ctx, discordgo.MessageEmbed{
		Title: "Pinging...",
		Color: 0xffff00,
	}))

	if err != nil {
		utils.ErrorHandler(err)
	}

	end := time.Now()
	diff := end.Sub(start)

	m, err = ctx.GetSession().ChannelMessageEditEmbed(m.ChannelID, m.ID, utils.GenerateEmbed(ctx, discordgo.MessageEmbed{
		Title:       "Pong!",
		Description: fmt.Sprintf(":ping_pong: Gateway Ping: `%dms`\n:desktop: API Ping: `%dms`", ctx.GetSession().HeartbeatLatency().Milliseconds(), diff.Milliseconds()),
		Color:       0xffff00,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:  "Region",
				Value: "Fetching...",
			},
		},
	}))

	if err != nil {
		utils.ErrorHandler(err)
	}

	database := db.New()

	var region utils.IRegion

	database.DB.From("regions").Select("*").Single().Eq("id", bot.Region).Execute(&region)

	regionString := region.Flag + " " + strings.ToUpper(strings.Split(region.ID, ".")[0]) + " (" + region.City + ", " + region.Region + ", " + region.Country + ")"

	_, err = ctx.GetSession().ChannelMessageEditEmbed(m.ChannelID, m.ID, utils.GenerateEmbed(ctx, discordgo.MessageEmbed{
		Title:       "Pong!",
		Description: fmt.Sprintf(":ping_pong: Gateway Ping: `%dms`\n:desktop: API Ping: `%dms`", ctx.GetSession().HeartbeatLatency().Milliseconds(), diff.Milliseconds()),
		Color:       0x00ff00,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:  "Region",
				Value: regionString,
			},
		},
	}))

	if err != nil {
		utils.ErrorHandler(err)
	}

	return err
}
