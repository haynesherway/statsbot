package statsbot

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
)

// BotID for discord
var (
	BotID string
	goBot *discordgo.Session

	ERR_COMMAND_UNRECOGNIZED = errors.New("Command not recognized")
)

type botResponse struct {
	s       *discordgo.Session
	m       *discordgo.MessageCreate
	command string
	fields  []string
	err     error
}

// Type Do is a placeholder for the function a command should execute
type Do func(b *botResponse) error

// BotCommand is a representation of a command the bot can handle
type BotCommand struct {
	Name    string
	Format  string
	Info    string
	Example []string
	Print   bool
	Aliases []string
	Do
}

var cmdMap map[string]BotCommand
var cmdList map[string]BotCommand

// PrintInfo prints the info for a discord command
func (cmd *BotCommand) PrintInfo(prefix string) string {
	examples := Example(strings.Replace(cmd.Format, "!", prefix, 1))
	for _, ex := range cmd.Example {
		examples += Example(strings.Replace(ex, "!", prefix, 1))
	}
	return fmt.Sprintln(cmd.Info, examples)
}

//NewBotResponse creates an instance of a bot interaction
func NewBotResponse(s *discordgo.Session, m *discordgo.MessageCreate, fields []string) *botResponse {
	return &botResponse{s: s, m: m, fields: fields}
}

// GetCommand gets the BotCommand for the input
func (b *botResponse) GetCommand(prefix string) (cmd *BotCommand) {
	if len(b.fields) == 0 {
		b.err = ERR_COMMAND_UNRECOGNIZED
		return cmd
	}

	name := strings.ToLower(strings.Replace(b.fields[1], prefix, "", 1))
	if c, ok := cmdMap[name]; ok {
		return &c
	} else {
		b.err = ERR_COMMAND_UNRECOGNIZED
		return cmd
	}
}

// Start starts the bot
func Start() {
	var err error
	goBot, err = discordgo.New("Bot " + config.Token)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	u, err := goBot.User("@me")
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	BotID = u.ID

	err = InitDB()
	if err != nil {
		log.Fatalf("Unable to start stats database: %+v\n", err.Error())
	}

	goBot.AddHandler(messageHandler)
	err = goBot.Open()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	err = goBot.UpdateStatus(0, "!stats")
	if err != nil {
		fmt.Println("Unable to update status: ", err.Error())
	}

	log.Println("Bot is running!")
}

func messageHandler(s *discordgo.Session, m *discordgo.MessageCreate) {

	if !strings.HasPrefix(m.Content, BotPrefix+"stats") {
		return
	}

	if m.Author.ID == BotID {
		return
	}

	lines := strings.Split(m.Content, "\n")

	for _, line := range lines {

		msg := strings.TrimSpace(strings.Replace(line, BotPrefix+"stats ", "", -1))
		fmt.Println(msg)
		if len(msg) == 0 {
			continue
		}
		handleLine(s, m, msg)
	}
	return
}

func handleLine(s *discordgo.Session, m *discordgo.MessageCreate, message string) {
	//bot := NewBotResponse(s, m, strings.Fields(m.Content))

	fields := strings.Split(message, " ")
	if len(fields) < 1 {
		return
	}

	switch fields[0] {
	case "help":
		helpmessage := "use `!stats user` to add your user to the stats program, if not already added."
		helpmessage += "\nuse `!stats add {category} {value}` to add your stats."
		helpmessage += "\nuse `!stats categories` to see all available categories"
		helpmessage += "\nuse `!stats users` to see all users"
		helpmessage += "\nuse `!stats print {category}` to print rankings"

		_, _ = s.ChannelMessageSend(m.ChannelID, helpmessage)
	case "print":
		c := strings.TrimSpace(strings.Replace(message, "print", "", 1))
		if c == "" || c == "all" {
			categories, err := GetCategories()
			if err != nil {
				log.Println(err.Error())
			}
			for _, category := range categories {
				stats, err := PrintStats(category.Name)
				if err != nil {
					_, _ = s.ChannelMessageSend(m.ChannelID, err.Error())
				} else {
					emb := NewEmbed().
						SetColor(0x0B9EFF).SetAuthor(category.FullName, category.Image).
						SetDescription(stats).MessageEmbed
					_, _ = s.ChannelMessageSendEmbed(m.ChannelID, emb)
					//_, _ = s.ChannelMessageSend(m.ChannelID, stats)
				}
			}
		} else {
			category, _ := GetCategory(c)
			stats, err := PrintStats(c)
			if err != nil {
				_, _ = s.ChannelMessageSend(m.ChannelID, err.Error())
			} else {
				emb := NewEmbed().
					SetColor(0x0B9EFF).SetAuthor(category.FullName, category.Image).
					SetDescription(stats).MessageEmbed
				_, _ = s.ChannelMessageSendEmbed(m.ChannelID, emb)
			}
			//_, _ = s.ChannelMessageSend(m.ChannelID, stats)
		}
	//Print all stats
	case "add":
		err := AddStat(m.Author, strings.Replace(message, "add ", "", 1))
		if err != nil {
			_, _ = s.ChannelMessageSend(m.ChannelID, err.Error())
		} else {
			_, _ = s.ChannelMessageSend(m.ChannelID, "Successfully added stat!")
		}
	case "user":
		message = strings.TrimSpace(strings.Replace(message, "user", "", 1))
		fields := strings.Split(message, " ")
		var user *discordgo.User
		var name string
		var err error

		if len(fields) > 0 {
			if !CheckAdmin(m.Author.ID) {
				return
			}
			if strings.Contains(message, "<@") {
				user = m.Mentions[0]
			} else {
				username := fields[0]

				channel, err := s.Channel(m.ChannelID)
				guild, err := s.Guild(channel.GuildID)
				if err != nil {
					break
				}

				for _, mem := range guild.Members {
					if mem.Nick == username {
						user = mem.User
						name = username
					} else if mem.User.Username == username {
						user = mem.User
					}
				}
			}

			if len(fields) == 2 {
				name = fields[1]
			}

			if user != nil {
				AddUser(user, name)
			} else {
				err = errors.New("User not found.")
			}

		} else {
			fmt.Println("Adding...")
			AddUser(m.Author, "")
			//err = errors.New("Command unrecognized.")
		}
		if err != nil {
			_, _ = s.ChannelMessageSend(m.ChannelID, err.Error())
		} else {
			_, _ = s.ChannelMessageSend(m.ChannelID, "Successfully added user!")
		}
	case "users":
		users, err := PrintUsers()
		if err != nil {
			_, _ = s.ChannelMessageSend(m.ChannelID, err.Error())
		} else {
			emb := NewEmbed().
				SetColor(0x0B9EFF).
				AddField("Users", users).MessageEmbed
			_, _ = s.ChannelMessageSendEmbed(m.ChannelID, emb)
			//_, _ = s.ChannelMessageSend(m.ChannelID, users)
		}
	case "categories":
		categories, err := PrintCategories()
		if err != nil {
			_, _ = s.ChannelMessageSend(m.ChannelID, err.Error())
		} else {
			emb := NewEmbed().
				SetColor(0x0B9EFF).
				AddField("Categories", categories).MessageEmbed
			_, _ = s.ChannelMessageSendEmbed(m.ChannelID, emb)
		}
	case "rank", "ranks":
		message = strings.TrimSpace(strings.Replace(strings.Replace(message, "ranks", "", 1), "rank", "", 1))
		fields := strings.Split(message, " ")
		userID := ""
		fmt.Println(fields)
		if len(fields) > 0 && fields[0] != "" {
			u, err := GetUser(fields[0])
			if err != nil {
				_, _ = s.ChannelMessageSend(m.ChannelID, "Unable to find user")
				return
			}
			userID = u.DiscordID
		} else {
			userID = m.Author.ID
		}
		user, err := GetUser(userID)
		if err != nil {
			_, _ = s.ChannelMessageSend(m.ChannelID, err.Error())
		}
		ranks, err := PrintRanks(user)
		if err != nil {
			_, _ = s.ChannelMessageSend(m.ChannelID, err.Error())
		} else {
			_, _ = s.ChannelMessageSend(m.ChannelID, ranks)
		}
	case "remind":
		if !CheckAdmin(m.Author.ID) {
			return
		}
		err := SendReminders(s, m.ChannelID)
		if err != nil {
			_, _ = s.ChannelMessageSend(m.ChannelID, err.Error())
		}
	}
	return

}

// PrintToDiscord prints the message string to discord
func (b *botResponse) PrintToDiscord(msg string) {
	_, _ = b.s.ChannelMessageSend(b.m.ChannelID, msg)
	return
}

// Print embed to discord prints an embed to discord
func (b *botResponse) PrintEmbedToDiscord(e *discordgo.MessageEmbed) {
	_, _ = b.s.ChannelMessageSendEmbed(b.m.ChannelID, e)
}

type botError struct {
	err   error
	value string
}
