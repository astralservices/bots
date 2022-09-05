package moderation

import (
	"fmt"
	"os"
	"strings"
	"time"

	db "github.com/astralservices/bots/pkg/database/supabase"
	"github.com/astralservices/bots/pkg/types"
	"github.com/astralservices/bots/pkg/utils"
	"github.com/astralservices/dgc"
	"github.com/bwmarrin/discordgo"
)

var KickCommand = &dgc.Command{
	Name:          "kick",
	Domain:        "astral.moderation.kick",
	Aliases:       []string{"kick"},
	Category:      "Moderation",
	Usage:         "kick <user> [length] [reason]",
	Example:       "kick @AmusedGrape 1d beans are good",
	Description:   "kick a user from the server.",
	Slash:         true,
	SlashGuilds:   []string{os.Getenv("DEV_GUILD")},
	IntegrationID: "",
	Arguments: []*discordgo.ApplicationCommandOption{
		{
			Name:        "user",
			Description: "The user to kick.",
			Type:        discordgo.ApplicationCommandOptionUser,
			Required:    true,
		},
		{
			Name:        "reason",
			Description: "The reason for the kick.",
			Type:        discordgo.ApplicationCommandOptionString,
			Required:    false,
		},
	},
	Handler: func(ctx *dgc.Ctx) {
		database := db.New()

		self := ctx.CustomObjects.MustGet("self").(types.Bot)

		if ctx.Arguments.Amount() < 2 {
			ctx.ReplyEmbed(utils.ErrorEmbed(*ctx, fmt.Errorf("Invalid arguments. Please provide a user and a reason.")))
			return
		}

		userId := ctx.Arguments.Get(0).AsUserMentionID()

		if userId == "" {
			userId = ctx.Arguments.Get(0).Raw()
		}

		victim, err := ctx.Session.GuildMember(ctx.Event.GuildID, userId)

		if err != nil {
			ctx.ReplyEmbed(utils.ErrorEmbed(*ctx, err))
			return
		}

		args := ctx.Arguments.GetAll()[1:]

		var argsStr []string

		for _, arg := range args {
			argsStr = append(argsStr, arg.Raw())
		}

		reason := strings.Join(argsStr, " ")

		if victim.User.ID == ctx.Event.Author.ID {
			ctx.ReplyEmbed(utils.ErrorEmbed(*ctx, fmt.Errorf("You cannot kick yourself.")))
			return
		}

		if victim.User.ID == ctx.Session.State.User.ID {
			ctx.ReplyEmbed(utils.ErrorEmbed(*ctx, fmt.Errorf("You cannot kick me.")))
			return
		}

		if victim.User.Bot {
			ctx.ReplyEmbed(utils.ErrorEmbed(*ctx, fmt.Errorf("You cannot kick bots.")))
			return
		}

		report, err := database.AddReport(types.Report{
			Bot:       *self.ID,
			Moderator: ctx.Message.Author.ID,
			User:      victim.User.ID,
			Guild:     ctx.Event.GuildID,
			Reason:    reason,
			Action:    "kick",
		})

		if err != nil {
			ctx.ReplyEmbed(utils.ErrorEmbed(*ctx, err))
			return
		}

		// send a message to the user then kick them
		userChannel, err := ctx.Session.UserChannelCreate(victim.User.ID)

		if err != nil {
			ctx.ReplyEmbed(utils.ErrorEmbed(*ctx, err))
			return
		}

		guild, err := ctx.Session.Guild(ctx.Message.GuildID)

		if err != nil {
			ctx.ReplyEmbed(utils.ErrorEmbed(*ctx, err))
			return
		}

		_, err = ctx.Session.ChannelMessageSendEmbed(userChannel.ID, &discordgo.MessageEmbed{
			Title: fmt.Sprintf(":hammer: kickned from %s", guild.Name),
			Color: 0xff0000,
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:  "Reason",
					Value: reason,
				},
				{
					Name:  "Moderator",
					Value: fmt.Sprintf("%s#%s", ctx.Message.Author.Username, ctx.Message.Author.Discriminator),
				},
				{
					Name:  "Case ID",
					Value: fmt.Sprintf("`%s`", *report.ID),
				},
			},
		})

		err = ctx.Session.GuildMemberDeleteWithReason(ctx.Event.GuildID, victim.User.ID, reason)

		if err != nil {
			ctx.ReplyEmbed(utils.ErrorEmbed(*ctx, err))
			return
		}

		var additionalFields []*discordgo.MessageEmbedField

		if err != nil {
			additionalFields = append(additionalFields, &discordgo.MessageEmbedField{
				Name:  "Error",
				Value: err.Error(),
			})
		}

		ctx.ReplyEmbed(utils.GenerateEmbed(*ctx, discordgo.MessageEmbed{
			Title:       "The Boot has spoken! :boot:",
			Description: fmt.Sprintf("%s#%s has been kicked.", victim.User.Username, victim.User.Discriminator),
			Color:       0x00ff00,
			Fields: append([]*discordgo.MessageEmbedField{
				{
					Name:  "Reason",
					Value: reason,
				},
				{
					Name:  "ID",
					Value: victim.User.ID,
				},
				{
					Name:  "Moderator",
					Value: fmt.Sprintf("%s#%s", ctx.Message.Author.Username, ctx.Message.Author.Discriminator),
				},
				{
					Name: "Kicked At",
					// use discord's timestamp formatting (<t:unix_timestamp>)
					Value: fmt.Sprintf("<t:%d>", time.Now().Unix()),
				},
				{
					Name:  "Case ID",
					Value: fmt.Sprintf("`%s`", *report.ID),
				},
			}, additionalFields...),
		}))
	},
}
