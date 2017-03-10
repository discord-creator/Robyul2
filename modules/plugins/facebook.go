package plugins

import (
    "fmt"
    "github.com/Seklfreak/Robyul2/cache"
    "github.com/Seklfreak/Robyul2/helpers"
    "github.com/Seklfreak/Robyul2/logger"
    "github.com/bwmarrin/discordgo"
    "github.com/dustin/go-humanize"
    rethink "github.com/gorethink/gorethink"
    fb "github.com/huandu/facebook"
    "strings"
    "time"
)

type Facebook struct{}

type DB_Facebook_Page struct {
    ID          string             `gorethink:"id,omitempty"`
    ServerID    string             `gorethink:"serverid"`
    ChannelID   string             `gorethink:"channelid"`
    Username    string             `gorethink:"username"`
    PostedPosts []DB_Facebook_Post `gorethink:"posted_posts"`
}

type DB_Facebook_Post struct {
    ID        string `gorethink:"id,omitempty"`
    CreatedAt string `gorethink:"createdat`
}

type Facebook_Page struct {
    ID                string
    Name              string
    About             string
    Likes             int64
    Username          string
    Verified          bool
    ProfilePictureUrl string
    Website           string
    Posts             []Facebook_Post
}

type Facebook_Post struct {
    ID         string
    Message    string
    CreatedAt  string
    PictureUrl string
    Url        string
}

const (
    facebookHexColor     string = "#3b5998"
    FacebookFriendlyPage string = "https://facebook.com/%s/"
)

func (m *Facebook) Commands() []string {
    return []string{
        "facebook",
    }
}

func (m *Facebook) Init(session *discordgo.Session) {
    go func() {
        defer helpers.Recover()

        for {
            entryBucket := make([]DB_Facebook_Page, 0)
            cursor, err := rethink.Table("facebook").Run(helpers.GetDB())
            helpers.Relax(err)

            err = cursor.All(&entryBucket)
            helpers.Relax(err)

            // TODO: Check multiple entries at once
            for _, entry := range entryBucket {
                changes := false
                logger.VERBOSE.L("facebook", fmt.Sprintf("checking Facebook Page %s", entry.Username))

                facebookPage, err := m.lookupFacebookPage(entry.Username)
                if err != nil {
                    logger.ERROR.L("facebook", fmt.Sprintf("updating facebook account %s failed: %s", entry.Username, err.Error()))
                    continue
                }

                // https://github.com/golang/go/wiki/SliceTricks#reversing
                for i := len(facebookPage.Posts)/2 - 1; i >= 0; i-- {
                    opp := len(facebookPage.Posts) - 1 - i
                    facebookPage.Posts[i], facebookPage.Posts[opp] = facebookPage.Posts[opp], facebookPage.Posts[i]
                }

                for _, post := range facebookPage.Posts {
                    postAlreadyPosted := false
                    for _, postedPost := range entry.PostedPosts {
                        if postedPost.ID == post.ID {
                            postAlreadyPosted = true
                        }
                    }
                    if postAlreadyPosted == false {
                        logger.VERBOSE.L("facebook", fmt.Sprintf("Posting Post: #%s", post.ID))
                        entry.PostedPosts = append(entry.PostedPosts, DB_Facebook_Post{ID: post.ID, CreatedAt: post.CreatedAt})
                        changes = true
                        go m.postPostToChannel(entry.ChannelID, post, facebookPage)
                    }

                }
                if changes == true {
                    m.setEntry(entry)
                }
            }

            time.Sleep(10 * time.Minute)
        }
    }()

    logger.PLUGIN.L("facebook", "Started Facebook loop (10m)")
}

func (m *Facebook) Action(command string, content string, msg *discordgo.Message, session *discordgo.Session) {
    content = strings.Trim(content, " ")
    args := strings.Split(content, " ")
    if len(args) >= 1 {
        switch args[0] {
        case "add": // [p]facebook add <facebook page name> <discord channel>
            helpers.RequireAdmin(msg, func() {
                session.ChannelTyping(msg.ChannelID)
                // get target channel
                var err error
                var targetChannel *discordgo.Channel
                var targetGuild *discordgo.Guild
                if len(args) >= 3 {
                    targetChannel, err = helpers.GetChannelFromMention(args[2])
                    if err != nil {
                        session.ChannelMessageSend(msg.ChannelID, helpers.GetTextF("bot.arguments.invalid"))
                        return
                    }
                } else {
                    session.ChannelMessageSend(msg.ChannelID, helpers.GetTextF("bot.arguments.too-few"))
                    return
                }
                targetGuild, err = session.Guild(targetChannel.GuildID)
                helpers.Relax(err)
                // get facebook account and tweets
                facebookPage, err := m.lookupFacebookPage(args[1])
                if err != nil {
                    if e, ok := err.(*fb.Error); ok {
                        if e.Code == 803 {
                            session.ChannelMessageSend(msg.ChannelID, helpers.GetTextF("plugins.facebook.page-not-found"))
                            return
                        }
                    }
                    helpers.Relax(err)
                }
                // Create DB Entries
                var dbPosts []DB_Facebook_Post
                for _, facebookPost := range facebookPage.Posts {
                    facebookPostEntry := DB_Facebook_Post{ID: facebookPost.ID, CreatedAt: facebookPost.CreatedAt}
                    dbPosts = append(dbPosts, facebookPostEntry)

                }
                // create new entry in db
                entry := m.getEntryByOrCreateEmpty("id", "")
                entry.ServerID = targetChannel.GuildID
                entry.ChannelID = targetChannel.ID
                entry.Username = facebookPage.Username
                entry.PostedPosts = dbPosts
                m.setEntry(entry)

                session.ChannelMessageSend(msg.ChannelID, helpers.GetTextF("plugins.facebook.account-added-success", entry.Username, entry.ChannelID))
                logger.INFO.L("facebook", fmt.Sprintf("Added Facebook Account %s to Channel %s (#%s) on Guild %s (#%s)", entry.Username, targetChannel.Name, entry.ChannelID, targetGuild.Name, targetGuild.ID))
            })
        case "delete", "del": // [p]facebook delete <id>
            helpers.RequireAdmin(msg, func() {
                if len(args) >= 2 {
                    session.ChannelTyping(msg.ChannelID)
                    entryId := args[1]
                    entryBucket := m.getEntryBy("id", entryId)
                    if entryBucket.ID != "" {
                        m.deleteEntryById(entryBucket.ID)

                        session.ChannelMessageSend(msg.ChannelID, helpers.GetTextF("plugins.facebook.account-delete-success", entryBucket.Username))
                        logger.INFO.L("facebook", fmt.Sprintf("Deleted Facebook Page `%s`", entryBucket.Username))
                    } else {
                        session.ChannelMessageSend(msg.ChannelID, helpers.GetText("plugins.facebook.account-delete-not-found-error"))
                        return
                    }
                } else {
                    session.ChannelMessageSend(msg.ChannelID, helpers.GetText("bot.arguments.too-few"))
                    return
                }
            })
        case "list": // [p]facebook list
            currentChannel, err := session.Channel(msg.ChannelID)
            helpers.Relax(err)
            var entryBucket []DB_Facebook_Page
            listCursor, err := rethink.Table("facebook").Filter(
                rethink.Row.Field("serverid").Eq(currentChannel.GuildID),
            ).Run(helpers.GetDB())
            helpers.Relax(err)
            defer listCursor.Close()
            err = listCursor.All(&entryBucket)

            if err == rethink.ErrEmptyResult || len(entryBucket) <= 0 {
                session.ChannelMessageSend(msg.ChannelID, helpers.GetTextF("plugins.facebook.account-list-no-accounts-error"))
                return
            } else if err != nil {
                helpers.Relax(err)
            }

            resultMessage := ""
            for _, entry := range entryBucket {
                resultMessage += fmt.Sprintf("`%s`: Facebook Page `%s` posting to <#%s>\n", entry.ID, entry.Username, entry.ChannelID)
            }
            resultMessage += fmt.Sprintf("Found **%d** Facebook Pages in total.", len(entryBucket))
            for _, resultPage := range helpers.Pagify(resultMessage, "\n") {
                _, err = session.ChannelMessageSend(msg.ChannelID, resultPage)
                helpers.Relax(err)
            }
        default:
            session.ChannelTyping(msg.ChannelID)

            if args[0] == "" {
                session.ChannelMessageSend(msg.ChannelID, helpers.GetTextF("plugins.facebook.page-not-found"))
                return
            }

            facebookPage, err := m.lookupFacebookPage(args[0])
            if err != nil {
                if e, ok := err.(*fb.Error); ok {
                    if e.Code == 803 {
                        session.ChannelMessageSend(msg.ChannelID, helpers.GetTextF("plugins.facebook.page-not-found"))
                        return
                    }
                }
                helpers.Relax(err)
            }

            facebookNameModifier := ""
            if facebookPage.Verified {
                facebookNameModifier += " :ballot_box_with_check:"
            }

            accountEmbed := &discordgo.MessageEmbed{
                Title:       helpers.GetTextF("plugins.facebook.page-embed-title", facebookPage.Name, facebookPage.Username, facebookNameModifier),
                URL:         fmt.Sprintf(FacebookFriendlyPage, facebookPage.Username),
                Thumbnail:   &discordgo.MessageEmbedThumbnail{URL: facebookPage.ProfilePictureUrl},
                Footer:      &discordgo.MessageEmbedFooter{Text: helpers.GetText("plugins.facebook.embed-footer")},
                Description: facebookPage.About,
                Fields: []*discordgo.MessageEmbedField{
                    {Name: "Likes", Value: humanize.Comma(facebookPage.Likes), Inline: true}},
                Color: helpers.GetDiscordColorFromHex(facebookHexColor),
            }
            if facebookPage.Website != "" {
                accountEmbed.Fields = append(accountEmbed.Fields, &discordgo.MessageEmbedField{
                    Name:   "Website",
                    Value:  facebookPage.Website,
                    Inline: true,
                })
            }
            _, _ = session.ChannelMessageSend(msg.ChannelID, fmt.Sprintf("<%s>", fmt.Sprintf(FacebookFriendlyPage, facebookPage.Username)))
            _, err = session.ChannelMessageSendEmbed(msg.ChannelID, accountEmbed)
            helpers.Relax(err)
            return
        }
    } else {
        session.ChannelMessageSend(msg.ChannelID, helpers.GetTextF("bot.arguments.too-few"))
    }
}

func (m *Facebook) lookupFacebookPage(siteName string) (Facebook_Page, error) {
    var facebookPage Facebook_Page
    facebookPageResult, err := fb.Get(fmt.Sprintf("/%s", siteName), fb.Params{
        "fields":       "id,name,about,fan_count,username,is_verified,picture,website",
        "access_token": helpers.GetConfig().Path("facebook.access_token").Data().(string),
    })
    if err != nil {
        return facebookPage, err
    }

    facebookPage.ID = facebookPageResult["id"].(string)
    facebookPage.Name = facebookPageResult["name"].(string)
    facebookPage.About = facebookPageResult["about"].(string)
    facebookPage.Username = facebookPageResult["username"].(string)
    facebookPage.Website = facebookPageResult["website"].(string)

    err = facebookPageResult.DecodeField("picture.data.url", &facebookPage.ProfilePictureUrl)
    if err != nil {
        return facebookPage, err
    }
    err = facebookPageResult.DecodeField("fan_count", &facebookPage.Likes)
    if err != nil {
        return facebookPage, err
    }
    err = facebookPageResult.DecodeField("is_verified", &facebookPage.Verified)
    if err != nil {
        return facebookPage, err
    }

    var facebookPosts []Facebook_Post
    facebookPostsResultData, err := fb.Get(fmt.Sprintf("/%s/posts", siteName), fb.Params{
        "fields":       "id,message,created_time,picture,permalink_url", // TODO: ,child_attachments (better image quality(?))
        "limit":        10,
        "access_token": helpers.GetConfig().Path("facebook.access_token").Data().(string),
    })
    var facebookPostsResult []fb.Result
    err = facebookPostsResultData.DecodeField("data", &facebookPostsResult)
    if err != nil {
        return facebookPage, err
    }

    for _, facebookPostResult := range facebookPostsResult {
        var facebookPost Facebook_Post
        facebookPost.ID = facebookPostResult["id"].(string)
        if _, ok := facebookPostResult["message"]; ok {
            facebookPost.Message = facebookPostResult["message"].(string)
        }
        facebookPost.CreatedAt = facebookPostResult["created_time"].(string)
        facebookPost.Url = facebookPostResult["permalink_url"].(string)
        facebookPost.PictureUrl = ""
        if _, ok := facebookPostResult["picture"]; ok {
            facebookPost.PictureUrl = facebookPostResult["picture"].(string)
        }
        facebookPosts = append(facebookPosts, facebookPost)
    }
    facebookPage.Posts = facebookPosts

    return facebookPage, nil
}

func (m *Facebook) postPostToChannel(channelID string, post Facebook_Post, facebookPage Facebook_Page) {
    facebookNameModifier := ""
    if facebookPage.Verified {
        facebookNameModifier += " :ballot_box_with_check:"
    }

    channelEmbed := &discordgo.MessageEmbed{
        Title:       helpers.GetTextF("plugins.facebook.post-embed-title", facebookPage.Name, facebookPage.Username, facebookNameModifier),
        URL:         post.Url,
        Thumbnail:   &discordgo.MessageEmbedThumbnail{URL: facebookPage.ProfilePictureUrl},
        Footer:      &discordgo.MessageEmbedFooter{Text: helpers.GetText("plugins.facebook.embed-footer")},
        Description: post.Message,
        Color:       helpers.GetDiscordColorFromHex(facebookHexColor),
    }

    if post.PictureUrl != "" {
        channelEmbed.Image = &discordgo.MessageEmbedImage{URL: post.PictureUrl}
    }

    _, _ = cache.GetSession().ChannelMessageSend(channelID, fmt.Sprintf("<%s>", post.Url))
    _, err := cache.GetSession().ChannelMessageSendEmbed(channelID, channelEmbed)
    if err != nil {
        logger.ERROR.L("vlive", fmt.Sprintf("posting post: #%d to channel: #%s failed: %s", post.ID, channelID, err))
    }
}

func (m *Facebook) getEntryBy(key string, id string) DB_Facebook_Page {
    var entryBucket DB_Facebook_Page
    listCursor, err := rethink.Table("facebook").Filter(
        rethink.Row.Field(key).Eq(id),
    ).Run(helpers.GetDB())
    defer listCursor.Close()
    err = listCursor.One(&entryBucket)

    if err == rethink.ErrEmptyResult {
        return entryBucket
    } else if err != nil {
        panic(err)
    }

    return entryBucket
}

func (m *Facebook) getEntryByOrCreateEmpty(key string, id string) DB_Facebook_Page {
    var entryBucket DB_Facebook_Page
    listCursor, err := rethink.Table("facebook").Filter(
        rethink.Row.Field(key).Eq(id),
    ).Run(helpers.GetDB())
    defer listCursor.Close()
    err = listCursor.One(&entryBucket)

    // If user has no DB entries create an empty document
    if err == rethink.ErrEmptyResult {
        insert := rethink.Table("facebook").Insert(DB_Facebook_Page{})
        res, e := insert.RunWrite(helpers.GetDB())
        // If the creation was successful read the document
        if e != nil {
            panic(e)
        } else {
            return m.getEntryByOrCreateEmpty("id", res.GeneratedKeys[0])
        }
    } else if err != nil {
        panic(err)
    }

    return entryBucket
}

func (m *Facebook) setEntry(entry DB_Facebook_Page) {
    _, err := rethink.Table("facebook").Update(entry).Run(helpers.GetDB())
    helpers.Relax(err)
}

func (m *Facebook) deleteEntryById(id string) {
    _, err := rethink.Table("facebook").Filter(
        rethink.Row.Field("id").Eq(id),
    ).Delete().RunWrite(helpers.GetDB())
    helpers.Relax(err)
}