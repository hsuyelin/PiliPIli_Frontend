<h1 align="center">PiliPili_Frontend</h1>
<p align="center">一个简单的Emby服务前后端分离的程序</p>

![Commit Activity](https://img.shields.io/github/commit-activity/m/hsuyelin/PiliPili_Frontend/main) ![Top Language](https://img.shields.io/github/languages/top/hsuyelin/PiliPili_Frontend) ![Github License](https://img.shields.io/github/license/hsuyelin/PiliPili_Frontend)

## 简介

1. 本项目是实现Emby媒体服务播前后端分离的前端程序, 需要与播放分离后端 [PiliPili播放后端](https://github.com/hsuyelin/PiliPili_Backend) 一起配套使用。
2. 本程序大部分参考了 [YASS-Frontend](https://github.com/FacMata/YASS-Frontend)，原版使用了`Python`，为了有更好的兼容性，将其修改为`Go`版本并在其基础上进行了优化，使其更加易用。

------

## 原理

- **通过配合指定的`nginx`配置(可参考[nginx.conf](https://github.com/hsuyelin/PiliPili_Frontend/blob/main/nginx/nginx.conf))，将Emby的播放链接重定向到指定的端口**
- **程序监听该端口过来的请求，获取对应的`MediaSourceId`，`ItemId`**
- **向`Emby`服务请求获取对应的文件相对路径 `EmbyPath`**
- **通过把配置文件中`Encipher`和过期时间`expireAt`进行加密获取签名`signature`**
- **将后端播放的远程地址`backendURL`与之前获取的`EmbyPath`和`signature`做拼接**
- **重定向到拼接后的地址交由后端处理**

![sequenceDiagram](https://github.com/hsuyelin/PiliPili_Frontend/blob/main/img/sequenceDiagram_CN.png)

------

## 功能

- **支持目前所有版本的Emby服务器**
- **支持请求多并发**
- **支持使用`strm`部署的Emby服务端**
- **支持请求缓存，对相同`MediaSourceId`以及`ItemId`的请求可以快速响应，增加起播时间**
- **支持链接签名，由前端签名，后端校验，签名不匹配的会向客户端发送`401`错误**
- **支持链接过期，通过在签名中增加过期时间，防止被恶意抓包导致服务器被持续盗链**

------

## 配置文件

```yaml
# Logging configuration
LogLevel: "INFO" # Log level (e.g., info, debug, warn, error)

# Encryption settings
Encipher: "vPQC5LWCN2CW2opz" # Key used for encryption and obfuscation

# Emby server configuration
Emby:
  url: "http://127.0.0.1" # The base URL for the Emby server
  port: 8096
  apiKey: "6a15d65893024675ba89ffee165f8f1c"  # API key for accessing the Emby server

# Backend streaming configuration
Backend:
    url: "https://streamer.xxxxxxxx.com/stream" # The backend URL for streaming service
    storageBasePath: "/mnt/anime"

# Streaming configuration
PlayURLMaxAliveTime: 21600 # Maximum lifetime of the play URL in seconds (e.g., 6 hours)

# Server configuration
Server:
  port: 60001

# Special medias configuration
SpecialMedias:
   - key: "MediaMissing"
     name: "Default media for missing cases"
     mediaPath: "specialMedia/mediaMissing"
     itemId: "mediaMissing-item-id"
     mediaSourceID: "mediaMissing-media-source-id"
   - key: "September18"
     name: "September 18 - Commemorative Media"
     mediaPath: "specialMedia/september18"
     itemId: "september18-item-id"
     mediaSourceID: "september18-media-source-id"
   - key: "October1"
     name: "October 1 - National Day Media"
     mediaPath: "specialMedia/october1"
     itemId: "october1-item-id"
     mediaSourceID: "october1-media-source-id"
   - key: "December13"
     name: "December 13 - Nanjing Massacre Commemoration"
     mediaPath: "specialMedia/december13"
     itemId: "december13-item-id"
     mediaSourceID: "december13-media-source-id"
   - key: "ChineseNewYearEve"
     name: "Chinese New Year's Eve Media"
     mediaPath: "specialMedia/chinesenewyeareve"
     itemId: "chinesenewyeareve-item-id"
     mediaSourceID: "chinesenewyeareve-media-source-id"
```

* LogLevel：打印日志的等级
	* `WARN`：会显示所有的日志，除非开启`DEBUG`后也没办法满足需求，一般不建议使用这个等级的日志
	* `DEBUG`：会显示`DEBUG`/`INFO`/`ERROR`等级的日志，如果需要调试尽量使用这个等级的
	* `INFO`：显示`INFO`/`EROR`的日志，正常情况下使用这个等级可以满足需求
	* `ERROR`：如果接入后足够稳定，已经达到无人值守的阶段，可以使用这个等级，降低日志数量
* Encipher：加密因子，格式是`16`位长度的字符串，用于混淆签名，`前端和后端必须保持一致`
* Emby:
	* url: Emby服务部署的地址，如果前端程序和Emby服务在一台机器上，可以使用`http://127.0.0.1`
	* port: Emby服务部署的端口，一般是`8096`，按需设置
	* apikey：Emby服务的`APIKey`，用于向Emby服务获取媒体文件地址
* Backend：
	* url：远程推流的地址
		* 如果是`http`必须要要加端口号，例如：`http://ip:port`
		* 如果是`https`且是`443`端口无需加端口号，例如：`https://streamer.xxxxxxxx.com/stream`
	* storageBasePath：
		* 前提：需要前端映射到Emby服务中存储路径和后端实际存储文件路径一致
		* 需要隐藏的目录相对于远程挂载目录的相对路径，例如：你本地获取的`EmbyPath`为`/mnt/anime/动漫/海贼王 (1999)/Season 22/37854 S22E1089 2160p.B-Global.mkv`，但是你想隐藏`/mnt`这个路径，你就在前端的`storageBasePath`中填写`/mnt`，相对的你需要在 [后端程序](https://github.com/hsuyelin/PiliPili_Backend) 配置的`StorageBasePath`填写`/mnt`
		* 也就是说你想隐藏哪部分路径，那么哪部分路径就是在后端中填写的
* PlayURLMaxAliveTime：播放链接的过期时间，单位是秒，一般是6小时（设置21600）就足够了，主要防止恶意抓包，导致链接一致可以被观看或者下载
* Server：
	* port: 需要监听的端口号，如果没有特殊需要，直接默认`60001`就可以了
* SpecialMedias: 用来重定向一些特殊意义的媒体，比如中国传统节日新年等，目前支持的特殊意义媒体如下（没有这个需求，设置成空就行）：
  * MediaMissing: 服务器文件丢失，显示默认的媒体文件
  * September18: 中国的“九一八事变”纪念日，对中国人很有意义，勿忘国耻，砥砺前行，珍惜和平
  * October1：10月1日，中国的国庆节
  * December13: 中国的“国家公祭日”纪念日，对中国人很有意义，勿忘国耻，砥砺前行，珍惜和平

------

## 如何使用

### 1. Docker安装(推荐)

#### 1.1 创建docker文件夹

```shell
mkdir -p /data/docker/pilipili_frontend
```

#### 1.2 创建配置文件夹和配置文件

```shell
cd /data/docker/pilipili_frontend
mkdir -p config && cd config
```

将 [config.yaml](https://github.com/hsuyelin/PiliPili_Frontend/blob/main/config.yaml) 复制到`config`文件夹中，并进行编辑

#### 1.3 创建docker-compose.yaml

返回到 `/data/docker/pilipili_frontend`目录，将 [docker-compose.yml](https://github.com/hsuyelin/PiliPili_Frontend/blob/main/docker/docker-compose.yml) 复制到该目录下

#### 1.4 启动容器

```shell
docker-compose pull && docker-compose up -d
```

### 2. 手动安装

#### 2.1 安装Go环境

##### 2.1.1 卸载本机的Go程序

强制删除本机安装的go，为防止`go`版本不匹配

```shell
rm -rf /usr/local/go
```

##### 2.1.2 下载并安装最新版本的Go程序

```shell
wget -q -O /tmp/go.tar.gz https://go.dev/dl/go1.23.5.linux-amd64.tar.gz && tar -C /usr/local -xzf /tmp/go.tar.gz && rm /tmp/go.tar.gz
```

##### 2.1.3 将Go程序写入环境变量

```shell
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc && source ~/.bashrc
```

##### 2.1.4 验证是否安装成功

```shell
go version #显示 go version go1.23.5 linux/amd64 就是安装成功
```

#### 2.2 克隆前端程序组到本地

假如你需要克隆到`/data/emby_fronted`这个目录

```shell
git clone https://github.com/hsuyelin/PiliPili_Frontend.git /data/emby_fronted
```

#### 2.3 进入前端程序目录编辑配置文件

```yaml
# Logging configuration
LogLevel: "INFO" # Log level (e.g., info, debug, warn, error)

# Encryption settings
Encipher: "vPQC5LWCN2CW2opz" # Key used for encryption and obfuscation

# Emby server configuration
Emby:
  url: "http://127.0.0.1" # The base URL for the Emby server
  port: 8096
  apiKey: "6a15d65893024675ba89ffee165f8f1c"  # API key for accessing the Emby server

# Backend streaming configuration
Backend:
    url: "https://streamer.xxxxxxxx.com/stream" # The backend URL for streaming service
    storageBasePath: "/mnt/anime"

# Streaming configuration
PlayURLMaxAliveTime: 21600 # Maximum lifetime of the play URL in seconds (e.g., 6 hours)

# Server configuration
Server:
  port: 60001
```

#### 2.4 运行程序

```shell
nohup go run main.go config.yaml > stream.log 2>&1 &
```
