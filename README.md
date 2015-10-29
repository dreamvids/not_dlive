# DreamVids Live Server

DreamVids uses this server to power its chat rooms and the live system

## Installation

An up and running Go environnement is required (with $GOPATH set up).

```
go get github.com/dreamvids/dlive
cd $GOPATH/src/github.com/dreamvids/dlive
```

If you want to be able to run the server more easily, add $GOPATH/bin to your $PATH.

By default, the executable's path is $GOPATH/bin/dlive.

## Configuration

Here is the default configuration file

```json
{
	"http-port": 8081,

	"chat-max-clients": 10,
	"chat-modo-rank": 4,
	"chat-admin-rank": 5,
	"chat-mute-message": "Vous n'avez pas été sage. Vous ne pouvez donc pas parler.",

	"db-host": "172.17.0.1:3306",
	"db-username": "root",
	"db-password": "root",
	"db-name": "database"

}
```

Here is how to run the server (assuming you have dlive in your path).

```
dlive --config /path/to/dlive.json --log log.txt
```

## Usage

To push a stream to the server using ffmpeg:

```
ffmpeg ... -f webm http://127.0.0.1:8081/stream/push/<id>
```

To watch it, use the following url: http://127.0.0.1:8081/stream/pull/<id>
