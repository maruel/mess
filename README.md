# mess

Massive Execution Structured Server.

Work in progress.

## Goals

- Scale to 20000 bots as a single homed server.
- Resilient (versus high availability).

mess has a non-goal scaling more than ~20000 bots. Use a cloud service if you
need something this large, or more simply segregate your fleet in smaller worker
pools.

mess has a non-goal high availability. Instead, the clients should retry HTTP
failures for a minute, and the server is designed to restart within seconds.

mess only implements the task execution where RBE protocol is lacking, it is not
implementing the RBE-CAS protocol. https://github.com/buchgr/bazel-remote seems
like a good RBE-CAS implementation but I haven't tried it yet.
