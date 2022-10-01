package reactionroles

import (
	"encoding/json"

	db "github.com/astralservices/bots/pkg/database/supabase"
	"github.com/astralservices/bots/pkg/types"
	"github.com/astralservices/bots/pkg/utils"
	"github.com/astralservices/dgc"
	"github.com/bwmarrin/discordgo"
)

var RemoveReactionRole = &dgc.Command{
	Name:        "rr-remove",
	Aliases:     []string{"rr-remove", "rr-rm"},
	Description: "Remove a reaction role from a message",
	Usage:       "rr-remove <message_id> <emoji>",
	Example:     "rr-remove 123456789012345678 🎉",
	Category:    "Reaction Roles",
	Domain:      "astral.integrations.reactionroles-rm",
	Handler: func(ctx *dgc.Ctx) {
		database := db.New()

		messageID := ctx.Arguments.Get(0).Raw()
		emoji := ctx.Arguments.Get(1).Raw()

		if messageID == "" || emoji == "" {
			ctx.ReplyEmbed(utils.GenerateEmbed(*ctx, discordgo.MessageEmbed{
				Title:       "Error",
				Description: "You must provide a message ID and emoji",
				Color:       0xff0000,
			}))
			return
		}

		self := ctx.CustomObjects.MustGet("self").(types.Bot)

		workspaceData, err := database.GetIntegrationDataForWorkspace(*self.Workspace, ReactionRolesIntegrationID)

		if err != nil {
			ctx.ReplyEmbed(utils.GenerateEmbed(*ctx, discordgo.MessageEmbed{
				Title:       "Error",
				Description: "An error occurred while fetching the integration data.",
				Color:       0xff0000,
				Fields: []*discordgo.MessageEmbedField{
					{
						Name:   "Error",
						Value:  err.Error(),
						Inline: false,
					},
				},
			}))
			return
		}

		if len(workspaceData) == 0 {
			ctx.ReplyEmbed(utils.GenerateEmbed(*ctx, discordgo.MessageEmbed{
				Title:       "Error",
				Description: "No reaction roles have been added to this server.",
				Color:       0xff0000,
			}))
			return
		}

		jsonStr, err := json.Marshal(workspaceData[0].Data)

		if err != nil {
			ctx.ReplyEmbed(utils.GenerateEmbed(*ctx, discordgo.MessageEmbed{
				Title:       "Error",
				Description: "An error occurred while fetching the integration data.",
				Color:       0xff0000,
				Fields: []*discordgo.MessageEmbedField{
					{
						Name:   "Error",
						Value:  err.Error(),
						Inline: false,
					},
				},
			}))
			return
		}

		var reactionRolesData ReactionRolesData

		err = json.Unmarshal(jsonStr, &reactionRolesData)

		if err != nil {
			ctx.ReplyEmbed(utils.GenerateEmbed(*ctx, discordgo.MessageEmbed{
				Title:       "Error",
				Description: "An error occurred while fetching the integration data.",
				Color:       0xff0000,
				Fields: []*discordgo.MessageEmbedField{
					{
						Name:   "Error",
						Value:  err.Error(),
						Inline: false,
					},
				},
			}))
			return
		}

		if len(reactionRolesData.ReactionRoles) == 0 {
			ctx.ReplyEmbed(utils.GenerateEmbed(*ctx, discordgo.MessageEmbed{
				Title:       "Error",
				Description: "No reaction roles have been added to this server.",
				Color:       0xff0000,
			}))
			return
		}

		var reactionRole ReactionRole

		for _, rr := range reactionRolesData.ReactionRoles {
			if rr.MessageID == messageID && rr.Emoji == emoji {
				reactionRole = rr
			}
		}

		if reactionRole.MessageID == "" {
			ctx.ReplyEmbed(utils.GenerateEmbed(*ctx, discordgo.MessageEmbed{
				Title:       "Error",
				Description: "No reaction role found for the provided message ID and emoji.",
				Color:       0xff0000,
			}))
			return
		}

		for i, rr := range reactionRolesData.ReactionRoles {
			if rr.MessageID == messageID && rr.Emoji == emoji {
				reactionRolesData.ReactionRoles = append(reactionRolesData.ReactionRoles[:i], reactionRolesData.ReactionRoles[i+1:]...)
			}
		}

		err = database.SetIntegrationDataForWorkspace(*self.Workspace, ReactionRolesIntegrationID, reactionRolesData)

		if err != nil {
			ctx.ReplyEmbed(utils.GenerateEmbed(*ctx, discordgo.MessageEmbed{
				Title:       "Error",
				Description: "An error occurred while saving the integration data.",
				Color:       0xff0000,
				Fields: []*discordgo.MessageEmbedField{
					{
						Name:   "Error",
						Value:  err.Error(),
						Inline: false,
					},
				},
			}))
			return
		}

		ctx.ReplyEmbed(utils.GenerateEmbed(*ctx, discordgo.MessageEmbed{
			Title:       "Success",
			Description: "Successfully removed the reaction role.",
			Color:       0x00ff00,
		}))
	},
}
