import discord
import random
import re
import os
import time
from dotenv import load_dotenv
from vaderSentiment.vaderSentiment import SentimentIntensityAnalyzer
from lingua import Language, LanguageDetectorBuilder

load_dotenv()

DISCORD_TOKEN = os.getenv('DISCORD_TOKEN')
ELECTROSHK_ID = int(os.getenv('ELECTROSHK_ID', 0))
MODRIVER_ID = int(os.getenv('MODRIVER_ID', 0))

intents = discord.Intents.default()
intents.message_content = True

client = discord.Client(intents=intents)
analyzer = SentimentIntensityAnalyzer()
language_detector = LanguageDetectorBuilder.from_all_languages().build()

# Quick reply trigger settings
QUICK_REPLY_WINDOW = 20  # seconds after bot reply
QUICK_REPLY_COOLDOWN = 300  # 5 minute cooldown
SPEAK_ENGLISH_COOLDOWN = 300  # 5 minute cooldown

# Track bot replies and cooldowns per channel
last_bot_reply_time = {}  # channel_id -> timestamp
last_quick_reply_trigger = {}  # channel_id -> timestamp
last_speak_english_trigger = {}  # channel_id -> timestamp

# Random quotes (marked with * in quotes.txt) - 2% chance any message
RANDOM_QUOTES = [
    "Yeah but he's the hood king!",
    "That's why I'm being odenigbo about it!",
    "RAUL WHAT THE HELL?!",
    "My grade went down… cuz it couldn't really go up! HEHEHEHE",
    "Rough year for Kanye",
    "We did the BDSM test this summer and let's just say, I was very confused",
    "I've seen all of the hockey team player's dicks",
    "Ben always gives me the satisfaction",
    "Public shaming is a good form of shaming",
]

# Positive sentiment responses when Fogel is mentioned
POSITIVE_RESPONSES = [
    "Am I right fo dat?",
    "I'm showing ya fam!",
]

# Negative sentiment responses when Fogel is mentioned
NEGATIVE_RESPONSES = [
    "Don't bite the hand that feeds you 😡",
    "…I don't care about this anymore!",
    "I will BEAT you!",
]

# "I think" responses
THINK_RESPONSES = [
    "DING DING DING",
    "Very good, {user}!",
    "Correct!",
]


def is_non_english(text):
    """Check if text is in a non-English language using lingua-py."""
    # Skip very short texts (less reliable detection)
    if len(text.strip()) < 3:
        return False
    detected = language_detector.detect_language_of(text)
    if detected is None:
        return False
    return detected != Language.ENGLISH


def is_fogel_mentioned(message):
    """Check if Fogel/fogelbot is mentioned."""
    content_lower = message.content.lower()
    if 'fogel' in content_lower:
        return True
    if client.user in message.mentions:
        return True
    return False


def is_quick_reply(message):
    """Check if message is within 20 seconds of bot's last reply (with 5 min cooldown)."""
    channel_id = message.channel.id
    current_time = time.time()

    # Check if bot has replied in this channel recently
    if channel_id not in last_bot_reply_time:
        return False

    # Check if within the quick reply window
    time_since_bot_reply = current_time - last_bot_reply_time[channel_id]
    if time_since_bot_reply > QUICK_REPLY_WINDOW:
        return False

    # Check cooldown
    if channel_id in last_quick_reply_trigger:
        time_since_last_trigger = current_time - last_quick_reply_trigger[channel_id]
        if time_since_last_trigger < QUICK_REPLY_COOLDOWN:
            return False

    return True


def user_mentioned_or_author(message, user_id):
    """Check if a specific user sent the message or was mentioned."""
    if message.author.id == user_id:
        return True
    for mention in message.mentions:
        if mention.id == user_id:
            return True
    return False


@client.event
async def on_ready():
    print(f'Logged in as {client.user}')
    for guild in client.guilds:
        print(f'Channels in {guild.name}:')
        for channel in guild.text_channels:
            print(f'  - {channel.name}')


@client.event
async def on_message(message):
    # Don't reply to ourselves
    if message.author == client.user:
        return

    content = message.content
    content_lower = content.lower()
    response = None

    # Check for Fogel mention with sentiment analysis
    if is_fogel_mentioned(message):
        sentiment = analyzer.polarity_scores(content)
        print(f"Fogel mentioned! Sentiment: {sentiment['compound']}")  # Debug
        if sentiment['compound'] >= 0.05:
            response = random.choice(POSITIVE_RESPONSES)
        elif sentiment['compound'] <= -0.05:
            response = random.choice(NEGATIVE_RESPONSES)
        else:
            # Neutral mention - pick randomly from positive responses
            response = random.choice(POSITIVE_RESPONSES)

    # Quick reply trigger (within 20 seconds of bot reply, 5 min cooldown)
    elif is_quick_reply(message):
        sentiment = analyzer.polarity_scores(content)
        print(f"Quick reply detected! Sentiment: {sentiment['compound']}")  # Debug
        if sentiment['compound'] >= 0.05:
            response = random.choice(POSITIVE_RESPONSES)
        elif sentiment['compound'] <= -0.05:
            response = random.choice(NEGATIVE_RESPONSES)
        else:
            response = random.choice(POSITIVE_RESPONSES)
        # Update cooldown
        last_quick_reply_trigger[message.channel.id] = time.time()

    # Michael/Jordan/hood/king trigger
    elif re.search(r'\b(michael|jordan|king|prince|hood|ghetto)\b', content_lower):
        response = "Yeah but he's the hood king!"

    # Odenigbo variations
    elif re.search(r'\bod.{0,2}n.{0,2}bo\b', content_lower):
        response = "That's why I'm being odenigbo about it!"

    # Love/like/hate/want/need patterns
    elif re.search(r'\bi (just )?(love|like|dislike|hate|want|need)\b', content_lower):
        response = "I know you do 😉"

    # Kanye/West
    elif re.search(r'\b(kanye|west)\b', content_lower):
        response = "Rough year for Kanye"

    # Food/eating
    elif re.search(r'\b(food|eat|eating|hungry|delicious|yummy|tasty|cook)\b', content_lower):
        response = f"{message.author.mention}, thats moan-worthy"

    # BDSM
    elif 'bdsm' in content_lower:
        response = "We did the BDSM test this summer and let's just say, I was very confused"

    # Hockey
    elif 'hockey' in content_lower:
        response = "I've seen all of the hockey team player's dicks"

    # Ben/Benjamin
    elif re.search(r'\b(ben|benjamin)\b', content_lower):
        response = "Ben always gives me the satisfaction"

    # Shame/shaming
    elif re.search(r'\bsham(e|ing)\b', content_lower):
        response = "Public shaming is a good form of shaming"

    # Grade/academic
    elif re.search(r'\b(grade|grades|smart|academic)\b', content_lower):
        response = "My grade went down… cuz it couldn't really go up! HEHEHEHE"

    # "I think" / "I believe" - always respond with one of 3 options
    elif re.search(r'\bi (think|believe)\b', content_lower):
        resp = random.choice(THINK_RESPONSES)
        response = resp.format(user=message.author.mention)

    # Non-English language - 100% chance with 5 minute cooldown
    elif is_non_english(content):
        channel_id = message.channel.id
        current_time = time.time()
        can_trigger = True
        if channel_id in last_speak_english_trigger:
            if current_time - last_speak_english_trigger[channel_id] < SPEAK_ENGLISH_COOLDOWN:
                can_trigger = False
        if can_trigger:
            response = "SPEAK ENGLISH!!"
            last_speak_english_trigger[channel_id] = current_time

    # User-specific triggers
    electroshk_triggered = (ELECTROSHK_ID and user_mentioned_or_author(message, ELECTROSHK_ID)) or re.search(r'\byifan\b', content_lower)
    if response is None and electroshk_triggered:
        roll = random.random()
        if roll < 0.01:  # 1% YEEEbowl
            response = "Did you really do the YEEEbowl without me?!?"
        elif roll < 0.06:  # 5% YEEE
            response = "YEEEEEEEEEEEEEEEEEEEEEE"

    modriver_triggered = (MODRIVER_ID and user_mentioned_or_author(message, MODRIVER_ID)) or re.search(r'\b(raul|rahul)\b', content_lower)
    if response is None and modriver_triggered:
        if random.random() < 0.05:  # 5% chance
            response = random.choice(["RAUL WHAT THE HELL?!", "Shut up, RAUL"])

    # 5% chance to respond to any negative sentiment (even without Fogel mention)
    if response is None and random.random() < 0.05:
        sentiment = analyzer.polarity_scores(content)
        if sentiment['compound'] <= -0.05:
            response = random.choice(NEGATIVE_RESPONSES)

    # 2% random quote chance (if no other response)
    if response is None and random.random() < 0.02:
        response = random.choice(RANDOM_QUOTES)

    # Send response if we have one
    if response:
        await message.channel.send(response)
        # Track bot reply time for quick reply detection
        last_bot_reply_time[message.channel.id] = time.time()


if __name__ == '__main__':
    if not DISCORD_TOKEN:
        print("Error: DISCORD_TOKEN not found in .env file")
        print("Copy .env.example to .env and add your token")
    else:
        client.run(DISCORD_TOKEN)
