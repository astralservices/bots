package commands

import (
	"errors"
	"strings"
	"time"

	db "github.com/astralservices/bots/supabase"
	"github.com/astralservices/bots/utils"
	"github.com/bwmarrin/discordgo"
	"github.com/zekroTJA/shireikan"
)

type Ban struct {
}

// GetInvoke returns the command invokes.
func (c *Ban) GetInvokes() []string {
	return []string{"Ban"}
}

// GetDescription returns the commands description.
func (c *Ban) GetDescription() string {
	return "Ban a user"
}

// GetHelp returns the commands help text.
func (c *Ban) GetHelp() string {
	return "`ban <user> <reason> (<duration>)` - Ban a user for a reason with an optional duration"
}

// GetGroup returns the commands group.
func (c *Ban) GetGroup() string {
	return utils.CategoryModeration
}

// GetDomainName returns the commands domain name.
func (c *Ban) GetDomainName() string {
	return "astral.moderation.ban"
}

// GetSubPermissionRules returns the commands sub
// permissions array.
func (c *Ban) GetSubPermissionRules() []shireikan.SubPermission {
	return nil
}

// IsExecutableInDMChannels returns whether
// the command is executable in DM channels.
func (c *Ban) IsExecutableInDMChannels() bool {
	return false
}

// Exec is the commands execution handler.
func (c *Ban) Exec(ctx shireikan.Context) error {
	bot := ctx.GetObject("bot").(utils.IBot)

	if len(ctx.GetArgs()) < 2 {
		_, err := utils.ReplyWithEmbed(ctx, utils.ErrorEmbed(ctx, errors.New("Not enough arguments. Please provide a user and a reason.")))
		return err
	}

	victim, err := utils.FetchMember(ctx.GetSession(), ctx.GetGuild().ID, ctx.GetArgs().Get(0).AsString())
	if err != nil || victim == nil {
		_, err := utils.ReplyWithEmbed(ctx, utils.ErrorEmbed(ctx, errors.New("User not found.")))
		return err
	}

	if victim.User.ID == ctx.GetUser().ID {
		_, err := utils.ReplyWithEmbed(ctx, utils.ErrorEmbed(ctx, errors.New("You can't ban yourself.")))
		return err
	}

	argsMsgs := ctx.GetArgs()[1:]

	timeout, err := time.ParseDuration(argsMsgs[len(argsMsgs)-1])
	if err == nil && timeout > 0 {
		argsMsgs = argsMsgs[:len(argsMsgs)-1]
	}

	if len(argsMsgs) < 1 {
		_, err := utils.ReplyWithEmbed(ctx, utils.ErrorEmbed(ctx, errors.New("Not enough arguments. Please provide a reason.")))
		return err
	}

	reason := strings.Join(argsMsgs, " ")

	action := utils.IBotModerationAction{
		Bot:       *bot.ID,
		Guild:     ctx.GetGuild().ID,
		Action:    "ban",
		Moderator: ctx.GetUser().ID,
		User:      victim.User.ID,
		Reason:    reason,
		Expiry:    ctx.GetMessage().Timestamp.Add(timeout),
		Expires:   timeout > 0,
	}

	err = ctx.GetSession().GuildBanCreateWithReason(ctx.GetGuild().ID, victim.User.ID, reason, 0)

	if err != nil {
		_, err := utils.ReplyWithEmbed(ctx, utils.ErrorEmbed(ctx, err))
		return err
	}

	database := db.New()

	err = database.DB.From("moderation_actions").Insert(&action).Execute(&action)

	if err != nil {
		_, err := utils.ReplyWithEmbed(ctx, utils.ErrorEmbed(ctx, errors.New("An error occured while saving the action to the database. The action has still been executed.\nError: `"+err.Error()+"`")))
		return err
	}

	var expiresString string

	if action.Expires {
		expiresString = action.Expiry.Format(time.RFC3339)
	} else {
		expiresString = "Permament"
	}

	_, err = utils.ReplyWithEmbed(ctx, utils.GenerateEmbed(ctx, discordgo.MessageEmbed{
		Title:       ":hammer: Banned " + victim.User.Username + "#" + victim.User.Discriminator,
		Description: "User " + victim.User.Username + "#" + victim.User.Discriminator + " has been banned",
		Color:       0xFF0000,
		Fields: []*discordgo.MessageEmbedField{
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
				Value: ctx.GetUser().Username + "#" + ctx.GetUser().Discriminator,
			},
			{
				Name:  "Banned At",
				Value: ctx.GetMessage().Timestamp.Format(time.RFC3339),
			},
			{
				Name:  "Expires",
				Value: expiresString,
			},
			{
				Name:  "Case ID",
				Value: *action.ID,
			},
		},
	}))

	return err
}
