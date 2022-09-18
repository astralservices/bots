package integrations

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"net/smtp"
	"os"
	"strings"

	"github.com/astralservices/bots/pkg/commands/integrations"
	db "github.com/astralservices/bots/pkg/database/supabase"
	"github.com/astralservices/bots/pkg/types"
	"github.com/astralservices/bots/pkg/utils"
	"github.com/astralservices/dgc"
	"github.com/bwmarrin/discordgo"
	uuid "github.com/nu7hatch/gouuid"
)

var VerifyStudentCommand = &dgc.Command{
	Name:        "verifystudent",
	Domain:      "astral.integrations.verifystudent",
	Aliases:     []string{"verifystudent", "verifystudentemail", "vs", "vse"},
	Description: "Verify your student email address.",
	Category:    "College",
	Usage:       "verifystudent <email>",
	Slash:       true,
	SlashGuilds: []string{os.Getenv("DEV_GUILD")},
	Handler: func(ctx *dgc.Ctx) {
		email := ctx.Arguments.Get(0).Raw()

		database := db.New()

		wi, err := integrations.GetWorkspaceIntegrationForCommand(ctx, CollegeIntegrationID)

		if err != nil {
			ctx.ReplyEmbed(utils.GenerateEmbed(*ctx, discordgo.MessageEmbed{
				Title:       "Error",
				Description: "An error occurred while fetching the workspace integration.",
				Fields: []*discordgo.MessageEmbedField{
					{
						Name:   "Error",
						Value:  err.Error(),
						Inline: false,
					},
				},
				Color: 0xff0000,
			}))

			return
		}

		// get the user's data
		data, err := database.GetIntegrationDataForUser(ctx.Event.Author.ID, CollegeIntegrationID, wi.ID)

		if err != nil {
			ctx.ReplyEmbed(utils.GenerateEmbed(*ctx, discordgo.MessageEmbed{
				Title:       "Error",
				Description: "An error occurred while fetching your data.",
				Fields: []*discordgo.MessageEmbedField{
					{
						Name:   "Error",
						Value:  err.Error(),
						Inline: false,
					},
				},
				Color: 0xff0000,
			}))

			return
		}

		var d types.CollegeIntegrationData

		jsonStr, err := json.Marshal(data.Data)

		if err != nil {
			ctx.ReplyEmbed(utils.GenerateEmbed(*ctx, discordgo.MessageEmbed{
				Title:       "Error",
				Description: "An error occurred while fetching your data.",
				Fields: []*discordgo.MessageEmbedField{
					{
						Name:   "Error",
						Value:  err.Error(),
						Inline: false,
					},
				},
				Color: 0xff0000,
			}))

			return
		}

		err = json.Unmarshal(jsonStr, &d)

		if err != nil {
			ctx.ReplyEmbed(utils.GenerateEmbed(*ctx, discordgo.MessageEmbed{
				Title:       "Error",
				Description: "An error occurred while fetching your data.",
				Fields: []*discordgo.MessageEmbedField{
					{
						Name:   "Error",
						Value:  err.Error(),
						Inline: false,
					},
				},
				Color: 0xff0000,
			}))

			return
		}

		// check if the user has already verified their email
		if data.Data.(map[string]interface{})["email"].(map[string]interface{})["verified"].(bool) {
			var fields []*discordgo.MessageEmbedField = []*discordgo.MessageEmbedField{
				{
					Name:   "Email",
					Value:  data.Data.(map[string]interface{})["email"].(map[string]interface{})["address"].(string),
					Inline: false,
				},
			}

			// give roles, if any
			if wi.Settings.(map[string]interface{})["verifiedRoleId"] != nil {
				roleID := wi.Settings.(map[string]interface{})["verifiedRoleId"].(string)

				err = ctx.Session.GuildMemberRoleAdd(ctx.Event.GuildID, ctx.Event.Author.ID, roleID)

				if err != nil {
					ctx.ReplyEmbed(utils.GenerateEmbed(*ctx, discordgo.MessageEmbed{
						Title:       "Error",
						Description: "An error occurred while adding your role.",
						Fields: []*discordgo.MessageEmbedField{
							{
								Name:   "Error",
								Value:  err.Error(),
								Inline: false,
							},
						},
						Color: 0xff0000,
					}))

					return
				}

				fields = append(fields, &discordgo.MessageEmbedField{
					Name:   "Role",
					Value:  fmt.Sprintf("<@&%s>", roleID),
					Inline: false,
				})
			}

			ctx.ReplyEmbed(utils.GenerateEmbed(*ctx, discordgo.MessageEmbed{
				Title:       "Verified!",
				Description: "You've verified your email, and I've given you any roles you're eligible for.",
				Fields:      fields,
				Color:       0x00ff00,
			}))

			return
		}

		// check the email domain
		if wi.Settings.(map[string]interface{})["emailDomain"] != email[strings.LastIndex(email, "@")+1:] {
			ctx.ReplyEmbed(utils.GenerateEmbed(*ctx, discordgo.MessageEmbed{
				Title:       "Error",
				Description: "The email domain you provided is not valid for this workspace.",
				Color:       0xff0000,
			}))

			return
		}

		if email == "" {
			ctx.ReplyEmbed(utils.ErrorEmbed(*ctx, fmt.Errorf("Please provide an email address.")))
			return
		}

		uuid, err := uuid.NewV4()

		if err != nil {
			ctx.ReplyEmbed(utils.GenerateEmbed(*ctx, discordgo.MessageEmbed{
				Title:       "Error",
				Description: "An error occurred while generating a verification code.",
				Fields: []*discordgo.MessageEmbedField{
					{
						Name:   "Error",
						Value:  err.Error(),
						Inline: false,
					},
				},
				Color: 0xff0000,
			}))

			return
		}

		// send the verification email
		err = SendEmail([]string{email}, uuid.String(), wi.Workspace, wi.Integration)

		if err != nil {
			ctx.ReplyEmbed(utils.GenerateEmbed(*ctx, discordgo.MessageEmbed{
				Title:       "Error",
				Description: "An error occurred while sending the verification email.",
				Fields: []*discordgo.MessageEmbedField{
					{
						Name:   "Error",
						Value:  err.Error(),
						Inline: false,
					},
				},
				Color: 0xff0000,
			}))

			return
		}

		err = database.SetIntegrationDataForUser(ctx.Event.Author.ID, CollegeIntegrationID, wi.ID, map[string]interface{}{
			"room":  d.Room,
			"house": d.House,
			"email": map[string]interface{}{
				"verified":         false,
				"address":          email,
				"verificationCode": uuid.String(),
			},
		})

		if err != nil {
			ctx.ReplyEmbed(utils.GenerateEmbed(*ctx, discordgo.MessageEmbed{
				Title:       "Error",
				Description: "An error occurred while setting your email.",
				Fields: []*discordgo.MessageEmbedField{
					{
						Name:   "Error",
						Value:  err.Error(),
						Inline: false,
					},
				},
				Color: 0xff0000,
			}))

			return
		}

		ctx.ReplyEmbed(utils.GenerateEmbed(*ctx, discordgo.MessageEmbed{
			Title:       "Success",
			Description: "Your email has been set to `" + email + "`! Check your email for a verification link.",
		}))
	},
}

func SendEmail(to []string, code string, workspace string, integration string) error {
	sender := "support@astralapp.io"

	user := os.Getenv("SMTP_USER")
	password := os.Getenv("SMTP_PASSWORD")

	subject := "Astral Verification Code"
	tmpl, err := template.ParseFiles("verify.html")

	if err != nil {
		return err
	}

	buf := new(bytes.Buffer)

	err = tmpl.Execute(buf, struct {
		Code        string
		AuthUrl     string
		Workspace   string
		Integration string
	}{
		Code:        code,
		AuthUrl:     os.Getenv("AUTH_URL"),
		Workspace:   workspace,
		Integration: integration,
	})

	if err != nil {
		return err
	}

	request := Mail{
		Sender:  sender,
		To:      to,
		Subject: subject,
		Body:    buf.String(),
	}

	addr := "smtp.gmail.com:587"
	host := "smtp.gmail.com"

	msg := BuildMessage(request)
	auth := smtp.PlainAuth("", user, password, host)
	err = smtp.SendMail(addr, auth, sender, to, []byte(msg))

	if err != nil {
		return err
	}

	return nil
}

func BuildMessage(mail Mail) string {
	msg := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\r\n"
	msg += fmt.Sprintf("From: %s\r\n", mail.Sender)
	msg += fmt.Sprintf("To: %s\r\n", strings.Join(mail.To, ";"))
	msg += fmt.Sprintf("Subject: %s\r\n", mail.Subject)
	msg += fmt.Sprintf("\r\n%s\r\n", mail.Body)

	return msg
}

type Mail struct {
	Sender  string
	To      []string
	Subject string
	Body    string
}
