Kaiju
=====

A lean and mean Bitcoin full node with high concurrency written in golang.

What and Why
-----

Kaiju aims to be the Bitcoin server for Bitcoin service providers, as opposed to Bitcoin Core being the Bitcoin client.

Services need a Bitcoin Server. There are two main reasons why Bitcoin Core (bitcoind) is not the perfect tool for Exchanges and Payment providers.
- It's not designed to connect a big number of nodes, but Services tend to like more connected nodes for better user experience.
- It's an all-in-one bundle containing a wallet and a miner, which are offen not used but introduce undesired complexcity.

Kaiju is designed to handle thouthands of connnections. And it does only one thing and try to be good at it: Being a lean and mean Bitcoin node, to communicate with the network, to verify Txs and Blocks.

<a href="https://github.com/oxfeeefeee/kaiju/blob/master/KDB.md" title="KDB">KDB</a>
----

A compact TXDatabase will be used in Kaiju, so that you don't need to store the full blockchain, but can sitll verify transactions.

Educational Purpose
----

When response to conspiracy theories, bitcoiner offen say that Bitcoin is an open source project, and everyone can check out the source code and see it for themselves. But the reality is, the code of Bitcoin Core is hard to read for various of reasons.

Kaiju is planning to be small, clean and simple, hopefully under 5K lines of code. Good documentation for code readers is also part of the plan so that more people could really read and learn.

Current Status
---
WIP, hopefully nobody sees this :) 