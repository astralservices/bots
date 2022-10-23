package reactionroles

import (
	"encoding/json"
	"fmt"

	db "github.com/astralservices/bots/pkg/database/supabase"
	"github.com/astralservices/bots/pkg/types"
	"github.com/astralservices/bots/pkg/utils"
	"github.com/astralservices/dgc"
	"github.com/bwmarrin/discordgo"
)

var ListReactionRoles = &dgc.Command{
	Name:        "list",
	Aliases:     []string{"list", "ls"},
	Description: "List all reaction roles for a message",
	Usage:       "rr list <message_id>",
	Example:     "rr list 123456789012345678",
	Category:    "Reaction Roles",
	Domain:      "astral.integrations.reactionroles.list",
	Handler: func(ctx *dgc.Ctx) {
		database := db.New()

		messageID := ctx.Arguments.Get(0).Raw()

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

		var reactionRoles []ReactionRole

		for _, reactionRole := range reactionRolesData.ReactionRoles {
			if reactionRole.MessageID == messageID && messageID != "" {
				reactionRoles = append(reactionRoles, reactionRole)
			}
		}

		if len(reactionRoles) == 0 && messageID != "" {
			ctx.ReplyEmbed(utils.GenerateEmbed(*ctx, discordgo.MessageEmbed{
				Title:       "Error",
				Description: "No reaction roles have been added to this message.",
				Color:       0xff0000,
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

		var fields []*discordgo.MessageEmbedField
		var description string

		if messageID == "" {
			description = "All reaction roles for this server."
			var fs map[string]*discordgo.MessageEmbedField
			for _, reactionRole := range reactionRolesData.ReactionRoles {
				if fs == nil {
					fs = make(map[string]*discordgo.MessageEmbedField)
				}

				if fs[reactionRole.MessageID] == nil {
					fs[reactionRole.MessageID] = &discordgo.MessageEmbedField{
						Name:   fmt.Sprintf("%s", reactionRole.MessageID),
						Value:  fmt.Sprintf("%s - <@&%s>", reactionRole.Emoji, reactionRole.RoleID),
						Inline: false,
					}
				} else {
					fs[reactionRole.MessageID].Value += fmt.Sprintf("\n%s - <@&%s>", reactionRole.Emoji, reactionRole.RoleID)
				}

				fs[reactionRole.MessageID].Value += fmt.Sprintf("\n\n[Visit Message](https://discord.com/channels/%s/%s/%s)", ctx.Event.GuildID, reactionRole.ChannelID, reactionRole.MessageID)
			}

			for _, f := range fs {
				fields = append(fields, f)
			}
		} else {
			description = fmt.Sprintf("Reaction roles for message [%s](https://discord.com/channels/%s/%s/%s)", messageID, ctx.Event.GuildID, reactionRoles[0].ChannelID, messageID)
			for _, reactionRole := range reactionRoles {
				fields = append(fields, &discordgo.MessageEmbedField{
					Name:   reactionRole.Emoji,
					Value:  fmt.Sprintf("<@&%s>", reactionRole.RoleID),
					Inline: false,
				})
			}
		}

		ctx.ReplyEmbed(utils.GenerateEmbed(*ctx, discordgo.MessageEmbed{
			Title:       "Reaction Roles",
			Description: description,
			Color:       0x00ff00,
			Fields:      fields,
		}))
	},
}
