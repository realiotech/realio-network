# Blacklist fork operator runbook

The fork block is `BlacklistForkHeight`, currently **19,573,266**, defined in
[`app/migrations/forks.go`](../app/migrations/forks.go). The binary swap point is
the last committed block immediately before it: **19,573,265**.
Confirm the height against the source for the release you deploy.

## Before the fork

Keep the old binary available for nodes that are offline, catching up, restoring
an older snapshot, or joining the network. Do not replace the only copy of it.
Arrange for the old binary to stop after committing block 19,573,265, including
when catching up after the rest of the network has already resumed.

## Choose the procedure by committed application height

| Local state | Procedure |
| --- | --- |
| Below 19,573,265, without the blacklist store | Run the old binary until block 19,573,265 is committed, then stop, swap to the fork binary, and restart. |
| Exactly 19,573,265 | Stop, swap to the fork binary, and restart. The loader adds the blacklist store for the fork block. |
| At or above 19,573,266, with the blacklist store | Start the fork binary normally. This also applies to snapshots and state sync that include the committed post-fork store. |

Check the **committed application height**, not the latest height advertised by
peers or an uncommitted block. A pre-fork snapshot or a newly joining node still
needs the old-binary catch-up procedure; installing the fork binary and letting
it sync forward from an earlier height does not work.

The fork binary mounts `blacklist` at startup, but adds the store only when the
last committed height is `BlacklistForkHeight - 1`. Starting it earlier against
an existing database without that store fails with an error such as:

```text
version of store blacklist mismatch root store's version; expected N got 0
```

If this happens, stop the node and use the old binary to finish catching up to
the swap point. Do not delete the database to resolve this error.

## Restarts and launch configuration

A crash before the fork block commits leaves the application at the pre-fork
height. Restart with the fork binary; the store upgrade is reapplied. Once the
fork block commits, later restarts load the existing blacklist store normally.

Review stale `--unsafe-skip-upgrades` entries in operator launch configuration.
The blacklist store loader is registered independently of that flag because
this is a hardcoded fork. Skip heights still apply to the scheduled v1.6 store
upgrade; they do not bypass the blacklist fork or its required binary swap point.
