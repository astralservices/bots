package commands

import (
	"log"

	db "github.com/astralservices/bots/supabase"
	"github.com/astralservices/bots/utils"
	"github.com/bwmarrin/discordgo"
	"github.com/zekroTJA/shireikan"
)

type Region struct {
}

// GetInvoke returns the command invokes.
func (c *Region) GetInvokes() []string {
	return []string{"region", "reg"}
}

// GetDescription returns the commands description.
func (c *Region) GetDescription() string {
	return "Retrieves the gateway and API latency"
}

// GetHelp returns the commands help text.
func (c *Region) GetHelp() string {
	return "`region` - get region for current bot \n" +
		"`region <region>` - get region info for specified region"
}

// GetGroup returns the commands group.
func (c *Region) GetGroup() string {
	return utils.CategoryBot
}

// GetDomainName returns the commands domain name.
func (c *Region) GetDomainName() string {
	return "internal.bot.region"
}

// GetSubPermissionRules returns the commands sub
// permissions array.
func (c *Region) GetSubPermissionRules() []shireikan.SubPermission {
	return nil
}

// IsExecutableInDMChannels returns whether
// the command is executable in DM channels.
func (c *Region) IsExecutableInDMChannels() bool {
	return false
}

// Exec is the commands execution handler.
func (c *Region) Exec(ctx shireikan.Context) error {
	bot := ctx.GetObject("bot").(utils.IBot)

	m, err := utils.ReplyWithEmbed(ctx, utils.GenerateEmbed(ctx, discordgo.MessageEmbed{
		Title: "Fetching...",
		Color: 0xffff00,
	}))

	if err != nil {
		utils.ErrorHandler(err)
	}

	database := db.New()

	var regions []utils.IRegion

	var regionId string

	if len(ctx.GetArgs()) == 0 {
		regionId = bot.Region
	} else {
		regionId = ctx.GetArgs()[0]
	}

	err = database.DB.From("regions").Select("*").Eq("id", regionId).Execute(&regions)

	if err != nil {
		utils.ErrorHandler(err)
	}

	log.Println(regions)

	if len(regions) == 0 {
		m, err = ctx.GetSession().ChannelMessageEditEmbed(m.ChannelID, m.ID, utils.GenerateEmbed(ctx, discordgo.MessageEmbed{
			Title:       "Region not found",
			Description: "The specified region was not found.",
			Color:       0xff0000,
		}))
		return err
	}

	region := regions[0]

	_, err = ctx.GetSession().ChannelMessageEditEmbed(m.ChannelID, m.ID, utils.GenerateEmbed(ctx, discordgo.MessageEmbed{
		Title: "Region Info for " + region.Flag + " `" + region.ID + "`",
		Color: 0x00ff00,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:  "Location",
				Value: region.City + ", " + region.Region + ", " + region.Country,
			},
		},
	}))

	if err != nil {
		utils.ErrorHandler(err)
	}

	return err
}
