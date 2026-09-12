# TrainingPeaks Bot

Telegram bot that monitors your TrainingPeaks workout schedule. Sends instant notifications when workouts are added or changed, delivers a daily summary every evening, and posts a weekly plan every Monday. Optionally integrates with DeepSeek AI for nutrition recommendations.

## Features

- 🔔 **Change notifications** — new and updated workouts sent to Telegram immediately (checked every 10 minutes)
- 📅 **Daily summary at 7 PM Moscow time** — what was done today and what's planned for tomorrow
- 🗓 **Weekly plan every Monday at 7 PM Moscow time** — all workouts for the current week
- 🍽 **Nutrition advice** — DeepSeek analyzes tomorrow's workouts and recommends daily calorie intake

## Setup

### 1. Configure

Copy the example config and fill in your values:

```bash
cp config.example.json config.json
```

```json
{
  "telegram_token": "your-telegram-bot-token",
  "telegram_chat_id": 0,
  "tp_token": "your-trainingpeaks-auth-cookie",
  "tp_user_id": 12345678,
  "deepseek_api_key": "your-deepseek-api-key",
  "deepseek_model": "deepseek-chat",
  "deepseek_base_url": "https://api.deepseek.com",
  "deepseek_prompt": "You are a nutritionist. Based on the workouts planned for tomorrow, give a brief recommendation on calorie intake and key nutrition tips. Be concise."
}
```

| Field | Required | Description |
|-------|----------|-------------|
| `telegram_token` | ✅ | Bot token from [@BotFather](https://t.me/BotFather) |
| `telegram_chat_id` | ✅ | Your Telegram chat ID (see below) |
| `tp_token` | ✅ | TrainingPeaks session cookie (see below) |
| `tp_user_id` | — | TrainingPeaks athlete ID; leave `0` to auto-detect on startup |
| `deepseek_api_key` | — | DeepSeek API key; omit to disable nutrition advice |
| `deepseek_prompt` | — | Custom prompt for nutrition recommendations |
| `deepseek_model` | — | DeepSeek model name (default: `deepseek-chat`) |
| `deepseek_base_url` | — | DeepSeek API base URL (default: `https://api.deepseek.com`) |

### 2. Find your Telegram chat ID

Leave `telegram_chat_id` as `0`, start the bot, and send it any message — it will reply with your chat ID. Add it to `config.json` and restart.

### 3. Get your TrainingPeaks token

1. Log in at [trainingpeaks.com](https://trainingpeaks.com)
2. Open DevTools → Application → Cookies
3. Find the `Production_tpAuth` cookie and copy its value → set as `tp_token`

`tp_user_id` is auto-detected from the TrainingPeaks API on startup — you can leave it as `0`. The session token is refreshed automatically in memory every 9 minutes.

### 4. Run

#### Locally

```bash
go build
./trainingpeaks_bot
```

#### Docker

```bash
docker run -d --restart=always --name trainingpeaks \
  -v $(pwd)/config.json:/app/config.json \
  ghcr.io/kotovaleksandr/trainingpeaks_bot:latest
```

## Bot commands

| Command | Action |
|---------|--------|
| `/today` | Quick list of today's workouts |
| `/week` | Quick list of workouts for the rest of the current week |
| `/digest` | Full daily digest: today's summary + tomorrow's workouts + nutrition advice |
| `/plan` | Full week plan for the current week (Mon–Sun) |

## Notification schedule

| Event | When |
|-------|------|
| Workout change | Immediately (polled every 10 min) |
| Daily summary | Every day at 7 PM Moscow time |
| Weekly plan | Every Monday at 7 PM Moscow time |
