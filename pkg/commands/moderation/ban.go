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

var BanCommand = &dgc.Command{
	Name:          "ban",
	Domain:        "astral.moderation.ban",
	Aliases:       []string{"ban"},
	Category:      "Moderation",
	Usage:         "ban <user> [length] [reason]",
	Example:       "ban @AmusedGrape 1d beans are good",
	Description:   "Ban a user from the server.",
	Slash:         true,
	SlashGuilds:   []string{os.Getenv("DEV_GUILD")},
	IntegrationID: "",
	Arguments: []*discordgo.ApplicationCommandOption{
		{
			Name:        "user",
			Description: "The user to ban.",
			Type:        discordgo.ApplicationCommandOptionUser,
			Required:    true,
		},
		{
			Name:        "reason",
			Description: "The reason for the ban.",
			Type:        discordgo.ApplicationCommandOptionString,
			Required:    false,
		},
		{
			Name:        "expiry",
			Description: "The amount of time the ban should last.",
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

		timeout, err := utils.ParseDuration(ctx.Arguments.Get(1).Raw())

		if err != nil {
			timeout = 0
		}

		var remArgs []*dgc.Argument

		if timeout > 0 {
			remArgs = ctx.Arguments.GetAll()[2:]
		} else {
			timeout = 0
			remArgs = ctx.Arguments.GetAll()[1:]
		}

		if len(remArgs) < 1 {
			ctx.ReplyEmbed(utils.ErrorEmbed(*ctx, fmt.Errorf("Invalid arguments. Please provide a reason.")))
			return
		}

		var strArgs []string

		for _, arg := range remArgs {
			strArgs = append(strArgs, arg.Raw())
		}

		reason := strings.Join(strArgs, " ")

		if victim.User.ID == ctx.Event.Author.ID {
			ctx.ReplyEmbed(utils.ErrorEmbed(*ctx, fmt.Errorf("You cannot ban yourself.")))
			return
		}

		if victim.User.ID == ctx.Session.State.User.ID {
			ctx.ReplyEmbed(utils.ErrorEmbed(*ctx, fmt.Errorf("You cannot ban me.")))
			return
		}

		if victim.User.Bot {
			ctx.ReplyEmbed(utils.ErrorEmbed(*ctx, fmt.Errorf("You cannot ban bots.")))
			return
		}

		report, err := database.AddReport(types.Report{
			Bot:       *self.ID,
			Moderator: ctx.Message.Author.ID,
			User:      victim.User.ID,
			Guild:     ctx.Event.GuildID,
			Reason:    reason,
			Action:    "ban",
			Expiry:    utils.NowAddPtr(timeout),
			Expires:   timeout > 0,
		})

		if err != nil {
			ctx.ReplyEmbed(utils.ErrorEmbed(*ctx, err))
			return
		}

		expiresValue := "Never"

		if timeout > 0 {
			expiresValue = fmt.Sprintf("<t:%d>", report.Expiry.Unix())
		}

		// send a message to the user then ban them
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
			Title: fmt.Sprintf(":hammer: Banned from %s", guild.Name),
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
					Name:  "Expires",
					Value: expiresValue,
				},
				{
					Name:  "Case ID",
					Value: fmt.Sprintf("`%s`", *report.ID),
				},
			},
		})

		err = ctx.Session.GuildBanCreateWithReason(ctx.Event.GuildID, victim.User.ID, reason, 0)

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
			Title:       "The Ban Hammer has spoken! 🔨",
			Description: fmt.Sprintf("%s#%s has been banned.", victim.User.Username, victim.User.Discriminator),
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
					Name: "Banned At",
					// use discord's timestamp formatting (<t:unix_timestamp>)
					Value: fmt.Sprintf("<t:%d>", time.Now().Unix()),
				},
				{
					Name:  "Expires",
					Value: expiresValue,
				},
				{
					Name:  "Case ID",
					Value: fmt.Sprintf("`%s`", *report.ID),
				},
			}, additionalFields...),
		}))
	},
}
