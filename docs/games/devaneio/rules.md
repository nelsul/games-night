# Devaneio

A storytelling game for 4 to 14 people, each on their own phone, in the same room or on a call.

## A round

1. The **teller** (it goes around the table) picks a card from their hand and writes a clue for it: a word, a sentence, anything.
2. Everyone else plays, in secret, the card from their own hand that best fits the clue.
3. The cards are shuffled and turned face up. Everyone but the teller votes for the card they think is the teller's. Nobody can vote for their own card.
4. The cards turn over: each shows who played it and who voted for it.

## Points

- If **everybody** found the teller's card, or **nobody** did, the clue was too easy or too hard: the teller gets 0 and everyone else gets 2.
- Otherwise the teller and each person who found the card get 3, and each card that fooled someone gives its owner 1 point per vote, up to 3.

The match is 5 or 8 rounds, or one round for each person ("uma volta"). The highest score wins; ties are shared. Everybody keeps 6 cards in hand, refilled after every round.

## The pictures

There are 160 cards. Each one is a small scene (a sky, something in it, a landscape, a main object and a few small ones, in one of eight palettes) drawn by the web app from the card number alone, so the server only deals numbers and every phone draws the same picture. All art is original.

## What the server hides

Hands never leave the server except to their owner. During the vote the table only has pictures: nobody can tell whose card is whose, the teller's included. Owners and votes show up once the round is scored.

## Bots

A bot can't see the pictures, so it plays at random (a clue from a fixed list, any card, any vote). It never moves the table on to the next round: a person always does.
