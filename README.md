<h1 align="center">PiliPili_Frontend</h1>
<p align="center">A program suite for separating the frontend and backend of Emby service playback.</p>

![Commit Activity](https://img.shields.io/github/commit-activity/m/hsuyelin/PiliPili_Frontend/main) ![Top Language](https://img.shields.io/github/languages/top/hsuyelin/PiliPili_Frontend) ![Github License](https://img.shields.io/github/license/hsuyelin/PiliPili_Frontend)


[中文版本](https://github.com/hsuyelin/PiliPili_Frontend/blob/main/README_CN.md)

## Introduction

1. This project is a frontend application for separating Emby media service playback into frontend and backend components. It works in conjunction with the playback backend [PiliPili Playback Backend](https://github.com/hsuyelin/PiliPili_Backend).
2. This program is largely based on [YASS-Frontend](https://github.com/FacMata/YASS-Frontend). The original version was implemented in `Python`. To achieve better compatibility, it has been rewritten in `Go` and optimized to enhance usability.

------

## Principles

1. Use a specific `nginx` configuration (refer to [nginx.conf](https://github.com/hsuyelin/PiliPili_Frontend/blob/main/nginx/nginx.conf)) to redirect Emby playback URLs to a designated port.
2. The program listens for requests arriving at the port and extracts the `MediaSourceId` and `ItemId`.
3. Request the corresponding file's relative path (`EmbyPath`) from the Emby service.
4. Generate a signed URL by encrypting the configuration's `Encipher` value with the expiration time (`expireAt`) to create a `signature`.
5. Concatenate the backend playback URL (`backendURL`) with the `EmbyPath` and `signature`.
6. Redirect the playback request to the generated URL for backend handling.

![sequenceDiagram](https://github.com/hsuyelin/PiliPili_Frontend/blob/main/img/sequenceDiagram.png)

------

## Features

- **Compatible with all Emby server versions**.
- **Supports high concurrency**, handling multiple requests simultaneously.
- **Support for deploying Emby server with `strm`.**
- **Request caching**, enabling fast responses for identical `MediaSourceId` and `ItemId` requests, reducing playback startup time.
- **Link signing**, where the frontend generates and the backend verifies the signature. Mismatched signatures result in a `401 Unauthorized` error.
- **Link expiration**, with an expiration time embedded in the signature to prevent unauthorized usage and continuous theft via packet sniffing.

------

## Configuration File

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

- **LogLevel**: The log level for printing logs
	- `WARN`: Displays all logs unless enabling `DEBUG` does not meet the needs; generally, this log level is not recommended.
	- `DEBUG`: Displays logs at the `DEBUG`/`INFO`/`ERROR` levels; if debugging is needed, it is recommended to use this level.
	- `INFO`: Displays logs at the `INFO`/`ERROR` levels; this level is usually sufficient under normal circumstances.
	- `ERROR`: If the system is stable enough and has reached an unattended stage, this level can be used to reduce the number of logs.

- **Encipher**: The encryption factor, which is a 16-character string used for signature obfuscation. **The frontend and backend must remain consistent**.

- **Emby**:
	- **url**: The address where the Emby service is deployed. If the frontend application and the Emby service are on the same machine, `http://127.0.0.1` can be used.
	- **port**: The port where the Emby service is deployed, usually `8096`. Configure as needed.
	- **apikey**: The `APIKey` for the Emby service, used to retrieve media file URLs from the Emby service.

- **Backend**:
	- **url**: The URL for remote streaming.
		- If using `http`, the port number must be included, e.g., `http://ip:port`.
		- If using `https` on port `443`, the port number can be omitted, e.g., `https://streamer.xxxxxxxx.com/stream`.
	- **storageBasePath**:
		- **Prerequisite**: The frontend needs to map the storage path in the Emby service to the actual storage file path on the backend.
		- The relative path of the directory to be hidden, relative to the remote mounted directory. For example: If the local `EmbyPath` is `/mnt/anime/动漫/海贼王 (1999)/Season 22/37854 S22E1089 2160p.B-Global.mkv`, but you want to hide the `/mnt` part, enter `/mnt` in the frontend's `storageBasePath`. Correspondingly, in the [backend configuration](https://github.com/hsuyelin/PiliPili_Backend), set `StorageBasePath` to `/mnt`.
		- In other words, the part of the path you want to hide must be configured in the backend.

- **PlayURLMaxAliveTime**: The expiration time for playback links, in seconds. Typically, 6 hours (set to `21600`) is sufficient to prevent malicious packet capturing, which could otherwise allow the same link to be watched or downloaded indefinitely.

- **Server**:
	- **port**: The port to be listened on. If there are no special requirements, the default value `60001` can be used.

- **SpecialMedias**: Used to redirect media with special significance, such as content related to Chinese traditional holidays or historical events. Currently supported events include (There's no need for that. Just set it to null.):
	- **MediaMissing**: Redirects to a default media file if the server file is missing.
	- **September18**: Commemorates the "Mukden Incident" of September 18, a significant historical date for China, promoting remembrance of history, peace, and perseverance.
	- **October1**：Celebrates October 1, China's National Day.
	- **December13**: Commemorates China's National Memorial Day on December 13, urging remembrance of history, peace, and perseverance.

------

## How to Use

### 1. Install Using Docker (Recommended)

#### 1.1 Create a Docker Directory

```shell
mkdir -p /data/docker/pilipili_frontend
```

#### 1.2 Create Configuration Folder and File

```shell
cd /data/docker/pilipili_frontend
mkdir -p config && cd config
```

Copy [config.yaml](https://github.com/hsuyelin/PiliPili_Frontend/blob/main/config.yaml) to the `config` folder and edit it as needed.

#### 1.3 Create docker-compose.yaml

Navigate back to the `/data/docker/pilipili_frontend` directory, and copy [docker-compose.yml](https://github.com/hsuyelin/PiliPili_Frontend/blob/main/docker/docker-compose.yml) to this directory.

#### 1.4 Start the Container

```shell
docker-compose pull && docker-compose up -d
```

### 2. Manual Installation

#### 2.1: Install Go Environment

##### 2.1.1 Remove Existing Go Installation

```bash
rm -rf /usr/local/go
```

##### 2.1.2 Download and Install Latest Go Version

```bash
wget -q -O /tmp/go.tar.gz https://go.dev/dl/go1.23.5.linux-amd64.tar.gz && tar -C /usr/local -xzf /tmp/go.tar.gz && rm /tmp/go.tar.gz
```

##### 2.1.3 Add Go to Environment Variables

```bash
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc && source ~/.bashrc
```

##### 2.1.4 Verify Installation

```bash
go version
# Expected output: go version go1.23.5 linux/amd64
```

------

#### 2.2: Clone the Frontend Repository

Clone the repository into a directory, e.g., `/data/emby_fronted`.

```bash
git clone https://github.com/hsuyelin/PiliPili_Frontend.git /data/emby_fronted
```

------

#### 2.3: Configure the Application

Edit the `config.yaml` file in the repository to match your setup.

------

#### 2.4: Run the Application

Run the program in the background:

```bash
nohup go run main.go config.yaml > stream.log 2>&1 &
```
