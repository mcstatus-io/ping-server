# API Server
The REST server that powers the API for mcstatus.io. This repository is open source to allow developers to run their own Minecraft server status API server.

## Requirements

- [Git](https://git-scm.com/)
- [Go](https://go.dev/)
- [Redis](https://redis.io/)
- [GNU Make](https://www.gnu.org/software/make/) (*optional*)

## Installation

1. Clone the repository to a folder
    - `git clone https://github.com/mcstatus-io/api-server.git`
2. Move the working directory into the folder
    - `cd api-server`
3. Install all required dependencies
    - `go get ...`
4. Build the executable
    - Using GNU make
        - `make build`
    - Without GNU make
        - `go build -o .\bin\main.exe .\src\*.go` (Windows)
        - `go build -o bin/main src/*.go` (Unix)
5. Copy the `config.example.yml` file to `config.yml` and edit the details
6. Start the API server
    - `.\bin\main.exe` (Windows)
    - `bin/main` (Unix)

## License

[MIT License](https://github.com/mcstatus-io/api-server/blob/main/LICENSE)
        