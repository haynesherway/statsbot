package statsbot

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

var (
	ERR_NOT_ADMIN     = errors.New("Only admins can input other user's stats")
	ERR_INVALID_VALUE = errors.New("Invalid value.")
)

func AddStat(author *discordgo.User, msg string) (err error) {
	msg = strings.ToLower(msg)
	fields := strings.Split(msg, " ")

	//category [user] value

	var c,u,v string
	if len(fields) == 3 {
		// Admin entering other user
		if !CheckAdmin(author.ID) {
			return ERR_NOT_ADMIN
		}

		c, u, v = fields[0], fields[1], fields[2]
	} else if len(fields) == 2 {
		c, u, v = fields[0], author.ID, fields[1]

	} else {
		return ERR_INVALID_VALUE
	}
	category, err := GetCategory(c)
	if err != nil {
		log.Printf("Unable to get category: %+v\n", err.Error())
		return err
	}

	user, err := GetUser(u)
	if err != nil {
		log.Printf("Unable to get user: %+v\n", err.Error())
		return err
	}

	value, err := strconv.Atoi(v)
	if err != nil {
		log.Printf("Invalid value: %+v\n", err.Error())
		return ERR_INVALID_VALUE
	}

	err = category.AddStat(user, value)
	if err != nil {
		log.Printf(err.Error())
		return err
	}

	return

}

func AddUser(user *discordgo.User, name string) (err error) {
	if user == nil {
		return errors.New("User not found.")
	}
	if name == "" {
		name = user.Username
	}

	u := User{
		DiscordID: user.ID,
		Name:      name,
	}

	u.Insert()

	return nil
}

func PrintStats(msg string) (string, error) {
	msg = strings.ToLower(msg)
	fields := strings.Split(msg, " ")

	if len(fields) == 1 {
		c := fields[0]

		category, err := GetCategory(c)
		if err != nil {
			log.Printf("Unable to get category: %+v\n", err.Error())
			return "", err
		}

		return category.PrintStats()

	}
		return "", nil
}

func PrintRanks(u User) (string, error) {
	categories, err := GetCategories()
	if err != nil {
		return "", err
	}

	ranks := fmt.Sprintf("Ranks for %s:\n", u.Name)
	for _, category := range categories {
		r, err := category.GetUserRank(u.ID)
		if err != nil {
			return "", err
		}
		ranks += fmt.Sprintf("%s: %v\n", category.FullName, r)
	}

	return ranks, nil
}

func SendReminders(s *discordgo.Session, channelID string) error {
	users, err := GetActiveUsers()
	if err != nil {
		return err
	}

	categories, err := GetCategories() 
	if err != nil {
		return err
	}

	now := time.Now()
	cutoff := now.Add(-24 * time.Hour * 7)

	for _, user := range users {
		outdated := []string{}
		checks := map[string]bool{}
	
		for _, cat := range categories {
			checks[cat.FullName] = true
		}
		stats, err := user.GetStats()
		if err != nil {
			return err
		}
		for _, stat := range stats {
			if cutoff.After(stat.UpdatedAt) {
				outdated = append(outdated, stat.Category.FullName)
			}
			delete(checks, stat.Category.FullName)

		}
		missing := []string{}
		for c, _ := range checks {
			missing = append(missing, c)
		}
		if len(outdated) > 0 || len(missing) > 0 {
			dUser, err := s.User(user.DiscordID)
			var message string
			if err != nil {
				message = user.Name + ":\n"
			} else {
				message = dUser.Mention() + ":\n"
			}
			if len(outdated) > 0 {
				message += "You need to update the following stats: " + strings.Join(outdated, ", ") + "\n"
			}
			if len(missing) > 0 {
				message += "You are missing the following stats: " + strings.Join(missing, ", ") + "\n"
			}
			message += "Use `!stats help` for more information."
			_, _ = s.ChannelMessageSend(channelID, message)
		}
	}
	return nil
}
