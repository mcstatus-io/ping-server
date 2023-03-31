# Ping Server
![](https://img.shields.io/github/languages/code-size/mcstatus-io/ping-server)
![](https://img.shields.io/github/issues/mcstatus-io/ping-server)
![](https://img.shields.io/github/actions/workflow/status/mcstatus-io/ping-server/go.yml)
![](https://img.shields.io/uptimerobot/ratio/m792379047-190f2d39d31ecafa9b1f86ab)

The status retrieval/ping server that powers the API for mcstatus.io. This repository is open source to allow developers to run their own Minecraft server status API server.

If you do not know what you are doing, or think that the cache durations enforced on our official website are tolerable, I would highly recommend using the official API instead. It is much more reliable and reduces the complexity of hosting it yourself.

## Official API Documentation

https://mcstatus.io/docs

## Requirements

- [Git](https://git-scm.com/)
- [Go](https://go.dev/)
- [Redis](https://redis.io/)
- [GNU Make](https://www.gnu.org/software/make/) (*optional*)

## Installation

1. Clone the repository to a folder
    - `git clone https://github.com/mcstatus-io/ping-server.git`
2. Move the working directory into the folder
    - `cd ping-server`
3. Install all required dependencies
    - `go get ...`
4. Build the executable
    - Using GNU make
        - `make`
    - Without GNU make
        - `go build -o .\bin\main.exe .\src\*.go` (Windows)
        - `go build -o bin/main src/*.go` (Unix)
5. Copy the `config.example.yml` file to `config.yml` and edit the details
6. Start the API server
    - `.\bin\main.exe` (Windows)
    - `bin/main` (Unix)

## License

[MIT License](https://github.com/mcstatus-io/ping-server/blob/main/LICENSE)