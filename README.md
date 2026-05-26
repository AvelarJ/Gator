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
go install github.com/AvelarJ/Gator
```

Make sure `$HOME/go/bin` (or `$GOPATH/bin`) is on your `$PATH` so the `gator` binary is accessible.

# PostgreSQL

(Use `brew` for Mac and check https://learn.microsoft.com/en-us/windows/wsl/tutorials/wsl-database#install-postgresql for WSL/Linux)

Verify instal worked:
```bash
psql --version
```

NOTE: (Linux / WSL only) Update postgres password (I used 'postgres'):
```bash
sudo passwd postgres
```

Start the postgres server in the background:
Mac
```bash
psql postgres
```
WSL/Linus
```bash
sudo -u postgres psql
```

Create a new database called `gator`
```bash
CREATE DATABASE gator;
```

Set the user password (Linux / WSL only)
```bash
ALTER USER postgres PASSWORD 'postgres';
```

The Postgres database should now be ready!

---

## Configuration

Gator reads its configuration from `~/.gatorconfig.json`. Create this file before running any commands:

```json
{
  "db_url": "postgres://user:password@localhost:5432/gator?sslmode=disable"
}
```

Replace `user`, `password`, and `gator` with your actual PostgreSQL credentials and database name (WSL: `user` and `password` were both postgres for my simplicity). The `current_user_name` field is managed automatically when you `register` or `login`.

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
Gator register alice

# 2. Add a feed (it will be automatically followed)
Gator addfeed "Top stories - Google News" "https://news.google.com/rss"
Gator addfeed "Hacker News" "https://news.ycombinator.com/rss"

# 3. Start aggregating (NOTE ctrl + c to stop)
Gator agg 30s

# 4. Browse your posts
Gator browse 10
```

## Possible Additions to come

- Sorting and filtering for `browse`
- Add a `search` command
- Add bookmarking or like posts
- Add a TUI to allow selecting and viewing each post in a better format
