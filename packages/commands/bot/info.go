package commands

import (
	db "github.com/astralservices/bots/supabase"
	"github.com/astralservices/bots/utils"
	"github.com/astralservices/bots/utils/constants"
	"github.com/bwmarrin/discordgo"
	"github.com/zekroTJA/shireikan"
)

type Info struct {
}

// GetInvoke returns the command invokes.
func (c *Info) GetInvokes() []string {
	return []string{"info"}
}

// GetDescription returns the commands description.
func (c *Info) GetDescription() string {
	return "Get generic info about the bot"
}

// GetHelp returns the commands help text.
func (c *Info) GetHelp() string {
	return "`info` - info"
}

// GetGroup returns the commands group.
func (c *Info) GetGroup() string {
	return utils.CategoryBot
}

// GetDomainName returns the commands domain name.
func (c *Info) GetDomainName() string {
	return "astral.bot.info"
}

// GetSubPermissionRules returns the commands sub
// permissions array.
func (c *Info) GetSubPermissionRules() []shireikan.SubPermission {
	return nil
}

// IsExecutableInDMChannels returns whether
// the command is executable in DM channels.
func (c *Info) IsExecutableInDMChannels() bool {
	return false
}

// Exec is the commands execution handler.
func (c *Info) Exec(ctx shireikan.Context) error {
	bot := ctx.GetObject("bot").(utils.IBot)

	// code

	embed := utils.GenerateEmbed(ctx, discordgo.MessageEmbed{
		Title:       "Astral",
		Description: "A multi-purpose Discord bot written in Go.",
		Color:       0x00ff00,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:  "Region",
				Value: "Loading...",
			},
			{
				Name:  "Version",
				Value: constants.VERSION,
			},
			{
				Name:  "Authors",
				Value: "AmusedGrape#0001",
			},
		},
	})

	m, err := utils.ReplyWithEmbed(ctx, embed)

	if err != nil {
		utils.ErrorHandler(err)
	}

	var region utils.IRegion

	database := db.New()

	err = database.DB.From("regions").Select("*").Single().Eq("id", bot.Region).Execute(&region)

	if err != nil {
		utils.ErrorHandler(err)
	}

	m, err = ctx.GetSession().ChannelMessageEditEmbed(m.ChannelID, m.ID, utils.GenerateEmbed(ctx, discordgo.MessageEmbed{
		Title:       "Astral",
		Description: "A multi-purpose Discord bot written in Go.",
		Color:       0x00ff00,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:  "Region",
				Value: region.Flag + " " + region.PrettyName,
			},
			{
				Name:  "Version",
				Value: constants.VERSION,
			},
			{
				Name:  "Authors",
				Value: "AmusedGrape#0001",
			},
		},
	}))

	return err
}
