## FileShare

[![wakatime](https://wakatime.com/badge/user/b9ae0171-376e-4d7d-9ceb-ea72185e2c2e/project/88b35923-83d5-4f62-b1e6-62d0e8c54e6c.svg)](https://wakatime.com/badge/user/b9ae0171-376e-4d7d-9ceb-ea72185e2c2e/project/88b35923-83d5-4f62-b1e6-62d0e8c54e6c)

This is a simple file server written in Go. It serves files from a specified directory and provides a simple web interface for browsing and downloading files.

![Screenshot 1](./screenshot.webp)

### Installation and Usage

Install the exetuable by downloading the latest release from the [releases page](https://github.com/amixaam/FileShare/releases/).

Run the server by executing the executable. The server will start listening on port 8080 by default and serve files from the `~/SharedFiles` directory.

```
./fileshare
```

When the server is ran, the console will display multiple addresses that can be used to access the server. Simply copy an adress and paste it into your browser to access the server. These adresses can also be sent to other people to allow them to access the server locally.

### CLI Usage

The CLI provides a simple way to change the server's configuration and start the server. This overrides the configuration file.

By default, the server will serve files from the `~/SharedFiles` directory. You can specify a different directory using the `-dir` flag.

```
./fileshare -dir /path/to/directory
```

The server will start listening on port 8080 by default. You can specify a different port using the `PORT` environment variable.

```
PORT=8000 ./fileshare
```

### Configuration

Using a config file, you can change the server's configuration without manually specifying command line arguments. The .yaml file must be named `fileshare.yaml` and be located in the same directory as the executable.

Or you can specify a custom config file using the `-config` flag.

```
./fileshare -config=/path/to/custom-config.yaml
```

The configuration file supports the following options:

| Option      | Description                               | Default         |
| ----------- | ----------------------------------------- | --------------- |
| `directory` | Directory to serve files from             | `~/SharedFiles` |
| `port`      | Port to listen on                         | `8080`          |
| `dotfiles`  | Show or hide hidden files and directories | `true`          |
| `domain`    | Domain to use for links                   | `localhost`     |

Check out the [example config file](fileshare.yaml) for an example of how to configure the server.
