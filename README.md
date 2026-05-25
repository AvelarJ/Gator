# Gator

A command-line RSS feed aggregator written in Go. Gator lets you register users, subscribe to RSS feeds, aggregate posts in the background on a schedule, and browse the latest content — all backed by a PostgreSQL database.

---

## Prerequisites

Before installing Gator, make sure you have the following:

- **Go** (1.21 or later) — verify with `go version`
- **PostgreSQL** — a running instance with a database created for Gator
- **Goose** — used to run database migrations

Install Goose with:

```bash
go install github.com/pressly/goose/v3/cmd/goose@latest
```

---

## Installation

```bash
go install github.com/AvelarJ/Gator@latest
```

Make sure `$HOME/go/bin` (or `$GOPATH/bin`) is on your `$PATH` so the `gator` binary is accessible.

---

## Configuration

Gator reads its configuration from `~/.gatorconfig.json`. Create this file before running any commands:

```json
{
  "db_url": "postgres://user:password@localhost:5432/gator?sslmode=disable"
}
```

Replace `user`, `password`, and `gator` with your actual PostgreSQL credentials and database name. The `current_user_name` field is managed automatically when you `register` or `login`.

---

## Database Setup

Clone the repository to access the migration files, then run Goose to create the required tables:

```bash
git clone https://github.com/AvelarJ/Gator.git
cd Gator
goose -dir sql/schema postgres "postgres://user:password@localhost:5432/gator" up
```

This creates four tables: `users`, `feed`, `feed_follows`, and `posts`.

---

## Commands

### User Management

| Command | Usage | Description |
|---|---|---|
| `register` | `gator register <username>` | Create a new user and log in as them |
| `login` | `gator login <username>` | Switch to an existing user |
| `users` | `gator users` | List all registered users (current user shown with `*`) |

### Feed Management

| Command | Usage | Description |
|---|---|---|
| `addfeed` | `gator addfeed <name> <url>` | Add a new RSS feed and follow it |
| `feeds` | `gator feeds` | List all feeds in the database |
| `follow` | `gator follow <feed_url>` | Follow an existing feed |
| `following` | `gator following` | Show all feeds you currently follow |
| `unfollow` | `gator unfollow <feed_url>` | Unfollow a feed |

### Aggregation & Browsing

| Command | Usage | Description |
|---|---|---|
| `agg` | `gator agg <interval>` | Fetch all feeds repeatedly at the given interval (e.g. `30s`, `5m`) |
| `browse` | `gator browse [limit]` | Display recent posts from your followed feeds (default: 2) |

### Development

| Command | Usage | Description |
|---|---|---|
| `reset` | `gator reset` | **Wipes all data from the database.** For development and testing only — do not run in production. |

---

## Quick Start

```bash
# 1. Register a user
gator register alice

# 2. Add a feed (it will be automatically followed)
gator addfeed "Go Blog" "https://go.dev/blog/feed.atom"
gator addfeed "Hacker News" "https://news.ycombinator.com/rss"

# 3. Start aggregating in the background
gator agg 30s &

# 4. Browse your posts
gator browse 10
```
