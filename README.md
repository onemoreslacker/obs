# 📡 Observer

<!-- TOC -->
- [Installation](#installation)
- [Configuration](#configuration)
<!-- /TOC -->

> A Telegram bot that keeps you in the loop by monitoring GitHub and StackOverflow activity.

## Features
- Instant Telegram notifications for:
    - New GitHub issues, pull requests, commits, comments
    - New StackOverflow questions, answers, comment activity
- Customizable tags and filters (e.g. `work` and `hobby` categories).

## Installation

To run Observer locally, make sure you have [Docker Compose](https://docs.docker.com/compose/install/standalone/) installed.

### 1. Clone the repository
```shell
git clone git@github.com:onemoreslacker/observer.git
```

### 2. Usage
```shell
make up # start application in verbose mode
make up-silent # start application in silent mode
make down # stop application
```

That's it! Your bot should now be up and running.

## Configuration

Edit `deploy/.env` file with your actual GitHub, StackOverflow and Telegram tokens
as well as Postgres environment variables.