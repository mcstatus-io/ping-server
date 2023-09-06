# Ping Server
![](https://img.shields.io/github/languages/code-size/mcstatus-io/ping-server)
![](https://img.shields.io/github/issues/mcstatus-io/ping-server)
![](https://img.shields.io/github/actions/workflow/status/mcstatus-io/ping-server/go.yml)
![](https://img.shields.io/uptimerobot/ratio/m792379047-190f2d39d31ecafa9b1f86ab)

This is the source code for the API of the [mcstatus.io](https://mcstatus.io) website (`api.mcstatus.io`). This API server is built using [Go](https://go.dev) with [Fiber](https://docs.gofiber.io/) as the HTTP server of choice. This uses a custom Minecraft utility library found in the [mcstatus-io/mcutil](https://github.com/mcstatus-io/mcutil) repository. You are free to modify and host your own copy of this server as long as the [license](https://github.com/mcstatus-io/ping-server/blob/main/LICENSE) permits. If you do not wish to self host, we host a public and free-to-use copy which you may learn more about by visiting the [official API documentation](https://mcstatus.io/docs).

Please note that while this repository may seem to conform to some versioning standard, it most certainly does not. Updates are pushed at random, with no semantic versioning in place. Any update (also known as a *commit*) may suddenly break existing configurations without notice or warranty. If you run a privately hosted ping server, please refer to the updated example configuration file before attempting to update to the latest commit. 

## API Documentation

https://mcstatus.io/docs

## Requirements

- [Go](https://go.dev/)
- [Redis](https://redis.io/)
- [GNU Make](https://www.gnu.org/software/make/)

## Getting Started

```bash
# 1. Clone the repository (or download from this page)
$ git clone https://github.com/mcstatus-io/ping-server.git

# 2. Move the working directory into the cloned repository
$ cd ping-server

# 3. Run the build script
$ make

# 4. Copy the `config.example.yml` file to `config.yml` and modify details as needed
$ cp config.example.yml config.yml

# 5. Start the development server
$ ./bin/main

# The server will be listening on http://localhost:3001 (default host + port)
```

## License

[MIT License](https://github.com/mcstatus-io/ping-server/blob/main/LICENSE)