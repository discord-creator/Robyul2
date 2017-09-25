package helpers

import (
	"errors"
	"fmt"
	"math/big"
	"regexp"
	"strconv"
	"strings"
	"time"

	"encoding/json"

	"github.com/Seklfreak/Robyul2/cache"
	"github.com/bwmarrin/discordgo"
	"github.com/getsentry/raven-go"
	redisCache "github.com/go-redis/cache"
)

const (
	DISCORD_EPOCH int64 = 1420070400000
)

var botAdmins = []string{
	"116620585638821891", // Sekl
	"134298438559858688", // Kakkela
}
var NukeMods = []string{
	"116620585638821891", // Sekl
	"134298438559858688", // Kakkela
	"68661361537712128",  // Berk
}
var RobyulMod = []string{
	"132633380628987904", // sunny
}
var Blacklisted = []string{
	"171883318386753536", // ForRyu
}
var ExtendedInspectRoleIDs = []string{
	"345209385821274113", // inspect extended (sekl's dev cord)
	"345209098100277248", // inspect (Moderator Chat)
}
var adminRoleNames = []string{"Admin", "Admins", "ADMIN", "School Board", "admin", "admins"}
var modRoleNames = []string{"Mod", "Mods", "Mod Trainee", "Moderator", "Moderators", "MOD", "Minimod", "Guard", "Janitor", "mod", "mods"}

func IsBlacklisted(id string) bool {
	for _, s := range Blacklisted {
		if s == id {
			return true
		}
	}

	return false
}

func IsNukeMod(id string) bool {
	for _, s := range NukeMods {
		if s == id {
			return true
		}
	}

	return false
}

// IsBotAdmin checks if $id is in $botAdmins
func IsBotAdmin(id string) bool {
	for _, s := range botAdmins {
		if s == id {
			return true
		}
	}

	return false
}

func IsRobyulMod(id string) bool {
	if IsBotAdmin(id) {
		return true
	}
	for _, s := range RobyulMod {
		if s == id {
			return true
		}
	}

	return false
}

func CanInspectExtended(msg *discordgo.Message) bool {
	if IsBotAdmin(msg.Author.ID) {
		return true
	}

	if IsRobyulMod(msg.Author.ID) {
		return true
	}

	if IsNukeMod(msg.Author.ID) {
		return true
	}

	channel, e := GetChannel(msg.ChannelID)
	if e != nil {
		return false
	}

	guild, e := GetGuild(channel.GuildID)
	if e != nil {
		return false
	}

	guildMember, e := GetGuildMember(guild.ID, msg.Author.ID)
	if e != nil {
		return false
	}
	for _, role := range guild.Roles {
		for _, userRole := range guildMember.Roles {
			if userRole == role.ID {
				for _, inspectRoleID := range ExtendedInspectRoleIDs {
					if role.ID == inspectRoleID {
						return true
					}
				}
			}
		}
	}
	return false
}

func IsAdmin(msg *discordgo.Message) bool {
	channel, e := GetChannel(msg.ChannelID)
	if e != nil {
		return false
	}

	guild, e := GetGuild(channel.GuildID)
	if e != nil {
		return false
	}

	if msg.Author.ID == guild.OwnerID || IsBotAdmin(msg.Author.ID) {
		return true
	}

	guildMember, e := GetGuildMember(guild.ID, msg.Author.ID)
	if e != nil {
		return false
	}
	// Check if role may manage server or a role is in admin role list
	for _, role := range guild.Roles {
		for _, userRole := range guildMember.Roles {
			if userRole == role.ID {
				if role.Permissions&discordgo.PermissionAdministrator == discordgo.PermissionAdministrator {
					return true
				}
				for _, adminRoleName := range adminRoleNames {
					if role.Name == adminRoleName {
						return true
					}
				}
			}
		}
	}
	return false
}

func IsAdminByID(guildID string, userID string) bool {
	guild, e := GetGuild(guildID)
	if e != nil {
		return false
	}

	if userID == guild.OwnerID || IsBotAdmin(userID) {
		return true
	}

	guildMember, e := GetGuildMember(guild.ID, userID)
	if e != nil {
		return false
	}
	// Check if role may manage server or a role is in admin role list
	for _, role := range guild.Roles {
		for _, userRole := range guildMember.Roles {
			if userRole == role.ID {
				if role.Permissions&discordgo.PermissionAdministrator == discordgo.PermissionAdministrator {
					return true
				}
				for _, adminRoleName := range adminRoleNames {
					if role.Name == adminRoleName {
						return true
					}
				}
			}
		}
	}
	return false
}

func HasPermissionByID(guildID string, userID string, permission int) bool {
	guild, e := GetGuild(guildID)
	if e != nil {
		return false
	}

	if userID == guild.OwnerID {
		return true
	}

	guildMember, e := GetGuildMember(guild.ID, userID)
	if e != nil {
		return false
	}
	for _, role := range guild.Roles {
		for _, userRole := range guildMember.Roles {
			if userRole == role.ID {
				if role.Permissions&permission == permission {
					return true
				}
			}
		}
	}
	return false
}

func IsMod(msg *discordgo.Message) bool {
	if IsAdmin(msg) == true {
		return true
	} else {
		channel, e := GetChannel(msg.ChannelID)
		if e != nil {
			return false
		}
		guild, e := GetGuild(channel.GuildID)
		if e != nil {
			return false
		}
		guildMember, e := GetGuildMember(guild.ID, msg.Author.ID)
		if e != nil {
			return false
		}
		// check if a role is in mod role list
		for _, role := range guild.Roles {
			for _, userRole := range guildMember.Roles {
				if userRole == role.ID {
					for _, modRoleName := range modRoleNames {
						if role.Name == modRoleName {
							return true
						}
					}
				}
			}
		}
	}
	return false
}

func IsModByID(guildID string, userID string) bool {
	if IsAdminByID(guildID, userID) == true {
		return true
	} else {
		guild, e := GetGuild(guildID)
		if e != nil {
			return false
		}
		guildMember, e := GetGuildMember(guild.ID, userID)
		if e != nil {
			return false
		}
		// check if a role is in mod role list
		for _, role := range guild.Roles {
			for _, userRole := range guildMember.Roles {
				if userRole == role.ID {
					for _, modRoleName := range modRoleNames {
						if role.Name == modRoleName {
							return true
						}
					}
				}
			}
		}
	}
	return false
}

// RequireAdmin only calls $cb if the author is an admin or has MANAGE_SERVER permission
func RequireAdmin(msg *discordgo.Message, cb Callback) {
	if !IsAdmin(msg) {
		cache.GetSession().ChannelMessageSend(msg.ChannelID, GetText("admin.no_permission"))
		return
	}

	cb()
}

// RequireAdmin only calls $cb if the author is an admin or has MANAGE_SERVER permission
func RequireMod(msg *discordgo.Message, cb Callback) {
	if !IsMod(msg) {
		cache.GetSession().ChannelMessageSend(msg.ChannelID, GetText("mod.no_permission"))
		return
	}

	cb()
}

// RequireBotAdmin only calls $cb if the author is a bot admin
func RequireBotAdmin(msg *discordgo.Message, cb Callback) {
	if !IsBotAdmin(msg.Author.ID) {
		cache.GetSession().ChannelMessageSend(msg.ChannelID, GetText("botadmin.no_permission"))
		return
	}

	cb()
}

// RequireSupportMod only calls $cb if the author is a support mod
func RequireRobyulMod(msg *discordgo.Message, cb Callback) {
	if !IsRobyulMod(msg.Author.ID) {
		cache.GetSession().ChannelMessageSend(msg.ChannelID, GetText("robyulmod.no_permission"))
		return
	}

	cb()
}

func ConfirmEmbed(channelID string, author *discordgo.User, confirmMessageText string, confirmEmojiID string, abortEmojiID string) bool {
	// send embed asking the user to confirm
	confirmMessage, err := cache.GetSession().ChannelMessageSendComplex(channelID,
		&discordgo.MessageSend{
			Content: "<@" + author.ID + ">",
			Embed: &discordgo.MessageEmbed{
				Title:       GetText("bot.embeds.please-confirm-title"),
				Description: confirmMessageText,
			},
		})
	if err != nil {
		cache.GetSession().ChannelMessageSend(channelID, GetTextF("bot.errors.general", err.Error()))
		return false
	}
	if len(confirmMessage.Embeds) <= 0 {
		cache.GetSession().ChannelMessageSend(channelID, GetText("bot.errors.no-embed"))
		return false
	}

	// add default reactions to embed
	cache.GetSession().MessageReactionAdd(confirmMessage.ChannelID, confirmMessage.ID, confirmEmojiID)
	cache.GetSession().MessageReactionAdd(confirmMessage.ChannelID, confirmMessage.ID, abortEmojiID)

	// check every second if a reaction has been clicked
	for {
		confirmes, _ := cache.GetSession().MessageReactions(confirmMessage.ChannelID, confirmMessage.ID, confirmEmojiID, 100)
		for _, confirm := range confirmes {
			if confirm.ID == author.ID {
				cache.GetSession().ChannelMessageDelete(confirmMessage.ChannelID, confirmMessage.ID)
				// user has confirmed the call
				return true
			}
		}
		aborts, _ := cache.GetSession().MessageReactions(confirmMessage.ChannelID, confirmMessage.ID, abortEmojiID, 100)
		for _, abort := range aborts {
			if abort.ID == author.ID {
				cache.GetSession().ChannelMessageDelete(confirmMessage.ChannelID, confirmMessage.ID)
				// User has aborted the call
				return false
			}
		}

		time.Sleep(1 * time.Second)
	}
}

func GetMuteRole(guildID string) (*discordgo.Role, error) {
	guild, err := GetGuild(guildID)
	Relax(err)
	var muteRole *discordgo.Role
	settings, err := GuildSettingsGet(guildID)
	for _, role := range guild.Roles {
		Relax(err)
		if role.Name == settings.MutedRoleName {
			muteRole = role
		}
	}
	if muteRole == nil {
		muteRole, err = cache.GetSession().GuildRoleCreate(guildID)
		if err != nil {
			return muteRole, err
		}
		muteRole, err = cache.GetSession().GuildRoleEdit(guildID, muteRole.ID, settings.MutedRoleName, muteRole.Color, muteRole.Hoist, 0, muteRole.Mentionable)
		if err != nil {
			return muteRole, err
		}
		for _, channel := range guild.Channels {
			err = cache.GetSession().ChannelPermissionSet(channel.ID, muteRole.ID, "role", 0, discordgo.PermissionSendMessages)
			if err != nil {
				cache.GetLogger().WithField("module", "discord").Error("Error disabling send messages on mute Role: " + err.Error())
			}
			// TODO: update discordgo
			//err = cache.GetSession().ChannelPermissionSet(channel.ID, muteRole.ID, "role", 0, discordgo.PermissionAddReactions)
			//if err != nil {
			//    logger.ERROR.L("discord", "Error disabling add reactions on mute Role: " + err.Error())
			//}
		}
	}
	return muteRole, nil
}

func GetGuildMember(guildID string, userID string) (*discordgo.Member, error) {
	targetMember, err := cache.GetSession().State.Member(guildID, userID)
	if targetMember == nil || targetMember.GuildID == "" || targetMember.JoinedAt == "" {
		cache.GetLogger().WithField("module", "discord").WithField("method", "GetGuildMember").Debug(
			fmt.Sprintf("api request: GuildMember: %s, %s", guildID, userID))
		targetMember, err = cache.GetSession().GuildMember(guildID, userID)
	}
	return targetMember, err
}

func GetGuildMemberWithoutApi(guildID string, userID string) (*discordgo.Member, error) {
	return cache.GetSession().State.Member(guildID, userID)
}

func GetIsInGuild(guildID string, userID string) bool {
	member, err := GetGuildMemberWithoutApi(guildID, userID)
	if err == nil && member != nil && member.User != nil && member.User.ID != "" {
		return true
	} else {
		return false
	}
}

func GetGuild(guildID string) (*discordgo.Guild, error) {
	targetGuild, err := cache.GetSession().State.Guild(guildID)
	if targetGuild == nil || targetGuild.ID == "" {
		cache.GetLogger().WithField("module", "discord").WithField("method", "GetGuild").Debug(
			fmt.Sprintf("api request: Guild: %s", guildID))
		targetGuild, err = cache.GetSession().Guild(guildID)
	}
	return targetGuild, err
}

func GetChannel(channelID string) (*discordgo.Channel, error) {
	targetChannel, err := cache.GetSession().State.Channel(channelID)
	if targetChannel == nil || targetChannel.ID == "" {
		cache.GetLogger().WithField("module", "discord").WithField("method", "GetChannel").Debug(
			fmt.Sprintf("api request: Channel: %s", channelID))
		targetChannel, err = cache.GetSession().Channel(channelID)
	}
	return targetChannel, err
}

func GetMessage(channelID string, messageID string) (*discordgo.Message, error) {
	targetMessage, err := cache.GetSession().State.Message(channelID, messageID)
	if targetMessage == nil || targetMessage.ID == "" {
		cache.GetLogger().WithField("module", "discord").WithField("method", "GetMessage").Debug(
			fmt.Sprintf("api request: Message: %s in Channel: %s", messageID, channelID))
		targetMessage, err = cache.GetSession().ChannelMessage(channelID, messageID)
		cache.GetSession().State.MessageAdd(targetMessage)
		return targetMessage, err
	}
	return targetMessage, nil
}

func GetChannelFromMention(msg *discordgo.Message, mention string) (*discordgo.Channel, error) {
	var targetChannel *discordgo.Channel
	re := regexp.MustCompile("(<#)?(\\d+)(>)?")
	result := re.FindStringSubmatch(mention)
	if len(result) == 4 {
		sourceChannel, err := GetChannel(msg.ChannelID)
		if err != nil {
			return targetChannel, err
		}
		if sourceChannel == nil {
			return targetChannel, errors.New("Channel not found.")
		}
		targetChannel, err := GetChannel(result[2])
		if err != nil {
			return targetChannel, err
		}
		if targetChannel.Type != discordgo.ChannelTypeGuildText {
			return targetChannel, errors.New("not a text channel")
		}
		if sourceChannel.GuildID != targetChannel.GuildID {
			return targetChannel, errors.New("Channel on different guild.")
		}
		return targetChannel, err
	} else {
		return targetChannel, errors.New("Channel not found.")
	}
}

func GetGlobalChannelFromMention(mention string) (*discordgo.Channel, error) {
	var targetChannel *discordgo.Channel
	re := regexp.MustCompile("(<#)?(\\d+)(>)?")
	result := re.FindStringSubmatch(mention)
	if len(result) == 4 {
		targetChannel, err := GetChannel(result[2])
		if err != nil {
			return targetChannel, err
		}
		return targetChannel, err
	} else {
		return targetChannel, errors.New("Channel not found.")
	}
}

func GetUser(userID string) (*discordgo.User, error) {
	var err error
	var targetUser discordgo.User
	cacheCodec := cache.GetRedisCacheCodec()
	key := fmt.Sprintf("robyul2-discord:api:user:%s", userID) // TOOD: Should we cache this?

	for _, guild := range cache.GetSession().State.Guilds {
		member, err := GetGuildMemberWithoutApi(guild.ID, userID)
		if err == nil && member != nil && member.User != nil && member.User.ID != "" {
			return member.User, nil
		}
	}

	if err = cacheCodec.Get(key, &targetUser); err != nil {
		cache.GetLogger().WithField("module", "discord").WithField("method", "GetUser").Debug(
			fmt.Sprintf("api request: User: %s", userID))
		targetUser, err := cache.GetSession().User(userID)
		if err == nil {
			err = cacheCodec.Set(&redisCache.Item{
				Key:        key,
				Object:     targetUser,
				Expiration: time.Minute * 10,
			})
			if err != nil {
				raven.CaptureError(fmt.Errorf("%#v", err), map[string]string{})
			}
		}
		return targetUser, err
	}
	return &targetUser, err
}

func GetUserFromMention(mention string) (*discordgo.User, error) {
	re := regexp.MustCompile("(<@)?(\\d+)(>)?")
	result := re.FindStringSubmatch(mention)
	if len(result) == 4 {
		return GetUser(result[2])
	} else {
		return &discordgo.User{}, errors.New("User not found.")
	}
}

func GetDiscordColorFromHex(hex string) int {
	colorInt, ok := new(big.Int).SetString(strings.Replace(hex, "#", "", 1), 16)
	if ok == true {
		return int(colorInt.Int64())
	} else {
		return 0x0FADED
	}
}

func GetTimeFromSnowflake(id string) time.Time {
	iid, err := strconv.ParseInt(id, 10, 64)
	Relax(err)

	return time.Unix(((iid>>22)+DISCORD_EPOCH)/1000, 0).UTC()
}

func GetAllPermissions(guild *discordgo.Guild, member *discordgo.Member) int64 {
	var perms int64 = 0
	for _, x := range guild.Roles {
		if x.Name == "@everyone" {
			perms |= int64(x.Permissions)
		}
	}
	for _, r := range member.Roles {
		for _, x := range guild.Roles {
			if x.ID == r {
				perms |= int64(x.Permissions)
			}
		}
	}
	return perms
}
func Pagify(text string, delimiter string) []string {
	result := make([]string, 0)
	textParts := strings.Split(text, delimiter)
	currentOutputPart := ""
	for _, textPart := range textParts {
		if len(currentOutputPart)+len(textPart)+len(delimiter) <= 1992 {
			if len(currentOutputPart) > 0 || len(result) > 0 {
				currentOutputPart += delimiter + textPart
			} else {
				currentOutputPart += textPart
			}
		} else {
			result = append(result, currentOutputPart)
			currentOutputPart = ""
			if len(textPart) <= 1992 { // @TODO: else: split text somehow
				currentOutputPart = textPart
			}
		}
	}
	if currentOutputPart != "" {
		result = append(result, currentOutputPart)
	}
	return result
}

func GetAvatarUrl(user *discordgo.User) string {
	return GetAvatarUrlWithSize(user, 1024)
}

func GetAvatarUrlWithSize(user *discordgo.User, size uint16) string {
	if user.Avatar == "" {
		return ""
	}

	avatarUrl := "https://cdn.discordapp.com/avatars/%s/%s.%s?size=%d"

	if strings.HasPrefix(user.Avatar, "a_") {
		return fmt.Sprintf(avatarUrl, user.ID, user.Avatar, "gif", size)
	}

	return fmt.Sprintf(avatarUrl, user.ID, user.Avatar, "jpg", size)
}

func CommandExists(name string) bool {
	for _, command := range cache.GetPluginList() {
		if command == strings.ToLower(name) {
			return true
		}
	}
	for _, command := range cache.GetPluginExtendedList() {
		if command == strings.ToLower(name) {
			return true
		}
	}
	for _, command := range cache.GetTriggerPluginList() {
		if command == strings.ToLower(name) {
			return true
		}
	}
	return false
}

func WebhookExecuteWithResult(webhookID, token string, data *discordgo.WebhookParams) (message *discordgo.Message, err error) {
	uri := discordgo.EndpointWebhookToken(webhookID, token) + "?wait=true"

	result, err := cache.GetSession().RequestWithBucketID("POST", uri, data, discordgo.EndpointWebhookToken("", ""))
	if err != nil {
		return message, err
	}

	err = json.Unmarshal(result, &message)
	return message, err
}
