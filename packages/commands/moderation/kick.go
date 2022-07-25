package commands

import (
	"errors"
	"strings"

	db "github.com/astralservices/bots/supabase"
	"github.com/astralservices/bots/utils"
	"github.com/zekroTJA/shireikan"
)

type Kick struct {
}

// GetInvoke returns the command invokes.
func (c *Kick) GetInvokes() []string {
	return []string{"kick"}
}

// GetDescription returns the commands description.
func (c *Kick) GetDescription() string {
	return "Kick a user"
}

// GetHelp returns the commands help text.
func (c *Kick) GetHelp() string {
	return "`ban <user> <reason> (<duration>)` - Kick a user for a reason"
}

// GetGroup returns the commands group.
func (c *Kick) GetGroup() string {
	return utils.CategoryModeration
}

// GetDomainName returns the commands domain name.
func (c *Kick) GetDomainName() string {
	return "astral.moderation.kick"
}

// GetSubPermissionRules returns the commands sub
// permissions array.
func (c *Kick) GetSubPermissionRules() []shireikan.SubPermission {
	return nil
}

// IsExecutableInDMChannels returns whether
// the command is executable in DM channels.
func (c *Kick) IsExecutableInDMChannels() bool {
	return false
}

// Exec is the commands execution handler.
func (c *Kick) Exec(ctx shireikan.Context) error {
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
		_, err := utils.ReplyWithEmbed(ctx, utils.ErrorEmbed(ctx, errors.New("You can't kick yourself.")))
		return err
	}

	argsMsgs := ctx.GetArgs()[1:]

	if len(argsMsgs) < 1 {
		_, err := utils.ReplyWithEmbed(ctx, utils.ErrorEmbed(ctx, errors.New("Not enough arguments. Please provide a reason.")))
		return err
	}

	reason := strings.Join(argsMsgs, " ")

	action := utils.IBotModerationAction{
		Bot:       *bot.ID,
		Guild:     ctx.GetGuild().ID,
		Action:    "kick",
		Moderator: ctx.GetUser().ID,
		User:      victim.User.ID,
		Reason:    reason,
	}

	err = ctx.GetSession().GuildMemberDeleteWithReason(ctx.GetGuild().ID, victim.User.ID, reason)

	if err != nil {
		_, err := utils.ReplyWithEmbed(ctx, utils.ErrorEmbed(ctx, err))
		return err
	}

	database := db.New()

	err = database.DB.From("moderation_actions").Insert(&action).Execute(&action)

	if err != nil {
		_, err := utils.ReplyWithEmbed(ctx, utils.ErrorEmbed(ctx, errors.New("An error occured while saving the action to the database. The action has still been executed.")))
		return err
	}

	return nil
}
