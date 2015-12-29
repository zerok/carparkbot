# Car Park Bot

This is a little bot that listens for messages in a predefined channel
indicating that a car is blocking another one and notifies the offending owner
that they should remove their car:

> /carblocks G12345K

... will reply with

> @otherperson: Your :car: is blocking @yourname. Please move it.

The mapping is taken from a simple CSV file with following format:

```
<licenseplate>,<slackusername>
```

The license plate should not include any hyphens or spaces.

If you change the content of that mapping file, the internal store is updated so
you don't have to restart the command endpoint every time a new mapping is added.

Details about how to configure a command in Slack can be found
[here](https://api.slack.com/slash-commands).


## Usage

```
$ ./carparkbot -token <SlackCommandToken> -channel <YourGeneralChannel> -mapping path/to/mapping.csv
```


### Direct messages for notification

If you additionally want to send a direct message to a car holder, you have to
generate an [API token](https://api.slack.com/web) and start the bot with the
`-dm` and `-api-token=YOURTOKEN` parameters.


### Dynamic mapping

If you don't specify a mapping file, you have to provide that dynamically. You
can do so by sending the CSV-formatted data via POST to the `/mapping/`
endpoint:

```
$ http POST https://localhost:8080/mapping/ < mapping.csv
```

**Warning:** Make sure to put some AUTH around the `/mapping/` endpoint.


## How to build

First you have to have [glide](https://github.com/Masterminds/glide)
installed. After that:

```
$ glide install
$ go build
```
