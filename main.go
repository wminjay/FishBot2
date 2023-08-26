package main

import (
	_ "embed"
	"log"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/BaiMeow/msauth"

	"github.com/Tnze/go-mc/bot"
	"github.com/Tnze/go-mc/bot/basic"
	"github.com/Tnze/go-mc/bot/msg"
	"github.com/Tnze/go-mc/bot/playerlist"
	"github.com/Tnze/go-mc/chat"
	"github.com/Tnze/go-mc/data/entity"
	_ "github.com/Tnze/go-mc/data/lang/zh-cn"
	"github.com/Tnze/go-mc/data/packetid"
	pk "github.com/Tnze/go-mc/net/packet"
	"github.com/Tnze/go-mc/yggdrasil"
	"github.com/mattn/go-colorable"
	"github.com/spf13/viper"
)

var (
	c           *bot.Client
	player      *basic.Player
	bobberID    int32
	watch       chan bool
	vp          *viper.Viper
	chatHandler *msg.Manager
	playerList  *playerlist.PlayerList
)

var updateBobber = bot.PacketHandler{
	ID:       packetid.ClientboundSetEntityData,
	Priority: 1,
	F:        checkBobber,
}

var newEntity = bot.PacketHandler{
	ID:       packetid.ClientboundAddEntity,
	Priority: 1,
	F:        newBobber,
}

//go:embed defaultConfig.toml
var defaultConfig []byte

// 定义一个全局变量记录是否抛竿
var isThrow bool = false
var fullHealth float32 = 20

func main() {
	log.SetOutput(colorable.NewColorableStdout())
	log.Println("自动钓鱼机器人")
	log.Println("版本号：mc1.20.1")
	vp = viper.New()
	vp.SetConfigName("config")
	vp.SetConfigType("toml")
	vp.AddConfigPath(".")
	if err := vp.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			os.WriteFile("config.toml", defaultConfig, 0666)
			log.Fatal("配置文件缺失，已创建默认配置文件，请打开\"config.toml\"修改并保存")
		} else {
			log.Fatal(err)
		}
	}
	c = bot.NewClient()
	player = basic.NewPlayer(c, basic.DefaultSettings, basic.EventsListener{
		GameStart:    onGameStart,
		Disconnect:   onDisconnect,
		HealthChange: onHelthChange,
		Death:        onDeath,
		Teleported:   nil})
	switch vp.GetString("profile.login") {
	case "offline":
		c.Auth.Name = vp.GetString("profile.name")
	case "mojang":
		resp, err := yggdrasil.Authenticate(vp.GetString("profile.account"), vp.GetString("profile.passwd"))
		if err != nil {
			log.Fatal("Authenticate:", err)
		}
		log.Println("验证成功")
		c.Auth.UUID, c.Auth.Name = resp.SelectedProfile()
		c.Auth.AsTk = resp.AccessToken()
	case "microsoft":
		msauth.SetClient("67e646fb-20f3-4595-9830-56773a07637d", "")
		msauth.SetRedirectURL("http://127.0.0.1:25595")
		profile, astk, err := msauth.Login()
		log.Println("验证成功")
		if err != nil {
			log.Fatal(err)
		}
		c.Auth.UUID = profile.Id
		c.Auth.Name = profile.Name
		c.Auth.AsTk = astk
	default:
		log.Fatal("无效的登陆模式")
	}
	c.Events.AddListener(updateBobber, newEntity, bot.PacketHandler{
		ID:       packetid.ClientboundSetExperience,
		Priority: 1,
		F:        onExperienceChange,
	})
	playerList = playerlist.New(c)
	chatHandler = msg.New(c, player, playerList, msg.EventsHandler{
		SystemChat:        onSystemChat,
		PlayerChatMessage: onPlayerChat,
		DisguisedChat:     onDisguisedChat,
	})

	addr := net.JoinHostPort(vp.GetString("setting.ip"), strconv.Itoa(vp.GetInt("setting.port")))
	for {
		if err := c.JoinServer(addr); err != nil {
			log.Fatal(err)
		}
		if err := c.HandleGame(); err != nil {
			log.Println(err)
		}
		log.Println("失去与服务器的连接，将在五秒后重连")
		time.Sleep(5 * time.Second)
	}
}

func onGameStart() error {
	log.Println("加入游戏")
	watch = make(chan bool)
	go watchdog()
	time.Sleep(3 * time.Second)
	throw(1)
	return nil
}

func onSystemChat(c chat.Message, overlay bool) error {
	log.Printf("System Chat: %v, Overlay: %v", c, overlay)
	return nil
}

func onPlayerChat(c chat.Message, _ bool) error {
	log.Println("Player Chat:", c)
	return nil
}

func onDisguisedChat(c chat.Message) error {
	log.Println("Disguised Chat:", c)
	return nil
}

func onHelthChange(health float32, food int32, saturation float32) error {
	log.Println("血量:", health, "饱食度:", food, "饱和度:", saturation)
	if health < fullHealth/2 {
		logout()
	}
	return nil
}

func onDeath() error {
	log.Println("死亡")
	logout()
	return nil
}

func onDisconnect(c chat.Message) error {
	log.Println("断开连接:", c)
	return nil
}

func checkBobber(p pk.Packet) error {
	var EID pk.VarInt
	p.Scan(&EID)
	if int32(EID) != bobberID {
		return nil
	}
	var (
		hookedEID pk.VarInt
		catchable pk.Boolean
	)
	p.Scan(&hookedEID, &catchable)
	if catchable {
		throw(2)
		watch <- true
		log.Println("gra~")
	}
	return nil
}

func newBobber(p pk.Packet) error {
	var (
		EID        pk.VarInt
		UUID       pk.UUID
		mobType    pk.VarInt
		x, y, z    pk.Double
		pitch, yaw pk.Angle
		data       pk.Int
	)
	p.Scan(&EID, &UUID, &mobType, &x, &y, &z, &pitch, &yaw, &data)
	//判断是否为浮漂
	if mobType != pk.VarInt(entity.FishingBobber.ID) {
		return nil
	}
	// if data == pk.Int(player.EID) {
	// 	bobberID = int32(EID)
	// }
	// data获取到的int值似乎并不再是实体owner的EID
	// 这里只好把逻辑变成了抛竿之后获取到的第一个浮漂的EID认为是自己的浮漂
	// 但这样在多人并发抛竿场景会判断错误，不过影响有限，暂时先这样
	if isThrow {
		bobberID = int32(EID)
		isThrow = false
	}
	return nil
}

func onExperienceChange(p pk.Packet) error {
	var (
		ExperienceBar   pk.Float
		Level           pk.VarInt
		TotalExperience pk.VarInt
	)
	p.Scan(&ExperienceBar, &Level, &TotalExperience)
	log.Printf("等级: %.2f\n", float32(Level)+float32(ExperienceBar))
	return nil
}

func throw(times int) {
	for ; times > 0; times-- {
		if err := useItem(); err != nil {
			log.Fatal("抛竿:", err)
			return
		}
		if times > 1 {
			time.Sleep(time.Millisecond * 500)
		}
	}
	isThrow = true
}

func watchdog() {
	timeout := time.Second * time.Duration(vp.GetInt("setting.timeout"))
	timer := time.NewTicker(timeout)
	for {
		select {
		case <-timer.C:
			log.Println("WatchDog:超时")
			throw(1)
		case <-watch:
		}
		timer.Reset(timeout)
	}
}

func useItem() error {
	p := pk.Marshal(
		packetid.ServerboundUseItem,
		pk.VarInt(0),
		pk.VarInt(0),
	)
	return c.Conn.WritePacket(p)
}

func logout() {
	//不知道怎么下线，干脆就直接退出
	os.Exit(0)
}
