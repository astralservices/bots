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

var MuteCommand = &dgc.Command{
	Name:        "mute",
	Domain:      "astral.moderation.mute",
	Aliases:     []string{"mute"},
	Category:    "Moderation",
	Usage:       "mute <user> [length] [reason]",
	Example:     "mute @AmusedGrape 1d beans are good",
	Description: "Mute a user from the server.",
	Slash:       true,
	SlashGuilds: []string{os.Getenv("DEV_GUILD")},
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
			ctx.ReplyEmbed(utils.ErrorEmbed(*ctx, fmt.Errorf("You cannot mute yourself.")))
			return
		}

		if victim.User.ID == ctx.Session.State.User.ID {
			ctx.ReplyEmbed(utils.ErrorEmbed(*ctx, fmt.Errorf("You cannot mute me.")))
			return
		}

		if victim.User.Bot {
			ctx.ReplyEmbed(utils.ErrorEmbed(*ctx, fmt.Errorf("You cannot mute bots.")))
			return
		}

		report, err := database.AddReport(types.Report{
			Bot:       *self.ID,
			Moderator: ctx.Message.Author.ID,
			User:      victim.User.ID,
			Guild:     ctx.Event.GuildID,
			Reason:    reason,
			Action:    "mute",
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

		// send a message to the user then mute them
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
			Title: fmt.Sprintf("🤬 Muted from %s", guild.Name),
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

		// if the timeout is not longer than 28 days we can use the discord mute
		if timeout.Seconds() > 0 && timeout.Hours() <= 672 {
			discordTimeout := utils.NowAddPtr(timeout)

			err = ctx.Session.GuildMemberTimeout(ctx.Event.GuildID, victim.User.ID, &discordTimeout)

			if err != nil {
				ctx.ReplyEmbed(utils.ErrorEmbed(*ctx, err))
				return
			}
		}

		// if the timeout is longer than 28 days, or no duration was provided, we need to use our own mute
		if timeout.Hours() > 672 || timeout.Seconds() == 0 {
			// first check if the role exists with the name "Muted"

			g, err := ctx.Session.State.Guild(ctx.Event.GuildID)

			if err != nil {
				ctx.ReplyEmbed(utils.ErrorEmbed(*ctx, err))
				return
			}

			roles := g.Roles

			if err != nil {
				ctx.ReplyEmbed(utils.ErrorEmbed(*ctx, err))
				return
			}

			var mutedRole *discordgo.Role

			for _, role := range roles {
				if strings.Contains(strings.ToLower(role.Name), "muted") {
					mutedRole = role
					break
				}
			}

			// if the role does not exist, create it

			if mutedRole == nil {
				// create a new role for the mute
				mutedRole, err = ctx.Session.GuildRoleCreate(ctx.Event.GuildID)

				if err != nil {
					ctx.ReplyEmbed(utils.ErrorEmbed(*ctx, err))
					return
				}

				// set the role permissions to 0, color to black, and name to "Muted"
				mutedRole, err = ctx.Session.GuildRoleEdit(ctx.Event.GuildID, mutedRole.ID, "Muted", 0x000000, false, 0, false)

				if err != nil {
					ctx.ReplyEmbed(utils.ErrorEmbed(*ctx, err))
					return
				}

				// edit every category to deny the mute role send messages
				categories, err := ctx.Session.GuildChannels(ctx.Event.GuildID)

				if err != nil {
					ctx.ReplyEmbed(utils.ErrorEmbed(*ctx, err))
					return
				}

				for _, category := range categories {
					if category.Type == discordgo.ChannelTypeGuildCategory {
						err = ctx.Session.ChannelPermissionSet(category.ID, mutedRole.ID, discordgo.PermissionOverwriteTypeRole, 0, discordgo.PermissionSendMessages)

						if err != nil {
							ctx.ReplyEmbed(utils.ErrorEmbed(*ctx, err))
							return
						}
					}
				}

			}

			// add the role to the user

			err = ctx.Session.GuildMemberRoleAdd(ctx.Event.GuildID, victim.User.ID, mutedRole.ID)

			if err != nil {
				ctx.ReplyEmbed(utils.ErrorEmbed(*ctx, err))
				return
			}
		}

		var additionalFields []*discordgo.MessageEmbedField

		if err != nil {
			additionalFields = append(additionalFields, &discordgo.MessageEmbedField{
				Name:  "Error",
				Value: err.Error(),
			})
		}

		ctx.ReplyEmbed(utils.GenerateEmbed(*ctx, discordgo.MessageEmbed{
			Title:       "The Duct Tape has spoken! 🦆",
			Description: fmt.Sprintf("%s#%s has been muted.", victim.User.Username, victim.User.Discriminator),
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
					Name: "Muted At",
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
