package reactionroles

import (
	"encoding/json"
	"fmt"

	db "github.com/astralservices/bots/pkg/database/supabase"
	"github.com/bwmarrin/discordgo"
)

var ReactionRolesIntegrationID = "827bc5bb-4be4-4cc4-9ef6-db0c2546662f"

type ReactionRole struct {
	MessageID      string `json:"message_id"`
	ChannelID      string `json:"channel_id"`
	Emoji          string `json:"emoji"`
	RoleID         string `json:"role_id"`
	DM             bool   `json:"dm"`
	RemoveReaction bool   `json:"remove_reaction"` // Removes if the user already has the role and
}

type ReactionRolesData struct {
	ReactionRoles []ReactionRole `json:"reaction_roles"`
}

func HandleReactionRolesAdd(workspace string) func(s *discordgo.Session, m *discordgo.MessageReactionAdd) {
	return func(s *discordgo.Session, m *discordgo.MessageReactionAdd) {
		user, err := s.GuildMember(m.GuildID, m.UserID)

		if err != nil {
			return
		}

		database := db.New()

		data, err := database.GetIntegrationDataForWorkspace(workspace, ReactionRolesIntegrationID)

		if err != nil {
			return
		}

		jsonStr, err := json.Marshal(data[0].Data)

		if err != nil {
			return
		}

		var reactionRolesData ReactionRolesData

		err = json.Unmarshal(jsonStr, &reactionRolesData)

		if err != nil {
			return
		}

		for _, role := range reactionRolesData.ReactionRoles {
			r, err := s.State.Role(m.GuildID, role.RoleID)

			if err != nil {
				return
			}

			if role.MessageID == m.MessageID {
				if m.UserID == s.State.User.ID {
					return
				}

				if role.Emoji == m.Emoji.APIName() || role.Emoji == m.Emoji.Name {
					err = s.GuildMemberRoleAdd(m.GuildID, m.UserID, role.RoleID)

					if err != nil {
						return
					}

					if role.DM {
						s.ChannelMessageSendEmbed(user.User.ID, &discordgo.MessageEmbed{
							Title:       "Reaction Roles",
							Description: "You have been given the role " + r.Name,
							Color:       0x00ff00,
						})
					}
				}
			}
		}
	}
}

func HandleReactionRolesRemove(workspace string) func(s *discordgo.Session, m *discordgo.MessageReactionRemove) {
	return func(s *discordgo.Session, m *discordgo.MessageReactionRemove) {
		user, err := s.GuildMember(m.GuildID, m.UserID)

		if err != nil {
			return
		}

		database := db.New()

		data, err := database.GetIntegrationDataForWorkspace(workspace, ReactionRolesIntegrationID)

		if err != nil {
			return
		}

		jsonStr, err := json.Marshal(data[0].Data)

		if err != nil {
			return
		}

		var reactionRolesData ReactionRolesData

		err = json.Unmarshal(jsonStr, &reactionRolesData)

		if err != nil {
			return
		}
		for _, role := range reactionRolesData.ReactionRoles {
			r, err := s.State.Role(m.GuildID, role.RoleID)

			if err != nil {
				return
			}

			if role.MessageID == m.MessageID {
				if role.Emoji == m.Emoji.APIName() || role.Emoji == m.Emoji.Name {
					for _, userRole := range user.Roles {
						if userRole == role.RoleID {
							if role.RemoveReaction {
								s.GuildMemberRoleRemove(m.GuildID, m.UserID, role.RoleID)

								if role.DM {
									s.ChannelMessageSendEmbed(m.ChannelID, &discordgo.MessageEmbed{
										Title:       "Reaction Role Removed",
										Description: fmt.Sprintf("You have been removed from the role **%s**", r.Name),
										Color:       0xff0000,
									})
								}

								return
							}
						}
					}
				}
			}
		}
	}
}
