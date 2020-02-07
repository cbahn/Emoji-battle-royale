package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/bwmarrin/discordgo"
)

func main() {

	battleid := "gogogogo"
	guildid := "152893724500819969"
	botkey := "NjcyNTQ3NTU0MDA0ODI4MTgw.Xj20TA.-PawAcfzEC8Mae-dNbl7WyLD1XM"

	discord, derr := discordgo.New("Bot " + botkey)
	if derr != nil {
		log.Fatal(derr)
	}

	guild, gerr := discord.Guild(guildid)
	if gerr != nil {
		log.Fatal(gerr)
	}

	path := "../public/" + battleid + "/"

	if _, err := os.Stat("../public"); os.IsNotExist(err) {
		os.Mkdir(path, 0755)
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.Mkdir(path, 0755)
	}
	moderr := os.Chmod(path, 0755)
	if moderr != nil {
		fmt.Println(moderr)
	}

	for _, emoji := range guild.Emojis {
		file := emoji.ID + ".png"
		if emoji.Animated {
			file = emoji.ID + ".gif"
		}
		url := discordgo.EndpointCDN + "emojis/" + file
		filepath := path + file

		err := DownloadFile(filepath, url)
		if err != nil {
			log.Fatal(err)
		}
	}
}

// DownloadFile saves each emoji to the project folder
func DownloadFile(filepath string, url string) error {

	// Create the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}
