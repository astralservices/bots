package commands

import (
	"errors"
	"math/rand"
	"strings"
	"time"

	"github.com/astralservices/bots/utils"
	"github.com/bwmarrin/discordgo"
	"github.com/zekroTJA/shireikan"
)

type Eightball struct {
}

// GetInvoke returns the command invokes.
func (c *Eightball) GetInvokes() []string {
	return []string{"eightball", "8ball"}
}

// GetDescription returns the commands description.
func (c *Eightball) GetDescription() string {
	return "Ask a question and get an answer"
}

// GetHelp returns the commands help text.
func (c *Eightball) GetHelp() string {
	return "`eightball <question>` - Eightball"
}

// GetGroup returns the commands group.
func (c *Eightball) GetGroup() string {
	return utils.CategoryFun
}

// GetDomainName returns the commands domain name.
func (c *Eightball) GetDomainName() string {
	return "astral.fun.eightball"
}

// GetSubPermissionRules returns the commands sub
// permissions array.
func (c *Eightball) GetSubPermissionRules() []shireikan.SubPermission {
	return nil
}

// IsExecutableInDMChannels returns whether
// the command is executable in DM channels.
func (c *Eightball) IsExecutableInDMChannels() bool {
	return false
}

// Exec is the commands execution handler.
func (c *Eightball) Exec(ctx shireikan.Context) error {
	bot := ctx.GetObject("bot").(utils.IBot)

	// code

	answers := []string{
		// Positive outcomes
		"It is certain",
		"It is decidedly so",
		"Without a doubt",
		"Yes definitely",
		"You may rely on it",
		"As I see it, yes",
		"Most likely",
		"Outlook good",
		"Yes",
		"Signs point to yes",

		// Neutral outcomes
		"Reply hazy try again",
		"Ask again later",
		"Better not tell you now",
		"Cannot predict now",
		"Concentrate and ask again",

		// Negative outcomes
		"Don't count on it",
		"My reply is no",
		"My sources say no",
		"Outlook not so good",
		"Very doubtful",
	}

	rand.Seed(time.Now().Unix() + bot.CreatedAt.Unix())

	args := ctx.GetArgs()
	if len(args) < 2 {
		_, err := utils.ReplyWithEmbed(ctx, utils.ErrorEmbed(
			ctx,
			errors.New("Missing question\n"+
				"Usage: `eightball <question>`"),
		))

		return err
	}

	question := strings.Join(args, " ")

	answer := answers[rand.Intn(len(answers))]

	embed := utils.GenerateEmbed(ctx, discordgo.MessageEmbed{
		Title: "Magic Eightball :8ball:",
		Color: 0x00ff00,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:  "Question",
				Value: question,
			},
			{
				Name:  "Answer",
				Value: ":8ball: *" + answer + "*",
			},
		},
	})

	_, err := utils.ReplyWithEmbed(ctx, embed)
	return err
}
