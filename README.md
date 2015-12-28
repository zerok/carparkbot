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

Details about how to configure a command in Slack can be found [here](https://api.slack.com/slash-commands).
