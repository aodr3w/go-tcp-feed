# Go TCP Feed

A simple TCP-based message feed written in Go. This application allows clients to publish messages to a shared feed in near-real-time, with all messages stored in a PostgreSQL database.

---

## Table of Contents

1. [Description](#description)
2. [Features](#features)
3. [Requirements](#requirements)
4. [Quick Start](#quick-start)
   - [Clone the Repository](#1-clone-the-repository)
   - [Set Up Environment Variables](#2-set-up-environment-variables-optional)
   - [Build and Start the Application](#3-build-and-start-the-application)
   - [View Running Sessions](#4-view-running-sessions)
5. [Usage and Sample Logs](#usage-and-sample-logs)
   - [Publisher Session](#publisher-session)
   - [Feed Session](#feed-session)
   - [Server Session](#server-session)
6. [Stopping Everything](#stopping-everything)
7. [Inspired By](#inspired-by)

---

## Description

Go TCP Feed is a command-line application that uses two TCP ports:
- **Port 3000** for reading messages (subscription).
- **Port 2000** for writing messages (publishing).

Multiple clients can connect to these ports to send and receive messages in real time. The PostgreSQL database ensures persistence and maintains message history.

---

## Features

- **Publish Messages**: Any user can send a message by entering a name (minimum 4 characters) and typing out their message.
- **Subscribe to Feed**: A “feed” client displays all new messages in near real time.
- **Server Logging**: Logs both read and write connections, as well as successful user creations.

---

## Requirements

- **OSX or Ubuntu Linux**.
- **Go** 1.18+.
- **Docker** and Docker daemon running (for the PostgreSQL container).
- **PostgreSQL 15 Docker image**.
- **tmux** installed (on macOS, you can install via `brew install tmux`).

---

## Quick Start

### 1. Clone the Repository

```
git clone https://github.com/aodr3w/go-tcp-feed.git
cd go-tcp-feed
```

### 2. Set Up Environment Variables (Optional)

If you have a .env file, environment variables (e.g., PG_PASS, DB_USER, DB_NAME) will be automatically loaded by the Makefile.

### 3. Build and Start the Application

```
make app
```

- Installs missing dependencies on macOS (like tmux).
- Checks/starts Docker.
- Launches a PostgreSQL container for storing messages.
- Spins up server, feed, and publisher in separate tmux sessions.

### 4. View Running Sessions
```
tmux ls
```

You should see sessions named server, feed, and publisher.


## Usage and Sample Logs

### Publisher Session

Publisher allows you to publish messages under a chosen username:

```
name (atleast 4 characters): << john
2025/01/10 21:44:37 userID-john
<< hello world my name is john  😄
<< happy to be here 💯 🔥😎
<<
```

### Feed Session

Feed session displays all messages as they arrive:

```
john >> hello world my name is jon 😄 [1/10/2025 18:45:01]
john >> happy to be here 💯🔥😎 [1/10/2025 18:45:26]
```

### Server Session

Server accepts connections on two ports (3000 for read, 2000 for write):

```
[readMessages] 2025/01/10 21:43:09 server is accepting connections on 3000
[writeMessages]  2025/01/10 21:43:09 server is accepting connections on 2000
[readMessages] 2025/01/10 21:43:11 new connection received
[writeMessages]  2025/01/10 21:44:37 user successfully created: &{1 john}
```


### Stopping Everything

When you’re done, stop the Docker container and tmux sessions with:

```
make stop
```

This terminates the PostgreSQL container and kills any related tmux sessions.


### Inspired By

•	https://codingchallenges.fyi/challenges/challenge-realtime-chat

Enjoy real-time publishing and subscribing with Go TCP Feed!
Feel free to open an issue or submit a PR if you have any questions or improvements.