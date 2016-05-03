# hdocker
hdocker to docker is what htop is to top

# Overview

hdocker presents an interactive realtime view of the running containers.
The program created mostly out of personal need to quickly kill a container,
or inspect the container's details, like command or network setting.

# Demo

[![asciicast](https://asciinema.org/a/bas355upyctx9nkllf7hqeklt.png)](https://asciinema.org/a/bas355upyctx9nkllf7hqeklt)

# Features

hdocker can:
 - Display a list of running containers
 - Parse a JSON config file with the inspect layout. Accepts any string accepted by ```docker inspect --format``` command
 - Adapt to terminal size changes
 - Kill a docker container
 
Planned features:
- Support for DOCKER_HOST environment variable
- Show disk space usage
- Emacs / Vim keybindings for navigation
