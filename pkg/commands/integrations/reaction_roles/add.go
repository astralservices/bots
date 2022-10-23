package reactionroles

import (
	"encoding/json"

	db "github.com/astralservices/bots/pkg/database/supabase"
	"github.com/astralservices/bots/pkg/types"
	"github.com/astralservices/bots/pkg/utils"
	"github.com/astralservices/dgc"
	"github.com/bwmarrin/discordgo"
)

var AddReactionRole = &dgc.Command{
	Name:        "add",
	Aliases:     []string{"add", "create"},
	Description: "Add a reaction role to a message. You must be in the same channel as the message.",
	Usage:       "rr add <message_id> <emoji> <role_id> <dm [false]> <remove_reaction [true]>",
	Example:     "rr add 123456789012345678 🎉 123456789012345678 true true",
	Category:    "Reaction Roles",
	Domain:      "astral.integrations.reactionroles.add",
	Handler: func(ctx *dgc.Ctx) {
		database := db.New()

		messageID := ctx.Arguments.Get(0).Raw()
		emoji := ctx.Arguments.Get(1).Raw()
		roleID := ctx.Arguments.Get(2).Raw()

		if messageID == "" || emoji == "" || roleID == "" {
			ctx.ReplyEmbed(utils.GenerateEmbed(*ctx, discordgo.MessageEmbed{
				Title:       "Error",
				Description: "You must provide a message ID, emoji, and role ID",
				Color:       0xff0000,
			}))
			return
		}

		dm, err := ctx.Arguments.Get(3).AsBool()

		if err != nil {
			dm = false
		}

		removeReaction, err := ctx.Arguments.Get(4).AsBool()

		if err != nil {
			removeReaction = true
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

		roles := ReactionRolesData{
			ReactionRoles: []ReactionRole{
				{
					MessageID:      messageID,
					ChannelID:      ctx.Event.ChannelID,
					Emoji:          emoji,
					RoleID:         roleID,
					DM:             dm,
					RemoveReaction: removeReaction,
				},
			},
		}

		if len(workspaceData) == 0 {
			err = database.SetIntegrationDataForWorkspace(*self.Workspace, ReactionRolesIntegrationID, roles)

			if err != nil {
				ctx.ReplyEmbed(utils.GenerateEmbed(*ctx, discordgo.MessageEmbed{
					Title:       "Error",
					Description: "An error occurred while setting the integration data.",
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
				Description: "Successfully added the reaction role.",
				Color:       0x00ff00,
			}))
			return
		} else {
			jsonStr, err := json.Marshal(workspaceData[0].Data)

			if err != nil {
				ctx.ReplyEmbed(utils.GenerateEmbed(*ctx, discordgo.MessageEmbed{
					Title:       "Error",
					Description: "An error occurred while parsing the integration data.",
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

			var data ReactionRolesData

			err = json.Unmarshal(jsonStr, &data)

			if err != nil {
				ctx.ReplyEmbed(utils.GenerateEmbed(*ctx, discordgo.MessageEmbed{
					Title:       "Error",
					Description: "An error occurred while parsing the integration data.",
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

			err = database.SetIntegrationDataForWorkspace(*self.Workspace, ReactionRolesIntegrationID, ReactionRolesData{
				ReactionRoles: append(data.ReactionRoles, roles.ReactionRoles...),
			})

			if err != nil {
				ctx.ReplyEmbed(utils.GenerateEmbed(*ctx, discordgo.MessageEmbed{
					Title:       "Error",
					Description: "An error occurred while setting the integration data.",
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

			m, err := ctx.Session.ChannelMessage(ctx.Event.ChannelID, messageID)

			if err != nil {
				ctx.ReplyEmbed(utils.GenerateEmbed(*ctx, discordgo.MessageEmbed{
					Title:       "Error",
					Description: "An error occurred while fetching the message. Make sure you're in the same channel as the message.",
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

			hasReaction := false

			for _, reaction := range m.Reactions {
				if reaction.Emoji.APIName() == emoji || reaction.Emoji.Name == emoji {
					hasReaction = true
					break
				}
			}

			str := "Successfully added the reaction role."

			if !hasReaction {
				err = ctx.Session.MessageReactionAdd(ctx.Message.ChannelID, messageID, emoji)

				if err != nil {
					ctx.ReplyEmbed(utils.GenerateEmbed(*ctx, discordgo.MessageEmbed{
						Title:       "Error",
						Description: "An error occurred while adding the reaction.",
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

				str += "\n\nThe reaction has been added to the message."
			}

			ctx.ReplyEmbed(utils.GenerateEmbed(*ctx, discordgo.MessageEmbed{
				Title:       "Success",
				Description: str,
				Color:       0x00ff00,
			}))
		}
	},
}
