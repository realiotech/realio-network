<!--
order: 1
-->

# Concepts

## The Realio Asset Token Model

The Realio Asset module is centeredd aroumd a token model. It contains the following fields:

```protobuf
message Token {
  string name = 1;
  string symbol = 2;
  int64 total = 3;
  int64 decimals = 4;
  bool authorizationRequired = 5;
  string creator = 6;
  map<string, TokenAuthorization> authorized = 7;
  int64 created = 8;
}

```

### Token Authorization

The `Token` model provides a means to whitelist users via the `authorizationRequired` and `authorized` fields
A token that has the `authorizationRequired` turned on, can maintain a whitelist map of user addresses. These addresses
are the only ones able to send/receive the token. The Realio Network is agnostic to the logic of applications that use
the whitelisting. It is up to the clients to determine when to whitelist and what to do with it. 



