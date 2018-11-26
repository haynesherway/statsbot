package statsbot

import (
	"encoding/json"
	"errors"
	"flag"
	"io/ioutil"
	"log"
	"path"
	"runtime"
)

// Config values
var (
	Token     string
	BotPrefix string
	Server    string
	test      bool

	config *configStruct
)

type configStruct struct {
	Token     string `json:"Token"`
	BotPrefix string `json:"BotPrefix"`
	Server    string `json:"Server"`
}

// ReadConfig reads the config file and initializes values using those configs
func ReadConfig() error {
	log.Println("Reading from config file...")

	_, filename, _, ok := runtime.Caller(1)
	if !ok {
		return errors.New("Unable to get config file location")
	}
	file, err := ioutil.ReadFile(path.Join(path.Dir(filename), "../config.json"))
	if err != nil {
		log.Println(err.Error())
		return err
	}

	err = json.Unmarshal(file, &config)
	if err != nil {
		log.Println(err.Error())
		return err
	}

	Token = config.Token
	BotPrefix = config.BotPrefix
	Server = config.Server

	if test {
		BotPrefix = "?"
	}
	return nil
}

func init() {
	flag.BoolVar(&test, "t", false, "Run for testing")
	flag.Parse()
}
