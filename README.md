# About cardano-p2p

A CLI application to simplify Cardano node topology files updates.

## Notice

Curl contains pieces of source code that is Copyright (c) Ole Tange. This [notice](./CITATION) is included here to comply with the distribution terms.

## Backers

Thank you to all our backers! 🙏 [[Become a backer](https://github.com/sponsors/regel)]

## Sponsors

Support this project by becoming a sponsor. Your logo will show up here with a
link to your website. [[Become a sponsor](https://github.com/sponsors/regel)]

# The Cardano Blockchain Needs A Decentralized Node Discovery Service

The Cardano blockchain does not have a p2p feature to update topology files,
although all required information is registered in the blockchain.

To workaround this limitation, the Cardano community created a centralized API to
exchange node's IP addresses and update their topology files.

The `cardano-p2p` application solves this issue and enables fully decentralized
topology file updates.

## How It Works

The `cardano-p2p` application:
* Reads registered pool metadata in the Cardano blockchain (testnet and mainnet)
* Vets the data and ensures the IP and port are reachable
* Selects valid Cardano node relays to produce valid topology files
* Selects valid nodes randomly to ensure *fairness* and produce *reliable* Graphs topologies

## Backward Compatibility

`cardano-p2p` implements an API that is backward compatible with CLIO hosted service api.clio.one
and therefore is designed to simplify the transition.


