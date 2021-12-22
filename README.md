# TrainingPeaks Bot
Telegram bot check your TrainingPeaks profile workouts and send notification if some workouts added or changed.

# Usage
* Get Telegram bot token (via BotFather)
* Get TrainingPeaks auth token, see cookies in browser. Cookie name is Production_tpAuth
* Create telegram_token and tp_token files with tokens
* Launch trainingpeaks_bot with env variable USER_ID = trainingpeaks user id, can see in browser network calls to trainingpeaks.com
* Or, via docker: 
```
docker run -d --restart=always --name trainingpeaks -v ./tp_token:/app/tp_token -v ./telegram_token:/app/telegram_token -e USER_ID=123456 ghcr.io/kotovaleksandr/trainingpeaks_bot:latest
```
