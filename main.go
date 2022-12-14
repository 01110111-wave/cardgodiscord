package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

type Downloadlog struct {
	Name   string `json:"filename"`
	Author string `json:"author"`
	Url    string `json:"messageurl"`
}

type Faillog struct {
	Author  string `json:"author"`
	Content string `json:"content"`
	Url     string `json:"messageurl"`
}

var downloadlogfile *os.File
var faileddownloadlogfile *os.File
var channelID string = "607148922057785344"
var count int = 0
var wg sync.WaitGroup

func main() {
	start := time.Now()
	os.Mkdir("cards", 0644)
	downloadlogfile, _ = os.Create("cards/downloadlog.json")
	faileddownloadlogfile, _ = os.Create("cards/faileddownloadlog.json")
	downloadlogfile.WriteString("[")
	faileddownloadlogfile.WriteString("[")
	firstid := "616927278529773571"
	err := godotenv.Load(".env")
	dg, err := discordgo.New(os.Getenv("token")) //"Bot " + "MTAwMDc3ODAwMDIzMDUxNDc0OQ.GzDwcG.DmSSBcwbAz_Bh1A29S2tAng4mGQdhRlI0myGj0")
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}
	wg.Add(1)
	go gotillast(dg, firstid, "")
	wg.Wait()
	fmt.Println("end!!!")
	downloadlogfile.WriteString("]")
	downloadlogfile.Close()
	faileddownloadlogfile.WriteString("]")
	faileddownloadlogfile.Close()
	fmt.Println("take:", time.Since(start))
}

func readmsg(message *discordgo.Message) {

	downloaded := false
	for _, att := range message.Attachments {
		if att.Width == 252 && att.Height == 352 {
			fmt.Println("i react now download", message.Author.Username, message.ID, message.Attachments)
			wg.Add(1)
			go downloadcard(att, message.ChannelID, message.ID, message.Author.Username+"#"+message.Author.Discriminator)
			downloaded = true
		}
	}
	if !downloaded {
		log := Faillog{
			message.Author.Username + "#" + message.Author.Discriminator,
			message.Content,
			"https://" + "discord.com/channels/" + "446784086539763712" + "/" + message.ChannelID + "/" + message.ID,
		}
		file, _ := json.MarshalIndent(log, "", " ")
		//lockfilefaillog.Lock()
		faileddownloadlogfile.Write(file)
		faileddownloadlogfile.WriteString(",")
		//lockfilefaillog.Unlock()
	}

	wg.Done()
}

func downloadcard(att *discordgo.MessageAttachment, channelID string, messageID string, author string) {
	res, err := http.Get(att.URL)
	fmt.Println("download", att.Filename, "from", author)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return
	}
	cheakandcreatedir(author)
	os.WriteFile("cards/"+author+"/"+att.Filename, body, 0644)
	log := Downloadlog{
		att.Filename,
		author,
		"https://" + "discord.com/channels/" + "446784086539763712" + "/" + channelID + "/" + messageID,
	}
	file, _ := json.MarshalIndent(log, "", " ")
	//lockfile.Lock()
	downloadlogfile.Write(file)
	downloadlogfile.WriteString(",")
	//lockfile.Unlock()
	wg.Done()
}

func gotillast(dg *discordgo.Session, firstid string, currentid string) {
	messages, err := dg.ChannelMessages(channelID, 0, currentid, "", "")
	if len(messages) == 0 {
		fmt.Println("done id:", currentid)
		wg.Done()
		return
	}
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}
	if messages[0].ID != firstid {
		currentid = messages[len(messages)-1].ID
		wg.Add(1)
		go gotillast(dg, firstid, currentid)
		for _, message := range messages {
			if check := ireact(message); check {
				wg.Add(1)
				go readmsg(message)
			}
		}
	} else {
		for _, message := range messages {
			if check := ireact(message); check {
				wg.Add(1)
				go readmsg(message)
			}
		}
	}
	wg.Done()
}

func ireact(message *discordgo.Message) bool {
	//ch := make(chan bool)
	for _, react := range message.Reactions {
		if react.Me {
			return true
		}
		// go reactme(react,ch)
	}
	// isreacted := <-ch
	return false
}

func cheakandcreatedir(path string) {
	if _, err := os.Stat("cards/" + path); os.IsNotExist(err) {
		err := os.Mkdir("cards/"+path, 0644)
		if err != nil {
			fmt.Printf("err create file: %v\n", err)
			return
		}
		return
	}
}

func reactme(react *discordgo.MessageReactions, ch chan bool) {
	if react.Me {
		ch <- react.Me
	}
	return
}

//mutex lock everything inside function
