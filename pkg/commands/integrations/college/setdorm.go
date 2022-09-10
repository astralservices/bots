package integrations

import (
	"os"

	"github.com/astralservices/bots/pkg/commands/integrations"
	db "github.com/astralservices/bots/pkg/database/supabase"
	"github.com/astralservices/bots/pkg/utils"
	"github.com/astralservices/dgc"
	"github.com/bwmarrin/discordgo"
)

var SetDormCommand = &dgc.Command{
	Name:        "setdorm",
	Domain:      "astral.integrations.setdorm",
	Aliases:     []string{"setdorm", "setdormitory"},
	Description: "Set your dorm house and room!",
	Category:    "College",
	Usage:       "setdorm <house> <room>",
	Slash:       true,
	SlashGuilds: []string{os.Getenv("DEV_GUILD")},
	Handler: func(ctx *dgc.Ctx) {
		house, room := ctx.Arguments.Get(0).Raw(), ctx.Arguments.Get(1).Raw()

		database := db.New()

		wiID, err := integrations.GetWorkspaceIntegrationForCommand(ctx, CollegeIntegrationID)

		if err != nil {
			ctx.ReplyEmbed(utils.GenerateEmbed(*ctx, discordgo.MessageEmbed{
				Title:       "Error",
				Description: "An error occurred while fetching the workspace integration.",
				Fields: []*discordgo.MessageEmbedField{
					{
						Name:   "Error",
						Value:  err.Error(),
						Inline: false,
					},
				},
				Color: 0xff0000,
			}))

			return
		}

		err = database.SetIntegrationDataForUser(ctx.Event.Author.ID, CollegeIntegrationID, wiID, map[string]interface{}{
			"house": house,
			"room":  room,
		})

		if err != nil {
			ctx.ReplyEmbed(utils.GenerateEmbed(*ctx, discordgo.MessageEmbed{
				Title:       "Error",
				Description: "An error occurred while setting your dorm.",
				Fields: []*discordgo.MessageEmbedField{
					{
						Name:   "Error",
						Value:  err.Error(),
						Inline: false,
					},
				},
				Color: 0xff0000,
			}))

			return
		}

		ctx.ReplyEmbed(utils.GenerateEmbed(*ctx, discordgo.MessageEmbed{
			Title:       "Success",
			Description: "Your dorm has been set to " + house + " " + room + "!",
			Color:       0x00ff00,
		}))
	},
}
