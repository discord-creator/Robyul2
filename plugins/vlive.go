package plugins

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Seklfreak/Robyul2/helpers"
	"github.com/Seklfreak/Robyul2/logger"
	"github.com/bwmarrin/discordgo"
	"math/big"
	"strconv"
	"strings"
	"time"
)

var (
	VliveAppId string = "8c6cc7b45d2568fb668be6e05b6e5a3b"
)

const (
	VliveEndpointDecodeChannelCode string = "http://api.vfan.vlive.tv/vproxy/channelplus/decodeChannelCode?app_id=%s&channelCode=%s"
	VliveEndpointChannel           string = "http://api.vfan.vlive.tv/channel.%d?app_id=%s&fields=channel_seq,channel_code,type,channel_name,fan_count,channel_cover_img,channel_profile_img,representative_color,celeb_board"
	VliveEndpointChannelVideoList  string = "http://api.vfan.vlive.tv/vproxy/channelplus/getChannelVideoList?app_id=%s&channelSeq=%d&maxNumOfRows=%d"
	VliveEndpointUpcomingVideoList string = "http://api.vfan.vlive.tv/vproxy/channelplus/getUpcomingVideoList?app_id=%s&channelSeq=%d&maxNumOfRows=%d"
	VliveEndpointNotices           string = "http://notice.vlive.tv/notice/list.json?channel_seq=%d"
	VliveEndpointCeleb             string = "http://api.vfan.vlive.tv/board.%d/posts?app_id=%s"
	VliveFriendlyChannel           string = "http://channels.vlive.tv/%s"
	VliveFriendlyVideo             string = "http://www.vlive.tv/video/%d"
	VliveFriendlyNotice            string = "http://channels.vlive.tv/%s/notice/%d"
	VliveFriendlyCeleb             string = "http://channels.vlive.tv/%s/celeb/%s"
)

type VLive struct{}

type DB_Entry struct {
	ID             string          `gorethink:"id,omitempty"`
	ServerID       string          `gorethink:"serverid"`
	ChannelID      string          `gorethink:"channelid"`
	VLiveChannel   DB_VLiveChannel `gorethink:"vlivechannel"`
	PostedUpcoming []DB_Video      `gorethink:"posted_upcoming"`
	PostedLive     []DB_Video      `gorethink:"posted_live"`
	PostedVOD      []DB_Video      `gorethink:"posted_vod"`
	PostedNotice   []DB_Notice     `gorethink:"posted_notices"`
	PostedCelebs   []DB_Celeb      `gorethink:"posted_celebs"`
}

type DB_VLiveChannel struct {
	Seq           int64  `gorethink:"seq,omitempty" json:"channel_seq"`
	Code          string `gorethink:"code,omitempty" json:"channel_code"`
	Type          string `json:"type"`
	Name          string `json:"channel_name"`
	Followers     int64  `json:"fan_count"`
	CoverImgUrl   string `json:"channel_cover_img"`
	ProfileImgUrl string `json:"channel_profile_img"`
	Color         string `json:"representative_color"`
	TotalVideos   int64  `json:"-"`
	CelebBoard    struct {
		BoardID int64 `json:"board_id"`
	} `json:"celeb_board"`
	Upcoming []DB_Video  `gorethink:"upcoming" json:"-"`
	Live     []DB_Video  `gorethink:"live" json:"-"`
	VOD      []DB_Video  `gorethink:"vod" json:"-"`
	Notices  []DB_Notice `gorethink:"notices" json:"-"`
	Celebs   []DB_Celeb  `gorethink:"celebs" json:"-"`
	Url      string      `json:"-"`
}

type DB_Video struct {
	Seq       int64  `gorethink:"seq,omitempty" json:"videoSeq"`
	Title     string `json:"title"`
	Plays     int64  `json:"playCount"`
	Likes     int64  `json:"likeCount"`
	Comments  int64  `json:"commentCount"`
	Thumbnail string `json:"thumbnail"`
	Date      string `json:"onAirStartAt"`
	Playtime  int64  `json:"playTime"`
	Type      string `json:"videoType"`
	Url       string `json:"-"`
}

type DB_Notice struct {
	Number   int64  `gorethink:"number,omitempty" json:"noticeNo"`
	Title    string `json:"title"`
	ImageUrl string `json:"listImageUrl"`
	Summary  string `json:"summary"`
	Url      string `json:"-"`
}

type DB_Celeb struct {
	ID      string `gorethink:"id,omitempty" json:"post_id"`
	Summary string `json:"body_summary"`
	Url     string `json:"-"`
}

func (r *VLive) Commands() []string {
	return []string{
		"vlive",
	}
}

func (r *VLive) Init(session *discordgo.Session) {
	go func() {
		defer helpers.Recover()

		for {
			// TODO: Everything
			// var reminderBucket []DB_VLive
			// cursor, err := rethink.Table("VLive").Run(helpers.GetDB())
			// helpers.Relax(err)

			// err = cursor.All(&reminderBucket)
			// helpers.Relax(err)

			// for _, VLive := range reminderBucket {
			// 	changes := false

			// 	// Downward loop for in-loop element removal
			// 	for idx := len(VLive.VLive) - 1; idx >= 0; idx-- {
			// 		reminder := VLive.VLive[idx]

			// 		if reminder.Timestamp <= time.Now().Unix() {
			// 			session.ChannelMessageSend(
			// 				reminder.ChannelID,
			// 				":alarm_clock: Ring! Ring! <@"+VLive.UserID+">\n"+
			// 					"You wanted me to remind you to `"+reminder.Message+"` :slight_smile:",
			// 			)

			// 			VLive.VLive = append(VLive.VLive[:idx], VLive.VLive[idx+1:]...)
			// 			changes = true
			// 		}
			// 	}

			// 	if changes {
			// 		setVLive(VLive.UserID, VLive)
			// 	}
			// }

			time.Sleep(60 * time.Second)
		}
	}()

	logger.PLUGIN.L("VLive", "Started vlive loop (60s)")
}

func (r *VLive) Action(command string, content string, msg *discordgo.Message, session *discordgo.Session) {
	args := strings.Split(content, " ")
	if len(args) >= 1 {
		switch args[0] {
		case "add":
			session.ChannelMessageSend(msg.ChannelID, "implement me please")
		default:
			// TODO: Search for channel name
			session.ChannelTyping(msg.ChannelID)
			vliveChannel, err := getVLiveChannelByVliveChannelId(args[0])
			if err != nil {
				session.ChannelMessageSend(msg.ChannelID, helpers.GetTextF("plugins.vlive.channel-not-found"))
				return
			}
			channelEmbed := &discordgo.MessageEmbed{
				Title:     helpers.GetTextF("plugins.vlive.channel-embed-title", vliveChannel.Name),
				URL:       vliveChannel.Url,
				Thumbnail: &discordgo.MessageEmbedThumbnail{URL: vliveChannel.ProfileImgUrl},
				Footer:    &discordgo.MessageEmbedFooter{Text: helpers.GetText("plugins.vlive.embed-footer")},
				Fields: []*discordgo.MessageEmbedField{
					{Name: "Followers", Value: strconv.FormatInt(vliveChannel.Followers, 10), Inline: true},
					{Name: "Videos", Value: strconv.FormatInt(vliveChannel.TotalVideos, 10), Inline: true}},
			}
			if len(vliveChannel.Live) > 0 {
				channelEmbed.Fields = append(channelEmbed.Fields, &discordgo.MessageEmbedField{
					Name:   helpers.GetTextF("plugins.vlive.channel-embed-name-live", vliveChannel.Live[0].Date),
					Value:  fmt.Sprintf("**%s**\n%s", vliveChannel.Live[0].Title, vliveChannel.Live[0].Url),
					Inline: false,
				})
				channelEmbed.Image = &discordgo.MessageEmbedImage{URL: vliveChannel.Live[0].Thumbnail}
			} else if len(vliveChannel.VOD) > 0 {
				channelEmbed.Fields = append(channelEmbed.Fields, &discordgo.MessageEmbedField{
					Name:   helpers.GetTextF("plugins.vlive.channel-embed-name-vod", vliveChannel.VOD[0].Date),
					Value:  fmt.Sprintf("**%s**\n**Plays:** %d\n**Likes:** %d\n%s", vliveChannel.VOD[0].Title, vliveChannel.VOD[0].Plays, vliveChannel.VOD[0].Likes, vliveChannel.VOD[0].Url),
					Inline: false,
				})
				channelEmbed.Image = &discordgo.MessageEmbedImage{URL: vliveChannel.VOD[0].Thumbnail}
			}
			if len(vliveChannel.Upcoming) > 0 {
				channelEmbed.Fields = append(channelEmbed.Fields, &discordgo.MessageEmbedField{
					Name:   helpers.GetTextF("plugins.vlive.channel-embed-name-upcoming", vliveChannel.Upcoming[0].Date),
					Value:  fmt.Sprintf("**%s**\n%s", vliveChannel.Upcoming[0].Title, vliveChannel.Upcoming[0].Url),
					Inline: false,
				})
			}
			vliveChannelColorInt, ok := new(big.Int).SetString(strings.Replace(vliveChannel.Color, "#", "", 1), 16)
			if ok == true {
				channelEmbed.Color = int(vliveChannelColorInt.Int64())
			}
			_, err = session.ChannelMessageSendEmbed(msg.ChannelID, channelEmbed)
			if err != nil {
				helpers.Relax(err)
			}
			return
			//fmt.Println("Found channel:", vliveChannel.Name, "vod:", len(vliveChannel.VOD), "upcoming:", len(vliveChannel.Upcoming), "live:", len(vliveChannel.Live), "notices:", len(vliveChannel.Notices), "celeb:", len(vliveChannel.Celebs))
		}
	} else {
		session.ChannelMessageSend(msg.ChannelID, helpers.GetTextF("bot.arguments.too-few"))
	}
	// switch command {
	// case "rm", "remind":
	// 	channel, err := cache.Channel(msg.ChannelID)
	// 	helpers.Relax(err)

	// 	parts := strings.Split(content, " ")

	// 	if len(parts) < 3 {
	// 		session.ChannelMessageSend(msg.ChannelID, ":x: Please check if the format is correct")
	// 		return
	// 	}

	// 	r, err := r.parser.Parse(content, time.Now())
	// 	helpers.Relax(err)
	// 	if r == nil {
	// 		session.ChannelMessageSend(msg.ChannelID, ":x: Please check if the format is correct")
	// 		return
	// 	}

	// 	VLive := getVLive(msg.Author.ID)
	// 	VLive.VLive = append(VLive.VLive, DB_VLive{
	// 		Message:   strings.Replace(content, r.Text, "", 1),
	// 		ChannelID: channel.ID,
	// 		GuildID:   channel.GuildID,
	// 		Timestamp: r.Time.Unix(),
	// 	})
	// 	setVLive(msg.Author.ID, VLive)

	// 	session.ChannelMessageSend(msg.ChannelID, "Ok I'll remind you :ok_hand:")
	// 	break

	// case "rms", "VLive":
	// 	VLive := getVLive(msg.Author.ID)
	// 	embedFields := []*discordgo.MessageEmbedField{}

	// 	for _, reminder := range VLive.VLive {
	// 		ts := time.Unix(reminder.Timestamp, 0)
	// 		channel := "?"
	// 		guild := "?"

	// 		chanRef, err := session.Channel(reminder.ChannelID)
	// 		if err == nil {
	// 			channel = chanRef.Name
	// 		}

	// 		guildRef, err := session.Guild(reminder.GuildID)
	// 		if err == nil {
	// 			guild = guildRef.Name
	// 		}

	// 		embedFields = append(embedFields, &discordgo.MessageEmbedField{
	// 			Inline: false,
	// 			Name:   reminder.Message,
	// 			Value:  "At " + ts.String() + " in #" + channel + " of " + guild,
	// 		})
	// 	}

	// 	if len(embedFields) == 0 {
	// 		session.ChannelMessageSend(msg.ChannelID, helpers.GetText("remiders.empty"))
	// 		return
	// 	}

	// 	session.ChannelMessageSendEmbed(msg.ChannelID, &discordgo.MessageEmbed{
	// 		Title:  "Pending VLive",
	// 		Fields: embedFields,
	// 		Color:  0x0FADED,
	// 	})
	// 	break
	// }
}

func getVLiveChannelByVliveChannelId(channelId string) (DB_VLiveChannel, error) {
	var vliveChannel DB_VLiveChannel
	endpointDecodeChannelCode := fmt.Sprintf(VliveEndpointDecodeChannelCode, VliveAppId, channelId)
	jsonGabs := helpers.GetJSON(endpointDecodeChannelCode)
	resN, ok := jsonGabs.Path("result.channelSeq").Data().(float64)
	if ok == false {
		return vliveChannel, errors.New("Unable to get channel seq")
	}
	vliveChannelSeq := int(resN)

	endpointChannel := fmt.Sprintf(VliveEndpointChannel, vliveChannelSeq, VliveAppId)
	resB := helpers.NetGet(endpointChannel)

	json.Unmarshal(resB, &vliveChannel)
	vliveChannel.Url = fmt.Sprintf(VliveFriendlyChannel, channelId)

	// Get VODs and LIVEs
	var vliveVideo DB_Video
	endpointChannelVideoList := fmt.Sprintf(VliveEndpointChannelVideoList, VliveAppId, vliveChannelSeq, 10)
	jsonGabs = helpers.GetJSON(endpointChannelVideoList)

	resN, ok = jsonGabs.Path("result.totalVideoCount").Data().(float64)
	if ok == true {
		vliveChannel.TotalVideos = int64(resN)
	}

	videoListChildren, err := jsonGabs.Path("result.videoList").Children()
	if err == nil {
		for _, videoListEntry := range videoListChildren {
			err = json.Unmarshal([]byte(videoListEntry.String()), &vliveVideo)
			if err != nil {
				helpers.Relax(err)
			}
			vliveVideo.Url = fmt.Sprintf(VliveFriendlyVideo, vliveVideo.Seq)
			if vliveVideo.Type == "LIVE" {
				vliveChannel.Live = append(vliveChannel.VOD, vliveVideo)
			} else {
				vliveChannel.VOD = append(vliveChannel.VOD, vliveVideo)
			}
		}
	}
	// Get Upcomings
	endpointUpcomingVideoList := fmt.Sprintf(VliveEndpointUpcomingVideoList, VliveAppId, vliveChannelSeq, 10)
	jsonGabs = helpers.GetJSON(endpointUpcomingVideoList)
	videoListChildren, err = jsonGabs.Path("result.videoList").Children()
	if err == nil {
		for _, videoListEntry := range videoListChildren {
			err = json.Unmarshal([]byte(videoListEntry.String()), &vliveVideo)
			if err != nil {
				helpers.Relax(err)
			}
			vliveChannel.Upcoming = append(vliveChannel.Upcoming, vliveVideo)
		}

	}
	// Get Notices
	var vliveNotice DB_Notice
	endpointNotices := fmt.Sprintf(VliveEndpointNotices, vliveChannelSeq)
	jsonGabs = helpers.GetJSON(endpointNotices)
	noticesChildren, err := jsonGabs.Path("data").Children()
	if err == nil {
		for _, noticeEntry := range noticesChildren {
			err = json.Unmarshal([]byte(noticeEntry.String()), &vliveNotice)
			if err != nil {
				helpers.Relax(err)
			}
			vliveNotice.Url = fmt.Sprintf(VliveFriendlyNotice, channelId, vliveNotice.Number)
			vliveChannel.Notices = append(vliveChannel.Notices, vliveNotice)
		}
	}
	// Get Celeb
	if vliveChannel.CelebBoard.BoardID != 0 {
		var vliveCeleb DB_Celeb
		endpointCeleb := fmt.Sprintf(VliveEndpointCeleb, vliveChannel.CelebBoard.BoardID, VliveAppId)
		jsonGabs = helpers.GetJSON(endpointCeleb)
		celebsChildren, err := jsonGabs.Path("data").Children()
		if err == nil {
			for _, celebEntry := range celebsChildren {
				err = json.Unmarshal([]byte(celebEntry.String()), &vliveCeleb)
				if err != nil {
					helpers.Relax(err)
				}
				vliveCeleb.Url = fmt.Sprintf(VliveFriendlyCeleb, channelId, vliveCeleb.ID)
				vliveChannel.Celebs = append(vliveChannel.Celebs, vliveCeleb)
			}
		}
	}

	return vliveChannel, nil
}

// func getVLive(uid string) DB_VLive {
// 	var vlive DB_VLive
// 	return vlive
// 	// var reminderBucket DB_VLive
// 	// listCursor, err := rethink.Table("VLive").Filter(
// 	// 	rethink.Row.Field("userid").Eq(uid),
// 	// ).Run(helpers.GetDB())
// 	// defer listCursor.Close()
// 	// err = listCursor.One(&reminderBucket)

// 	// // If user has no DB entries create an empty document
// 	// if err == rethink.ErrEmptyResult {
// 	// 	_, e := rethink.Table("VLive").Insert(DB_VLive{
// 	// 		UserID: uid,
// 	// 		VLive:  make([]DB_VLive, 0),
// 	// 	}).RunWrite(helpers.GetDB())

// 	// 	// If the creation was successful read the document
// 	// 	if e != nil {
// 	// 		panic(e)
// 	// 	} else {
// 	// 		return getVLive(uid)
// 	// 	}
// 	// } else if err != nil {
// 	// 	panic(err)
// 	// }

// 	// return reminderBucket
// }

// func setVLive(uid string, VLive DB_VLive) {
// 	return
// 	// _, err := rethink.Table("VLive").Update(VLive).Run(helpers.GetDB())
// 	// helpers.Relax(err)
// }