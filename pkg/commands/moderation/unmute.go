package moderation

import (
	"fmt"
	"os"
	"strings"

	"github.com/astralservices/bots/pkg/utils"
	"github.com/astralservices/dgc"
	"github.com/bwmarrin/discordgo"
)

var UnmuteCommand = &dgc.Command{
	Name:        "unmute",
	Domain:      "astral.moderation.unmute",
	Aliases:     []string{"unmute"},
	Category:    "Moderation",
	Usage:       "unmute <user>",
	Example:     "unmute @AmusedGrape",
	Description: "Unmute a user from the server.",
	Slash:       true,
	SlashGuilds: []string{os.Getenv("DEV_GUILD")},
	Handler: func(ctx *dgc.Ctx) {
		// find the muted role by name
		// remove the role from the user
		// if the user is muted through discord, remove the timeout

		guild, err := ctx.Session.Guild(ctx.Event.GuildID)

		if err != nil {
			ctx.ReplyEmbed(utils.ErrorEmbed(*ctx, err))
			return
		}

		// find the muted role
		var mutedRole *discordgo.Role

		for _, role := range guild.Roles {
			if strings.Contains(strings.ToLower(role.Name), "muted") {
				mutedRole = role
				break
			}
		}

		// get the user
		userId := ctx.Arguments.Get(0).AsUserMentionID()

		if userId == "" {
			userId = ctx.Arguments.Get(0).Raw()
		}

		victim, err := ctx.Session.GuildMember(ctx.Event.GuildID, userId)

		if err != nil {
			ctx.ReplyEmbed(utils.ErrorEmbed(*ctx, err))
			return
		}

		var wasMuted bool

		// remove the role, if the user has it
		for _, role := range victim.Roles {
			if role == mutedRole.ID {
				err := ctx.Session.GuildMemberRoleRemove(ctx.Event.GuildID, victim.User.ID, mutedRole.ID)

				if err != nil {
					ctx.ReplyEmbed(utils.ErrorEmbed(*ctx, err))
					return
				}

				wasMuted = true
			}
		}

		// remove the timeout, if the user has one
		if victim.CommunicationDisabledUntil != nil {
			err := ctx.Session.GuildMemberTimeout(ctx.Event.GuildID, victim.User.ID, nil)

			if err != nil {
				ctx.ReplyEmbed(utils.ErrorEmbed(*ctx, err))
				return
			}

			wasMuted = true
		}

		if !wasMuted {
			ctx.ReplyEmbed(utils.ErrorEmbed(*ctx, fmt.Errorf("%s is not muted.", victim.User.Mention())))
			return
		}

		ctx.ReplyEmbed(utils.GenerateEmbed(*ctx, discordgo.MessageEmbed{
			Title:       "Unmuted",
			Description: fmt.Sprintf("%s has been unmuted by %s", victim.User.Mention(), ctx.Event.Author.Mention()),
			Color:       0x00ff00,
		}))
	},
}
