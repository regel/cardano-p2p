# About cardano-p2p

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Go Report Card](https://goreportcard.com/badge/github.com/regel/cardano-p2p)](https://goreportcard.com/report/github.com/regel/cardano-p2p)
[![Build](https://github.com/regel/cardano-p2p/actions/workflows/build.yaml/badge.svg)](https://github.com/regel/cardano-p2p/actions/workflows/build.yaml)
[![Docker pulls](https://img.shields.io/docker/pulls/regel/cardano-p2p)](https://hub.docker.com/r/regel/cardano-p2p)

A CLI application to simplify Cardano node topology files updates.

## Notice

`cardano-p2p` contains pieces of source code that is Copyright (c) Ole Tange. This [notice](./CITATION) is included here to comply with the distribution terms.

## Backers

Thank you to all our backers! 🙏 [[Become a backer](https://opencollective.com/gh-regel#backer)]

<a href="https://opencollective.com/gh-regel#backers" target="_blank"><img src="https://opencollective.com/gh-regel/backers.svg?width=890"></a>

## Sponsors

Support this project by becoming a sponsor. Your logo will show up here with a
link to your website. [[Become a
sponsor](https://opencollective.com/gh-regel#sponsor)]

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


