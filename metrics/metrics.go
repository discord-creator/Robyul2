package metrics

import (
    "net/http"
    "expvar"
    "github.com/bwmarrin/discordgo"
    "time"
    "github.com/sn0w/Karen/logger"
    "runtime"
)

var (
    MessagesReceived = expvar.NewInt("messages_received")
    UserCount = expvar.NewInt("user_count")
    ChannelCount = expvar.NewInt("channel_count")
    GuildCount = expvar.NewInt("guild_count")
    CommandsExecuted = expvar.NewInt("commands_executed")
    CleverbotRequests = expvar.NewInt("cleverbot_requests")
    GoroutineCount = expvar.NewInt("coroutine_count")
)

func Init() {
    logger.INF("[METRICS] Listening on http://[::1]:1337")
    go http.ListenAndServe(":1337", nil)
}

func OnReady(session *discordgo.Session, event *discordgo.Ready) {
    go CollectDiscordMetrics(session)
    go CollectRuntimeMetrics()
}

func OnMessageCreate(session *discordgo.Session, event *discordgo.MessageCreate) {
    MessagesReceived.Add(1)
}

func CollectDiscordMetrics(session *discordgo.Session) {
    for {
        time.Sleep(15 * time.Second)

        users := make(map[string]string)
        channels := 0
        guilds := session.State.Guilds

        for _, guild := range guilds {
            channels += len(guild.Channels)

            for _, u := range guild.Members {
                users[u.User.ID] = u.User.Username
            }
        }

        UserCount.Set(int64(len(users)))
        ChannelCount.Set(int64(channels))
        GuildCount.Set(int64(len(guilds)))
    }
}

func CollectRuntimeMetrics() {
    for {
        time.Sleep(15 * time.Second)
        GoroutineCount.Set(runtime.NumGoroutine())
    }
}
