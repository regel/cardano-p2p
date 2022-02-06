# About cardano-p2p

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Go Report Card](https://goreportcard.com/badge/github.com/regel/cardano-p2p)](https://goreportcard.com/report/github.com/regel/cardano-p2p)
[![Build](https://github.com/regel/cardano-p2p/actions/workflows/build.yaml/badge.svg)](https://github.com/regel/cardano-p2p/actions/workflows/build.yaml)
[![Docker pulls](https://img.shields.io/docker/pulls/regel/cardano-p2p)](https://hub.docker.com/r/regel/cardano-p2p)
[![codecov](https://codecov.io/github/regel/cardano-p2p/coverage.svg)](https://codecov.io/gh/regel/cardano-p2p)


A CLI application to simplify Cardano node topology files updates.

## Backers

Thank you to all our backers! 🙏 [[Become a backer](https://opencollective.com/gh-regel#backer)]

<a href="https://opencollective.com/gh-regel#backers" target="_blank"><img src="https://opencollective.com/gh-regel/backers.svg?width=890"></a>

## Sponsors

Support this project by becoming a sponsor. Your logo will show up here with a
link to your website. [[Become a
sponsor](https://opencollective.com/gh-regel#sponsor)]

# The Cardano Blockchain Needs A Decentralized Node Discovery Service

The Cardano blockchain does not yet have a production ready p2p feature to update topology files,
although all required information is registered in the blockchain.

To find a temporary workaround, the Cardano community created a centralized API (known as CLIO1, api.clio.one) to
exchange node's IP addresses and update their topology files.

The `cardano-p2p` application solves this issue and enables fully decentralized
topology file updates.

## How It Works

The `cardano-p2p` application:
* Connects to [Ogmios](https://ogmios.dev/) websocket in order to get registered pool parameters found in the Cardano blockchain
* Verifies each pool metadata and sends a TCP probe to ensure their IP and port are still reachable
* Selects Cardano nodes that passed the above test in order to produce valid topology files
* Serves list of Cardano nodes randomly to ensure *fairness* and produce *reliable* Graphs topologies

## Backward Compatibility

`cardano-p2p` implements an API that is backward compatible with CLIO hosted service api.clio.one
and therefore is designed to simplify the transition.


