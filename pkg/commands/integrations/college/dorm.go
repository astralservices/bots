package integrations

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/astralservices/bots/pkg/commands/integrations"
	db "github.com/astralservices/bots/pkg/database/supabase"
	"github.com/astralservices/bots/pkg/types"
	"github.com/astralservices/bots/pkg/utils"
	"github.com/astralservices/dgc"
	"github.com/bwmarrin/discordgo"
)

var DormCommand = &dgc.Command{
	Name:          "dorm",
	Domain:        "astral.integrations.dorm",
	Aliases:       []string{"dorm", "dormitory"},
	Description:   "Get a user's dorm house and room!",
	Category:      "College",
	Usage:         "dorm <user>",
	Slash:         true,
	SlashGuilds:   []string{os.Getenv("DEV_GUILD")},
	IntegrationID: CollegeIntegrationID,
	Handler: func(ctx *dgc.Ctx) {
		db := db.New()

		var user string

		if ctx.Arguments.Amount() < 1 {
			user = ctx.Message.Author.ID
		} else {
			user = ctx.Arguments.Get(0).AsUserMentionID()
		}

		wi, err := integrations.GetWorkspaceIntegrationForCommand(ctx, CollegeIntegrationID)

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

		data, err := db.GetIntegrationDataForUser(user, CollegeIntegrationID, wi.ID)

		if err != nil {
			ctx.ReplyEmbed(utils.GenerateEmbed(*ctx, discordgo.MessageEmbed{
				Title:       "Error",
				Description: "An error occurred while fetching the dorm.",
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

		jsonStr, err := json.Marshal(data.Data)

		if err != nil {
			ctx.ReplyEmbed(utils.GenerateEmbed(*ctx, discordgo.MessageEmbed{
				Title:       "Error",
				Description: "An error occurred while fetching the dorm.",
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

		var dorm *types.CollegeIntegrationData

		err = json.Unmarshal(jsonStr, &dorm)

		if err != nil {
			ctx.ReplyEmbed(utils.GenerateEmbed(*ctx, discordgo.MessageEmbed{
				Title:       "Error",
				Description: "An error occurred while fetching the dorm.",
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

		fullUser, err := ctx.Session.GuildMember(ctx.Message.GuildID, user)

		if err != nil {
			ctx.ReplyEmbed(utils.GenerateEmbed(*ctx, discordgo.MessageEmbed{
				Title:       "Error",
				Description: "An error occurred while fetching the user.",
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
			Title: fmt.Sprintf("%s's Dorm is in %s %s", fullUser.User.Username, dorm.House, dorm.Room),
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:   "House",
					Value:  dorm.House,
					Inline: true,
				},
				{
					Name:   "Room",
					Value:  dorm.Room,
					Inline: true,
				},
			},
			Color: 0x00ff00,
		}))
	},
}
