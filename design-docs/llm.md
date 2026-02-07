# LLM Integration

Let's integrate with an external LLM to provide the illusion of sentience. Our strategy is thus:

* We will build a collection of facts about every user in the server.
* Occasionally, we will respond to a user's message with a generated response. This response will be tailored by the facts we have for this user.
* If a user engages after we respond, we will continue responding for a period of time before dismissing them. The length of the conversation is dictated by API rate limits.

## Fact collection

Let's keep an in-memory buffer for each user containing the messages that have not yet been processed for that user. Every time a non-bot user sends a message in the server, we will append the message to the corresponding buffer. We will flush every non-empty buffer when either one of two conditions are met: 1) one minute passes since the last flush or 2) the size of the buffer reaches 10.

We do not concern ourselves with a WAL/commit log because processing semantics are not a big concern.

When we flush, we send the buffer off to the LLM (aside: the LLM should be abstracted behind an interface, allowing us to switch implementations and hide complexity). The format of the message should be something along the lines of: "Given the following set of messages sent by a user, infer some facts about them. Output your response as a JSON list. Here are the messages: <messages>." In fact, let us standardize on requesting JSON responses and then deserializing those in the LLM abstraction we create.

Let us then store these facts in a local database. Either a hand-rolled solution or a standard one like RocksDB will do. Pick the simplest. We should still be able to deploy the application as a single binary.

## Fact compaction

Whenever the list of facts for a user grows above 15, send the list to the LLM with a prompt asking it to compact it down to 7, keeping only the most salient and merging related ones. Replace the existing facts for the user in our local database with this new set.

## Generated responses

On any given message, there should be a 5% chance we respond with a generated message.

We begin by taking the user's message and then sending the following prompt (or something like it) to the LLM: "Your job is to present yourself as annoying, dismissive, and obnoxious in a humorous manner. Some examples of messages you have sent in the past are <quote examples>. Someone has just sent you this message: <user message>. Respond to it as the aforementioned persona." We get the response and send it off to the user.

## LLM provider

We will be using Google AI Studio's free tier. The API key should be expected in the .env file. We will only use Gemini 2.5 Flash-Lite.

## Testing

All behavior described above should be tested comprehensively with unit tests.

## Follow-up messages

TODO: In the future, we will want to implement the follow-up feature described at the beginning of this document.

