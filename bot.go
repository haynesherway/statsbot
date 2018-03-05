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
)

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
	if !strings.HasPrefix(m.Content, "!stats") {
		return
	}

	if m.Author.ID == BotID {
		return
	}

	message := strings.TrimSpace(strings.Replace(m.Content, "!stats", "", 1))

	if len(message) == 0 {
		return
	}
	
	fmt.Println(message)

	fields := strings.Split(message, " ")
	if len(fields) < 1 {
		return
	}

	switch fields[0] {
	case "print":
		stats, err := PrintStats(strings.Replace(message, "print ", "", 1))
		if err != nil {
			_, _ = s.ChannelMessageSend(m.ChannelID, err.Error())
		}
		
		_, _ = s.ChannelMessageSend(m.ChannelID, stats)
	//Print all stats
	case "add":
		err := AddStat(m.Author, strings.Replace(message, "add ", "", 1))
		if err != nil {
			_, _ = s.ChannelMessageSend(m.ChannelID, err.Error())
		} else {
			_, _ = s.ChannelMessageSend(m.ChannelID, "Successfully added stat!")
		}
	case "user":
		message := strings.Replace(message, "user ", "", 1)
		fields := strings.Split(message, " ")
		var user *discordgo.User
		var name string
		var err error
		if len(fields) > 0 {
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
			
			AddUser(user, name)
			
		} else {
			err = errors.New("Command unrecognized.")
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
				_, _ = s.ChannelMessageSend(m.ChannelID, users)
			}
		case "categories":
			categories, err := PrintCategories()
			if err != nil {
				_, _ = s.ChannelMessageSend(m.ChannelID, err.Error())
			} else {
				_, _ = s.ChannelMessageSend(m.ChannelID, categories)
			}
	}
	

	return

}
