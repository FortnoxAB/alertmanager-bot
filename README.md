# alertmanager-bot
A bot to send alertmanager notifications to slack to reduce webhook configuration overhead. 
Just setup the api_url in slack_config in alertmanager to your running instance with path: /webhook

# Running

```
Usage of alertmanager-bot:
  -log-format
    	Change value of Log-Format. (default <nil>)
  -log-level
    	Change value of Log-Level.
  -port
    	Change value of Port. (default 8080)
  -token
    	Change value of Token.

Generated environment variables:
   CONFIG_LOG_FORMAT
   CONFIG_LOG_LEVEL
   CONFIG_PORT
   CONFIG_TOKEN

```
