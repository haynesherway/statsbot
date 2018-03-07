package statsbot

import (
	"errors"
	"log"
	"strconv"
	"strings"

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
