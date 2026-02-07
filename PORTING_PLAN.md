# Fogelbot Python to Golang Port Plan

## Overview

Port the Python Discord bot (Fogelbot) to Golang to reduce memory usage while maintaining feature parity.

## Key Decisions

- **Structure**: Multi-package organization for maintainability
- **Sentiment**: `github.com/knuppe/vader` (6.5x faster than govader)
- **Language Detection**: `github.com/pemistahl/lingua-go` with limited language set
- **Discord**: `github.com/bwmarrin/discordgo` (most mature Go Discord library)
- **Config**: `github.com/joho/godotenv` for .env loading

## Project Structure

```
/Users/dichlorodiphen/Desktop/fogelbot/go/
├── main.go                 # Entry point, Discord client setup
├── go.mod                  # Module definition
├── config/
│   └── config.go           # Configuration loading and constants
├── bot/
│   ├── bot.go              # Discord bot struct and initialization
│   ├── handlers.go         # Event handlers (on_ready, on_message)
│   └── responses.go        # Response string constants
├── triggers/
│   ├── triggers.go         # Trigger interface and registry
│   ├── fogel.go            # Fogel mention trigger
│   ├── quickreply.go       # Quick reply trigger
│   ├── keywords.go         # Keyword pattern triggers (11 patterns)
│   ├── language.go         # Non-English detection trigger
│   ├── users.go            # User-specific triggers (Electroshk, Modriver)
│   └── random.go           # Random chance triggers (quotes, negative sentiment)
├── sentiment/
│   └── analyzer.go         # Sentiment analysis wrapper
├── language/
│   └── detector.go         # Language detection wrapper
└── state/
    └── cooldowns.go        # Thread-safe cooldown management
```

## Features to Port

From `/Users/dichlorodiphen/Desktop/fogelbot/main.py`:

1. **Fogel Mention** - Sentiment-based responses when "fogel" mentioned or bot @mentioned
2. **Quick Reply** - Respond within 20s of bot's last reply (5min cooldown)
3. **Keyword Triggers** (11 patterns):
   - michael/jordan/king/prince/hood/ghetto
   - odenigbo
   - love/like/dislike/hate/want/need
   - kanye/west
   - food/eat/eating/hungry/delicious/yummy/tasty/cook
   - bdsm
   - hockey
   - ben/benjamin
   - shame/shaming
   - grade/grades/smart/academic
   - "i think"/"i believe"
4. **Language Detection** - "SPEAK ENGLISH!!" for non-English (5min cooldown)
5. **User-Specific Triggers** - Special responses for ELECTROSHK_ID and MODRIVER_ID
6. **Random Triggers** - 5% negative sentiment, 2% random quotes

## Implementation Steps

### Step 1: Update go.mod with dependencies
```bash
cd /Users/dichlorodiphen/Desktop/fogelbot/go
go get github.com/bwmarrin/discordgo
go get github.com/knuppe/vader
go get github.com/pemistahl/lingua-go
go get github.com/joho/godotenv
```

### Step 2: Create config package
- `config/config.go`: Load DISCORD_TOKEN, ELECTROSHK_ID, MODRIVER_ID from .env
- Define constants: QUICK_REPLY_WINDOW=20s, cooldowns=5min, sentiment threshold=0.05

### Step 3: Create state package
- `state/cooldowns.go`: Thread-safe maps using sync.RWMutex
  - lastBotReplyTime (per channel)
  - lastQuickReplyTrigger (per channel)
  - lastSpeakEnglishTrigger (per channel)

### Step 4: Create sentiment package
- `sentiment/analyzer.go`: Wrapper around knuppe/vader
- Compound() method returning score from -1 to 1

### Step 5: Create language package
- `language/detector.go`: Wrapper around lingua-go
- Limit to common languages: English, Spanish, French, German, Italian, Portuguese, Chinese, Japanese, Korean, Russian
- IsNonEnglish() method

### Step 6: Create triggers package
- `triggers/triggers.go`: Define Trigger interface and Registry
- Implement each trigger type with Check(ctx) (response, matched) pattern
- Register triggers in priority order matching Python implementation

### Step 7: Create bot package
- `bot/responses.go`: POSITIVE_RESPONSES, NEGATIVE_RESPONSES, RANDOM_QUOTES, etc.
- `bot/bot.go`: Bot struct with session, cooldowns, triggers, config
- `bot/handlers.go`: OnReady and OnMessage handlers

### Step 8: Update main.go
- Initialize all services
- Set Discord intents (GuildMessages + MessageContent)
- Register handlers and open connection
- Graceful shutdown with signal handling

## Key Implementation Details

### Thread-Safe Cooldowns
```go
type CooldownManager struct {
    mu sync.RWMutex
    lastBotReplyTime        map[string]time.Time
    lastQuickReplyTrigger   map[string]time.Time
    lastSpeakEnglishTrigger map[string]time.Time
}
```

### Trigger Interface
```go
type Trigger interface {
    Name() string
    Check(ctx *Context) (response string, matched bool)
}
```

### Trigger Priority Order (same as Python)
1. Fogel mention (sentiment-based)
2. Quick reply (within 20s window)
3. Keyword triggers (in order from Python)
4. Language detection
5. User-specific triggers
6. Random negative sentiment (5%)
7. Random quotes (2%)

## Files to Reference

- `/Users/dichlorodiphen/Desktop/fogelbot/main.py` - All Python logic, patterns, responses
- `/Users/dichlorodiphen/Desktop/fogelbot/quotes.txt` - Random quotes list
- `/Users/dichlorodiphen/Desktop/fogelbot/.env` - Environment variables

## Verification

1. **Build**: `go build -o fogelbot` compiles without errors
2. **Run**: Bot connects to Discord and logs "Logged in as..."
3. **Test triggers**: Send messages containing trigger words and verify responses match Python behavior
4. **Test cooldowns**: Verify 5-minute cooldowns work correctly
5. **Test sentiment**: Verify positive/negative sentiment detection works
6. **Memory comparison**: Compare memory usage between Python and Go versions using `top` or `htop`
