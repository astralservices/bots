package utility

import (
	"fmt"
	"strings"

	"github.com/astralservices/bots/pkg/utils"
	"github.com/astralservices/dgc"
	"github.com/bwmarrin/discordgo"
)

var WhoisCommand = &dgc.Command{
	Name:        "whois",
	Aliases:     []string{"whois", "user", "userinfo"},
	Domain:      "astral.utility.whois",
	Category:    "Utility",
	Usage:       "whois <user>",
	Description: "View information about a user.",
	Example:     "whois @AmusedGrape",
	Handler: func(ctx *dgc.Ctx) {
		userId := ctx.Arguments.Get(0).AsUserMentionID()

		if userId == "" {
			userId = ctx.Arguments.Get(0).Raw()
		}

		if userId == "" {
			userId = ctx.Event.Author.ID
		}

		user, err := ctx.Session.User(userId)

		if err != nil {
			ctx.ReplyEmbed(utils.ErrorEmbed(*ctx, err))
			return
		}

		member, err := ctx.Session.GuildMember(ctx.Event.GuildID, userId)

		if err != nil {
			ctx.ReplyEmbed(utils.ErrorEmbed(*ctx, err))
			return
		}

		userJoined, err := discordgo.SnowflakeTimestamp(member.User.ID)

		if err != nil {
			ctx.ReplyEmbed(utils.ErrorEmbed(*ctx, err))
			return
		}

		var roleText []string

		for _, role := range member.Roles {
			roleText = append(roleText, fmt.Sprintf("<@&%s>", role))
		}

		var highestRole string

		if len(member.Roles) > 0 {
			highestRole = fmt.Sprintf("<@&%s>", member.Roles[0])
		} else {
			highestRole = "None"
		}

		var roles string

		if len(roleText) > 0 {
			roles = strings.Join(roleText, ", ")
		} else {
			roles = "None"
		}

		err = ctx.ReplyEmbed(utils.GenerateEmbed(*ctx, discordgo.MessageEmbed{
			Title: fmt.Sprintf("%s#%s's Information", user.Username, user.Discriminator),
			Thumbnail: &discordgo.MessageEmbedThumbnail{
				URL: user.AvatarURL(""),
			},
			Fields: []*discordgo.MessageEmbedField{
				{Name: "ID", Value: user.ID, Inline: true},
				{Name: "Joined", Value: fmt.Sprintf("<t:%d>", member.JoinedAt.Unix()), Inline: true},
				{Name: "Created", Value: fmt.Sprintf("<t:%d>", userJoined.Unix()), Inline: true},
				{Name: "Roles", Value: roles, Inline: true},
				{Name: "Highest Role", Value: highestRole, Inline: true},
				{Name: "Bot", Value: fmt.Sprintf("%t", user.Bot), Inline: true},
				{Name: "Discord Tag", Value: member.Mention(), Inline: true},
				{Name: "Discord Username", Value: user.Username, Inline: true},
				{Name: "Discord Discriminator", Value: user.Discriminator, Inline: true},
			},
		}))
	},
}
