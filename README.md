<h1 align="center">PiliPili_Frontend</h1>
<p align="center">A program suite for separating the frontend and backend of Emby service playback.</p>

![Commit Activity](https://img.shields.io/github/commit-activity/m/hsuyelin/PiliPili_Frontend/main) ![Top Language](https://img.shields.io/github/languages/top/hsuyelin/PiliPili_Frontend) ![Github License](https://img.shields.io/github/license/hsuyelin/PiliPili_Frontend)


[中文版本](https://github.com/hsuyelin/PiliPili_Frontend/blob/main/READEME_CN.md)

### Introduction

1. This project is a frontend application for separating Emby media service playback into frontend and backend components. It works in conjunction with the playback backend [PiliPili Playback Backend](https://github.com/hsuyelin/PiliPili_Backend).
2. This program is largely based on [YASS-Frontend](https://github.com/FacMata/YASS-Fronted). The original version was implemented in `Python`. To achieve better compatibility, it has been rewritten in `Go` and optimized to enhance usability.

------

### Principles

1. Use a specific `nginx` configuration (refer to [nginx.conf](https://github.com/hsuyelin/PiliPili_Frontend/blob/main/nginx/nginx.conf)) to redirect Emby playback URLs to a designated port.
2. The program listens for requests arriving at the port and extracts the `MediaSourceId` and `ItemId`.
3. Request the corresponding file's relative path (`EmbyPath`) from the Emby service.
4. Generate a signed URL by encrypting the configuration's `Encipher` value with the expiration time (`expireAt`) to create a `signature`.
5. Concatenate the backend playback URL (`backendURL`) with the `EmbyPath` and `signature`.
6. Redirect the playback request to the generated URL for backend handling.

![sequenceDiagram](https://github.com/hsuyelin/PiliPili_Frontend/blob/main/img/sequenceDiagram.png)

------

### Features

- **Compatible with all Emby server versions**.
- **Supports high concurrency**, handling multiple requests simultaneously.
- **Request caching**, enabling fast responses for identical `MediaSourceId` and `ItemId` requests, reducing playback startup time.
- **Link signing**, where the frontend generates and the backend verifies the signature. Mismatched signatures result in a `401 Unauthorized` error.
- **Link expiration**, with an expiration time embedded in the signature to prevent unauthorized usage and continuous theft via packet sniffing.

------

### Configuration File

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

#### Key Settings:

1. **`LogLevel`**: Logging verbosity levels:

	- `WARN`: Minimal logging unless debugging is insufficient.
	- `DEBUG`: Logs `DEBUG`, `INFO`, and `ERROR`. Recommended for debugging.
	- `INFO`: Logs `INFO` and `ERROR`. Adequate for normal operations.
	- `ERROR`: For stable, unattended setups, this minimizes log entries.

2. **`Encipher`**: A 16-character encryption key for obfuscating signatures. **Must match between frontend and backend.**

3. **`Emby` Configuration**:

	- `url`: The Emby server's base URL (e.g., `http://127.0.0.1` for local deployment).
	- `port`: The Emby server port, typically `8096`.
	- `apiKey`: API key for media file path requests.

4. **`Backend` Configuration**:

	- `url`: Remote streaming service's address (e.g., `http://ip:port` for HTTP, or `https://domain.com` for HTTPS on port `443`).

	- ```storageBasePath```: Relative path to map Emby storage paths with backend storage paths. For example:
		- Local `EmbyPath`: `/mnt/anime/OnePiece/Season 22/file.mkv`.
		- If `/mnt` is to be hidden, set `storageBasePath: "/mnt"`.
		- Ensure matching backend configuration ([details here](https://github.com/hsuyelin/PiliPili_Backend)).

5. **`PlayURLMaxAliveTime`**: The maximum lifetime of playback URLs (in seconds). Typically set to `21600` (6 hours) to prevent link misuse.

6. **`Server` Configuration**:

	- `port`: Listening port, default is `60001`.

------

### Usage

#### Step 1: Install Go Environment

##### 1.1 Remove Existing Go Installation

```bash
rm -rf /usr/local/go
```

##### 1.2 Download and Install Latest Go Version

```bash
wget -q -O /tmp/go.tar.gz https://go.dev/dl/go1.23.5.linux-amd64.tar.gz && tar -C /usr/local -xzf /tmp/go.tar.gz && rm /tmp/go.tar.gz
```

##### 1.3 Add Go to Environment Variables

```bash
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc && source ~/.bashrc
```

##### 1.4 Verify Installation

```bash
go version
# Expected output: go version go1.23.5 linux/amd64
```

------

#### Step 2: Clone the Frontend Repository

Clone the repository into a directory, e.g., `/data/emby_fronted`.

```bash
git clone https://github.com/hsuyelin/PiliPili_Frontend.git /data/emby_fronted
```

------

#### Step 3: Configure the Application

Edit the `config.yaml` file in the repository to match your setup.

------

#### Step 4: Run the Application

Run the program in the background:

```bash
nohup go run main.go config.yaml > stream.log 2>&1 &
```
