<!--
order: 7
-->

# Events

The `x/asset` module emits the following events:

## Create new token

| Type            | Attribute Key | Attribute Value |
| --------------- |---------------|-----------------|
| `create_token` | `"amount"`    | `{totatl}`      |
| `create_token` | `"symbol"`    | `{symbol}`      |

## Update token

| Type            | Attribute Key | Attribute Value |
| --------------- |---------------|-----------------|
| `update_token` | `"symbol"`    | `{symbol}`      |


## Authorize address

| Type            | Attribute Key | Attribute Value  |
| --------------- |--------------|------------------|
| `authorize_token` | `"symbol"`   | `{symbol}`       |
| `authorize_token` | `"address"`  | `{sdk_address}`  |


## Un Authorize address

| Type            | Attribute Key | Attribute Value |
| --------------- |---------------|-----------------|
| `unauthorize_token` | `"symbol"`    | `{symbol}`      |
| `unauthorize_token` | `"address"`   | `{sdk_address}` |

