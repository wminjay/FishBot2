# FishBot2  

![minecraft version](https://img.shields.io/badge/Minecraft-1.20.1-green?style=flat)
[![Go Report Card](https://goreportcard.com/badge/github.com/wminjay/FishBot2)](https://goreportcard.com/report/github.com/wminjay/FishBot2)

Minecraft钓鱼机器人

A bot who fishes in Minecraft server

## 更新说明

最近在重新玩 Minecraft ，想偷懒地钓鱼。但发现之前使用的 FishBot2 不再维护了，版本停留在了1.18.2。于是只能自己动手兼容到了 1.20.1 版本。

### 更新内容

- 更新 FishBot2 支持到 Minecraft 1.20.1 版本。
- 新增功能：
  - 当玩家的血量低于一半时，会自动断开连接，避免装备爆一地。
  - 在获得经验的时候会打印当前等级。
- 逻辑变动：原来是读取浮漂中的owner信息来判断是当前玩家的浮漂，现变更为挥杆后获取到的一个浮漂为玩家的浮漂。

### 注意事项
- 1.18.2-1.20.1当中的任何一个版本的兼容情况都未知。
- 由于逻辑变更，导致多人钓鱼场景，并发挥杆时，有可能因为别人的鱼儿上钩导致收空杆。
- 仅在offline测试过。

老版本见:<https://github.com/MscBaiMeow/FishBot>

an older version see also:<https://github.com/MscBaiMeow/FishBot>

## 使用

双击打开钓鱼机，首次使用会要求填写配置文件

```TOML
# account 是你的登陆账号在offline时不用填写  
# login 是你的登陆模式，可以在 microsoft，mojang，offline中选择
# name 在离线登陆（offline）时必填,在其他登陆模式时会被忽略
# passwd 是你的登陆密码在offline时不用填写  
[profile]
  account = "example@example.com"
  login = "mojang"
  name = "yourid"
  passwd = "password"

# ip 请填写你的服务器ip
# port 一般情况都是25565，少数服务器会使用其他端口
# timeout 是钓鱼等待时间，超过这个时间即使没有钓到鱼也会收杆，一般而言tps20的服务器timeout应该设置为45
[setting]
  ip = "minecraftserver.com"
  port = 25565
  timeout = 45

```

按照要求填写后双击即可启动
