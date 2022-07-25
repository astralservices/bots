package commands

import (
	"github.com/astralservices/bots/utils"
	"github.com/bwmarrin/discordgo"
	"github.com/zekroTJA/shireikan"
)

type Cat struct {
}

// GetInvoke returns the command invokes.
func (c *Cat) GetInvokes() []string {
	return []string{"cat", "pussy"}
}

// GetDescription returns the commands description.
func (c *Cat) GetDescription() string {
	return "Gets a random cat picture from Reddit"
}

// GetHelp returns the commands help text.
func (c *Cat) GetHelp() string {
	return "`Cat` - Cat"
}

// GetGroup returns the commands group.
func (c *Cat) GetGroup() string {
	return utils.CategoryFun
}

// GetDomainName returns the commands domain name.
func (c *Cat) GetDomainName() string {
	return "astral.fun.cat"
}

// GetSubPermissionRules returns the commands sub
// permissions array.
func (c *Cat) GetSubPermissionRules() []shireikan.SubPermission {
	return nil
}

// IsExecutableInDMChannels returns whether
// the command is executable in DM channels.
func (c *Cat) IsExecutableInDMChannels() bool {
	return false
}

// Exec is the commands execution handler.
func (c *Cat) Exec(ctx shireikan.Context) error {
	// bot := ctx.GetObject("bot").(utils.IBot)

	// code

	subreddit := utils.Subreddit{
		Name: "catpictures",
	}

	post, err := subreddit.RandomHot()

	if err != nil {
		return utils.ErrorHandler(err)
	}

	embed := utils.GenerateEmbed(ctx, discordgo.MessageEmbed{
		Title: post.Title,
	})

	_, err = utils.ReplyWithEmbed(ctx, embed)

	return err
}
