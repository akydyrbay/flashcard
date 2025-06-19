# 📚 Telegram Flashcard Bot

A Telegram bot built in Go that allows users to **create, quiz, list, and delete named flashcard sets directly in chat**. Users can save a series of question-answer pairs under a custom name, retrieve and quiz themselves on demand, and manage their sets easily. Includes optional integration with AI (ChatGPT) to automatically generate flashcards from a user prompt.

---

## ✨ Features

- 📥 Save flashcard sets with custom names  
- ❓ Automatically generate flashcards using ChatGPT  
- 📋 List all saved flashcard sets  
- 🔄 Retrieve and quiz yourself on a set (flashcard style)  
- ❌ Delete flashcard sets you no longer need  
- 🔐 User-specific storage  
- 💾 Lightweight SQLite persistence  
- 🔁 Polling-based message handling *(or Webhook-ready)*

---

## 🛠️ Tech Stack

- **Golang** – core application logic and Telegram bot  
- **SQLite** – lightweight embedded database  
- **Telegram Bot API** – handles user interaction  
- **OpenAI API** – generates flashcards from prompts  
- **Custom Event Consumer** – decouples event fetching and processing  

---

## 📦 Project Structure

```text
├── cmd/ # Entry point
├── client/telegram/ # Telegram Bot API client
├── consumer/eventconsumer # Event consumer loop
├── events/telegram/ # Event fetching and command processing
├── storage/sqlite/ # SQLite storage implementation
├── lib/e/ # Error wrapping helpers
├── go.mod / go.sum # Go modules
├── data/sqlite/ # Data storage
```

---

## ⚙️ Setup & Run

1. **Clone the repo:**

```bash
git clone https://github.com/yourusername/telegram-flashcard-bot.git
cd telegram-flashcard-bot
```
2. **Start the program:**
```bash 
go build .
./flashcard -tg-bot-token 'token'
```
