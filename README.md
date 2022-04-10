# mess

Massive Execution Structured Server. Helps when your workload is a mess.

Exposes the same API that
[Swarming](https://chromium.googlesource.com/infra/luci/luci-py/+/HEAD/appengine/swarming/README.md)
exposes.

Work in progress.


## Goals

- Drop-in replacement for a Swarming server, change the URL and it's a go.
- Scale to 20000 bots as a single homed server.
- Resilient (versus high availability).
- Very fast shutdown and restart time. Target: <1min
  - Requires the service to be essentially stateless. It's already designed this
    way.

### Non goals

- Scaling to more than 20000 bots.
  - Use a cloud service if you need something this large, or more simply
    segregate your fleet in smaller worker pools.
- High availability. It's overrated.
  - Instead, the clients should retry HTTP failures for a minute. This way we
    can upgrade the server without any downtime.
- RBE-CAS implementation. There is already an
  [implementation](https://github.com/buchgr/bazel-remote) which I haven't tried
  yet.


## State as of e8c5b866d4fc

### Works

- Bot. Unmodified from
  [upstream](https://chromium.googlesource.com/infra/luci/luci-py/+/HEAD/appengine/swarming/swarming_bot/)!
  - `swarming_bot.zip` generation.
  - Bot code delivery: curl and run!
  - Bot versioning based on schema, host and port.
  - Bot self-update.
  - Bot events.
  - Connecting multiple bots.
  - "Last seen".
  - Deleting a bot is partially implemented.
- Web UI. Unmodified from
  [upstream](https://chromium.googlesource.com/infra/luci/luci-py/+/HEAD/appengine/swarming/ui2/)!
  - Dimensions prefill in /botlist and /tasklist
- Google Acccount OAuth2 based login!
- Two DB backends!
  - [Sqlite3](https://pkg.go.dev/github.com/mattn/go-sqlite3).
  - In-memory with JSON serialization.
- Structured logging with [zerolog](https://pkg.go.dev/github.com/rs/zerolog)!
- Server version generation based on [go1.18
  buildinfo](https://tip.golang.org/doc/go1.18#debug/buildinfo).
  - Includes "tainted" versioning when there's local modifications.
- HTTPS (fronted with caddy) or localhost.
- Primitive task scheduling.
- Primitive ACL.

### Not working

- Full task execution:
  - Bot is not able to send updates to a task.
  - Task stdout output support.
  - Terminating a bot and fully deleting it.
  - Service accounts for the bot.
- Task queues precomputation. Only unnecessary once >100 bots.
- Bot configuration.
  - Server injection dimensions, custom `bot_config.py`, pools.
- DB:
  - Queries with filters, e.g. bot counts are incorrect, /botlist and /tasklist
    do not take filters into effect.
  - Schema migration, albeit the design is preemptively defensive.
  - Task output as file system or external storage?
- LUCI integration
  - luci-config
  - buildbucket token
  - resultdb token
  - token server
- A few bugs here and there left in the client API.
  - swarming CLI tool doesn't like the server replies.
  - e.g. tasks for a bot are not showing up in the web ui yet.
- TLS server with
  [certmagic](https://pkg.go.dev/github.com/caddyserver/certmagic). Front with
  [Caddy](https://caddyserver.com/) in the meantime.
- Web UI doesn't understand the version when it tries to extract the "git
  revision". Need to fix upstream since it's hardcoded in the Web UI.
- Cleanup cron jobs
  - Marking bot as deleted.
  - Expiring tasks.
  - Data eviction, deleting old tasks and bots after 18 months (or less).
- Monitoring time series.
  - Should be trivial to compared on how hard it was on AppEngine.
- BigQuery export.
- DDoS protection.
- Injecting HTTP 500s randomly. I really want the client to be resilient.

## Usage

### Server

1. (Optional) Get a OAuth2 client id with [Swarming's APIs & Services > Credentials
   instructions](https://chromium.googlesource.com/infra/luci/luci-py/+/HEAD/appengine/swarming/README.md#setting-up).
2. (Optional) Get a domain. Install [Caddy](https://caddyserver.com/) and
   configure it with your domain. If so, remove `-local` from the example below.

```
go install github.com/maruel/mess/cmd/mess@latest
mess \
  -usr <yourself>@gmail.com,<friend>@google.com \
  -local \
  -cid 1111111111111-aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa.apps.googleusercontent.com
```

### Bot

One bot:

```
mkdir bot
cd bot
curl -lL -o swarming_bot.zip http://localhost:7899/bot_code
python3 swarming_bot.zip start_bot
```

Many bots:

```
go install github.com/maruel/mess/cmd/loadtestmess@latest
loadtestmess -num 10 -S http://localhost:7899
```

### CLI Client

It's not installable as-is so it requires a bit of fiddling:

```
git clone https://chromium.googlesource.com/infra/luci/luci-go
cd luci-go
# Comment out replace and exclude at the bottom:
vi go.mod
# Run time it will complain and tell to run commands, do it and retry.
go install ./client/cmd/swarming
swarming trigger -S http://localhost:7899 -d os=Ubuntu -d pool=default \
  -expiration 3600 \
  -- ls
```


## Ideas

Here's the vaporware ideas:

- Options if we need to scale up more (unlikely):
  - Offloading read-only data to append-only DB. Reduce the strain on primary
    DB.
  - Sidecar server that serves immutable data (as soon as the task is
    completed), reducing workload on the primary server.
  - Separate client API and bot API into separate servers.
    - A client DDoS wouldn't affect bots execution.
    - Reduces network connections per VM, probably can help 2x scale.
- Exposing the RBE API as a gRPC server. Provides an incremental path to migrate
  to RBE! Deprecating the funky Swarming CloudEndpoint API.
