# JSON Web Tokens (`jwt`)

Decode JSON Web Tokens, and sign or verify them (HS, RS, ES and EdDSA).

No permission is needed: these are pure functions. Generated from the tools registry.

4 functions. Use them as `jwt.<name>(...)` in any extension.

## jwt.decode

`jwt.decode(token: str) -> Any`

Split a JWT into a dict with header, payload (both dicts), signature (the base64url text) and alg. The signature is NOT checked: never trust the payload of a token you have not passed to jwt.verify. Raises when the token does not have three dot-separated base64url parts holding JSON objects.

| Parameter | Type | Description |
|---|---|---|
| `token` | `str` | The compact JWT; a leading "Bearer " is ignored. |

Returns `Any`. Example: `jwt.decode("eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjMifQ.sig")`

Keywords: read jwt, inspect token, jwt claims, jwt payload, bearer token, id token, access token, jwt debugger, parse jwt, jwt, decode, token, claims, inspect

## jwt.is_valid

`jwt.is_valid(token: str, key: str, alg: str = "", check_time: bool = True, leeway: int = 0, now: int = 0, key_encoding: str = "utf-8") -> bool`

Report whether a JWT passes jwt.verify with these arguments, as true or false instead of an error.

| Parameter | Type | Description |
|---|---|---|
| `token` | `str` | The compact JWT. |
| `key` | `str` | Shared secret for HS*, or a PEM public key, certificate or private key. |
| `alg` | `str` (default `""`) | The only algorithm to accept; empty accepts the token's alg. One of: HS256, HS384, HS512, RS256, RS384, RS512, PS256, PS384, PS512, ES256, ES384, ES512, EdDSA. |
| `check_time` | `bool` (default `True`) | Check exp and nbf. |
| `leeway` | `int` (default `0`) | Seconds of tolerance for exp and nbf. |
| `now` | `int` (default `0`) | Unix time to compare against; 0 reads the clock. |
| `key_encoding` | `str` (default `"utf-8"`) | How an HS* secret is written; PEM keys ignore it. One of: utf-8, hex, base64, base64url. |

Returns `bool`. Example: `jwt.is_valid("a.b.c", "secret")`

Keywords: token valid, check jwt, signature ok, is expired, jwt, valid, verify, check

## jwt.sign

`jwt.sign(payload: Any, key: str, alg: str = "HS256", header: Any = None, expires_in: int = 0, issued_at: bool = False, now: int = 0, key_encoding: str = "utf-8") -> str`

Create a signed JWT from a payload dict. alg is one of HS256/384/512 (key is the shared secret), RS256/384/512 and PS256/384/512 (key is an RSA private key in PEM), ES256/384/512 (EC private key in PEM, curve must match) or EdDSA (Ed25519 private key in PEM). expires_in adds exp, issued_at adds iat, both from now (0 means the current time). header adds or overrides header fields; alg and typ are always set. Raises on a key that does not fit the algorithm or a payload that is not a dict.

| Parameter | Type | Description |
|---|---|---|
| `payload` | `Any` | The claims, a dict. |
| `key` | `str` | Shared secret for HS*, or a private key in PEM for the others. |
| `alg` | `str` (default `"HS256"`) | The signing algorithm. One of: HS256, HS384, HS512, RS256, RS384, RS512, PS256, PS384, PS512, ES256, ES384, ES512, EdDSA. |
| `header` | `Any` (default `None`) | Extra header fields, a dict. |
| `expires_in` | `int` (default `0`) | Seconds until the token expires; 0 adds no exp. |
| `issued_at` | `bool` (default `False`) | Add the iat claim. |
| `now` | `int` (default `0`) | Unix time to use for iat and exp; 0 reads the clock. |
| `key_encoding` | `str` (default `"utf-8"`) | How an HS* secret is written; PEM keys ignore it. One of: utf-8, hex, base64, base64url. |

Returns `str`. Example: `jwt.sign({"sub": "123", "name": "Ada"}, "secret", expires_in=3600)`

Keywords: create jwt, generate token, make jwt, encode jwt, jws, hs256, rs256, access token, issue token, jwt, sign, token, create

## jwt.verify

`jwt.verify(token: str, key: str, alg: str = "", check_time: bool = True, leeway: int = 0, now: int = 0, key_encoding: str = "utf-8") -> Any`

Check a JWT's signature and time claims and return what jwt.decode returns. key is the shared secret for HS*, or a public key, certificate or private key in PEM for the others. alg pins the accepted algorithm (strongly advised); empty trusts the token's own header, but "none" is always rejected and an HS* token is refused when the key is PEM, which blocks the key-confusion attack. exp and nbf are checked with leeway seconds of slack (check_time=False skips them). Raises with the reason when anything fails.

| Parameter | Type | Description |
|---|---|---|
| `token` | `str` | The compact JWT. |
| `key` | `str` | Shared secret for HS*, or a PEM public key, certificate or private key. |
| `alg` | `str` (default `""`) | The only algorithm to accept; empty accepts the token's alg. One of: HS256, HS384, HS512, RS256, RS384, RS512, PS256, PS384, PS512, ES256, ES384, ES512, EdDSA. |
| `check_time` | `bool` (default `True`) | Check exp and nbf. |
| `leeway` | `int` (default `0`) | Seconds of tolerance for exp and nbf. |
| `now` | `int` (default `0`) | Unix time to compare against; 0 reads the clock. |
| `key_encoding` | `str` (default `"utf-8"`) | How an HS* secret is written; PEM keys ignore it. One of: utf-8, hex, base64, base64url. |

Returns `Any`. Example: `jwt.verify(jwt.sign({"sub": "1"}, "secret"), "secret", alg="HS256")`

Keywords: validate jwt, check token, verify signature, jwt valid, authenticate token, is token expired, check signature, jwt, verify, validate, signature, check

