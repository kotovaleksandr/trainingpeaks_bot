# TrainingPeaks Bot
Telegram bot check your TrainingPeaks profile workouts and send notification if some workouts added or changed.

# Usage
* Get Telegram bot token (via BotFather)
* Get TrainingPeaks auth token, see cookies in browser. Cookie name is Production_tpAuth
* Create telegram_token
* Launch trainingpeaks_bot
* Start chat with bot, via commands token and id send TrainingPeaks token and user id
* Or, via docker: 
```
docker run -d --restart=always --name trainingpeaks -v $(pwd)/telegram_token:/app/telegram_token -v $(pwd)/users.database:/app/users.database ghcr.io/kotovaleksandr/trainingpeaks_bot:latest
```
