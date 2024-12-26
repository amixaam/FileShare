### FileShare

This is a simple file server written in Go. It serves files from a specified directory and provides a simple web interface for browsing and downloading files.

![Screenshot 1](./screenshot.webp)

### Features

-   Serves files from a specified directory
-   Provides a simple web interface for browsing and downloading files
-   Supports ZIP downloads for directories
-   Displays local IP addresses for easy copy and paste for sharing
-   Column sorting for file listing
-   Dark / light mode toggle
-   Easy to use and customize with a simple configuration file

### Installation and Usage

Install the exetuable by downloading the latest release from the [releases page](https://github.com/amixaam/FileShare/releases/).

Run the server by executing the executable. The server will start listening on port 8080 by default and serve files from the `~/SharedFiles` directory.

```
./fileshare
```

### CLI Usage

By default, the server will serve files from the `~/SharedFiles` directory. You can specify a different directory using the `-dir` flag.

```
./fileshare -dir /path/to/directory
```

The server will start listening on port 8080 by default. You can specify a different port using the `PORT` environment variable.

```
PORT=8000 ./fileshare
```

### Configuration

You can configure the server using a YAML file or command line flags. The default configuration file is `fileshare.yaml` and can be found in the **same directory as the executable**. You can also specify a custom configuration file using the `-config` flag.

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
