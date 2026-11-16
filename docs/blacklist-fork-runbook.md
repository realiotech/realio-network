# Blacklist fork operator runbook

The fork block is `BlacklistForkHeight`, currently **19,573,266**, defined in
[`app/migrations/forks.go`](../app/migrations/forks.go). The binary swap point is
the last committed block immediately before it: **19,573,265**.
Confirm the height against the source for the release you deploy.

## Before the fork

Existing nodes participating in the coordinated fork must reach committed block
19,573,265 on the old binary, then stop, swap to the fork binary, and restart.

A post-fork snapshot will be provided for late-joining and lagging nodes. These
nodes should restore that snapshot and start the fork binary normally; they do
not need to replay the fork or catch up on the old binary first.

## Choose the procedure by committed application height

| Local state | Procedure |
| --- | --- |
| Below 19,573,265, without the blacklist store, or joining later | Restore the provided post-fork snapshot at height 19,573,266 or later, then start the fork binary. |
| Exactly 19,573,265 | Stop, swap to the fork binary, and restart. The loader adds the blacklist store for the fork block. |
| At or above 19,573,266, with the blacklist store and completed fork migrations | Start the fork binary normally. The fork has already been applied and must not be rerun. |

Check the **committed application height**, not the latest height advertised by
peers or an uncommitted block. The provided snapshot must contain committed state
at height **19,573,266 or later**, including the `blacklist` store and completed
fork migrations. A pre-fork snapshot does not meet this requirement; installing
the fork binary and letting it sync forward from an earlier height does not work.

The fork binary mounts `blacklist` at startup, but adds the store only when the
last committed height is `BlacklistForkHeight - 1`. Starting it earlier against
an existing database without that store fails with an error such as:

```text
version of store blacklist mismatch root store's version; expected N got 0
```

If this happens, stop the node and follow the provided post-fork snapshot's
restore instructions before restarting with the fork binary.

## Restarts and launch configuration

A crash before the fork block commits leaves the application at the pre-fork
height. Restart with the fork binary; the store upgrade is reapplied. Once the
fork block commits, later restarts load the existing blacklist store normally.

Review stale `--unsafe-skip-upgrades` entries in operator launch configuration.
The blacklist store loader is registered independently of that flag because
this is a hardcoded fork. Skip heights still apply to the scheduled v1.6 store
upgrade; they do not bypass the blacklist fork. Nodes using the provided
post-fork snapshot have already passed the fork and need no store upgrade.
