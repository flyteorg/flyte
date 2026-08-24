# CloudEvents egress — design

Status: proposal
Scope: `runs/`
Tracks: flyteorg/flyte#7829

## What is missing

Flyte 1 pushed every execution event to a message broker. Systems outside Flyte — lineage
catalogs, alerting, cost accounting, pipelines in other orchestrators — subscribed to a topic
and reacted. They never talked to Flyte, and Flyte never knew they existed.

Flyte 2 records more events than v1 did, in a queryable table, with a live stream on top. What
it does not do is push them anywhere. Every consumer must now hold a long-lived gRPC stream
against the control plane, which makes each of them a client Flyte has to serve, keep
connected, and survive restarts with.

```
v1
+------------+       +--------------+       +---------------------+
| flyteadmin |------>| broker       |------>| lineage / catalog   |
+------------+ push  |              |       +---------------------+
                     | durable      |       +---------------------+
                     | replayable   |------>| alerting            |
                     | fans out     |       +---------------------+
                     |              |       +---------------------+
                     |              |------>| downstream pipeline |
                     +--------------+       +---------------------+
                                            consumers are decoupled;
                                            admin never knows they exist

v2 proposed          [NEW] = added by this document; everything else exists today

+--------------+
| runs service |   every action change already does BOTH of these
+------+-------+
       |
       +----------------------------------+
       |                                  |
       v                                  v
+--------------------------+   +---------------------------------+
| action_events            |   | notifyActionUpdate()            |
|   the durable record     |   |   -> NOTIFY  action_updates     |
|   runs/repository/impl/  |   |   -> actionSubscribers          |
|     action.go            |   |      channel full -> DROPPED    |
+------------+-------------+   +----------------+----------------+
             |                                  |
             |                                  +---> +------------------+
             |                                  |     | watch streams    |
             |                                  |     | console / CLI    |
             |                                  |     | existing sub     |
             |                                  |     +------------------+
             | READ -- authoritative            |
             | every published event comes      | HINT -- best effort
             | from here, survives a restart    | id only, no content
             |                                  | lost on restart or
             |                                  | on channel overflow
             |                                  |
             |                                  |    ATTACH THE PUBLISHER
             |                                  |    HERE, as one more
             |                                  |    subscriber
             v                                  v
      +--------------------------------------------------+
      | publisher                                  [NEW] |
      |   Flyte event --> CloudEvent                     |
      |   knows how far it has published                 |
      +------------------------+-------------------------+
                               |
                               v
      +--------------------------------------------------+
      | sender                                     [NEW] |
      |   one per transport                              |
      +------------------------+-------------------------+
                               |
                               v
              +---------------------+     +---------------------+
              | broker        [NEW] |---->| lineage / catalog   |
              |                     |     +---------------------+
              | operator-run,       |     +---------------------+
              | not shipped by      |---->| alerting            |
              | Flyte               |     +---------------------+
              |                     |     +---------------------+
              |                     |---->| downstream pipeline |
              +---------------------+     +---------------------+
                                          new consumers land here
```

## What carries over from v1

v1 split the feature along a seam worth keeping: a **publisher** that turns a Flyte event into
a CloudEvent, and a **sender** that puts a CloudEvent on a wire. The publisher knows the Flyte
domain and nothing about brokers; the sender knows brokers and nothing about Flyte.

```
+--------------------------------------+
| PUBLISHER                            |
|   knows Flyte events                 |
|   decides what an event looks like    |
|   on the wire: type, id, time, data  |
+------------------+-------------------+
                   |
                   v
+--------------------------------------+
|          one narrow interface        |  <-- the seam
+------------------+-------------------+
                   |
      +------------+------------+------------+
      v            v            v            v
+-----------+ +---------+ +-----------+ +---------+
| Kafka     | | NATS    | | cloud     | | no-op   |
| sender    | | sender  | | pub/sub   | |         |
+-----+-----+ +----+----+ +-----+-----+ +---------+
      v            v            v
+-----------+ +---------+ +-----------+
| operator's brokers -- not shipped or run by Flyte  |
+---------------------------------------------------+
```

**The sender half transfers almost unchanged.** Four transports, one interface, and the
envelope conventions that go with them — a stable id that doubles as the consumer's
deduplication key, a payload encoding that survives protobuf `oneof` fields, a schema
reference that lets a consumer validate without asking Flyte. None of that depends on the
Flyte data model, and it was working in production for years.

**The publisher half does not transfer.** v1's is written against workflow, node and task
executions. v2 has runs and actions. Mapping one onto the other is the substantive design work
here, and it is what #7829 asks to be documented.

## Turning it on

Config only, resolved once at startup. No per-run parameter, no launch-time opt-in, nothing in
the SDK surface: an operator enables egress for a deployment, and every event flows.

This mirrors v1, and the shape is worth repeating for two reasons. Operators already know it.
And the alternative — letting individual runs choose — makes the event stream unreliable as a
source of truth, because a consumer can no longer assume that silence means nothing happened.

Two properties the config gate should preserve:

- **Off by default.** A deployment that says nothing about egress publishes nothing.
- **Filterable.** An operator who wants only terminal events should be able to say so without
  filtering client-side, because the cost of the events they do not want is paid on the wire.
