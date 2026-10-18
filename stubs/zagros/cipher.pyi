"""Encryption module (generated from the tools registry; do not edit by hand).

AES, ChaCha20, RSA, Ed25519, XOR and classic ciphers, each with a decrypt counterpart.
Needs: none. These are pure functions: no file, network, clipboard or environment access.
"""

from typing import Dict, List

def aes_gcm_encrypt(text: str, key: str, key_encoding: str = "passphrase", bits: int = 0, iv: str = "", iv_encoding: str = "hex", aad: str = "", input_encoding: str = "utf-8", output: str = "base64") -> str:
    """Encrypts text with AES-GCM. Authenticated (AEAD): a wrong key or tampered data fails with an error. 12-byte nonce; never reuse a nonce with the same key. With key_encoding=passphrase (default) the key is a passphrase stretched by scrypt (N=32768) with a random 16-byte salt and the result is the text envelope zg1.<algo>.<base64url(salt|nonce|ciphertext|tag)>, which aes_gcm_decrypt reads back; with key_encoding utf-8, hex or base64 the key is raw and the result is base64(iv|ciphertext|tag) rendered per output (a random IV unless iv is given).

        text: Plaintext (UTF-8 unless input_encoding says hex or base64).
        key: Passphrase (key_encoding=passphrase, the default) or the raw key (AES key must be 16, 24 or 32 bytes.).
        key_encoding: passphrase (default): key is a password stretched by scrypt with a random salt, output is a zg1 envelope. utf-8, hex or base64: key is the raw key bytes in that encoding. One of: passphrase, utf-8, hex, base64.
        bits: AES key size: 128, 192 or 256. 0 (default) means 256 for passphrases and the key's own length for raw keys. Ignored by non-AES ciphers.
        iv: Optional IV/nonce, for deterministic output or interop (random by default). Encoded per iv_encoding. On decrypt with a raw key it means the data has no IV prefix.
        iv_encoding: How iv is turned into bytes: hex (default), base64 or utf-8. One of: hex, base64, utf-8.
        aad: Optional additional authenticated data (UTF-8), used by the AEAD ciphers (GCM, ChaCha20-Poly1305) and ignored by the others; it must be the same on decrypt.
        input_encoding: How text is turned into bytes: utf-8 (default), hex or base64. One of: utf-8, hex, base64.
        output: Raw-key mode only: how iv|ciphertext|tag is rendered (base64 default). Passphrase mode always returns the zg1 envelope. One of: base64, base64url, hex, HEX, base32, utf-8.

    Example: cipher.aes_gcm_encrypt("secret message", "correct horse battery staple")
    Needs: none (pure function).
    """

def algorithms() -> List[str]:
    """Returns the names accepted by encrypt() and decrypt() (symmetric ciphers), e.g. aes_gcm, chacha20_poly1305. The classic ciphers (caesar, vigenere, ...), xor_cipher and the rsa/ed25519/ecdsa functions have their own functions.

    Example: cipher.algorithms()
    Needs: none (pure function).
    """

def ecdsa_generate(curve: str = "p256") -> Dict[str, str]:
    """Generates an ECDSA key pair on curve p256 (default, a.k.a. secp256r1/prime256v1), p384 or p521 and returns a dict with private_pem (PKCS#8) and public_pem (PKIX). Random every call.

        curve: Curve (default p256). One of: p256, p384, p521.

    Example: cipher.ecdsa_generate("p256")
    Needs: none (pure function).
    """

def ecdsa_sign(text: str, private_pem: str, hash: str = "sha256", input_encoding: str = "utf-8", output: str = "base64") -> str:
    """Signs text with an ECDSA private key (PEM) over the SHA-256 (or hash) digest and returns the ASN.1 DER signature, base64 by default. Randomized: two calls give different valid signatures.

        text: Message to sign.
        private_pem: ECDSA private key in PEM.
        hash: Hash function (default sha256). One of: sha256, sha384, sha512, sha1.
        input_encoding: How text is turned into bytes: utf-8 (default), hex or base64. One of: utf-8, hex, base64.
        output: Signature format (base64 default). One of: base64, base64url, hex, HEX, base32, utf-8.

    Example: cipher.ecdsa_sign("message", private_pem)
    Needs: none (pure function).
    """

def ecdsa_verify(text: str, signature: str, public_pem: str, hash: str = "sha256", input_encoding: str = "utf-8", signature_encoding: str = "base64") -> bool:
    """Verifies an ASN.1 DER ECDSA signature (base64 by default) of text against a public key PEM and returns true or false. Malformed keys are errors; a bad signature is false.

        text: Signed message.
        signature: DER signature in signature_encoding.
        public_pem: ECDSA public key in PEM.
        hash: Hash function (default sha256). One of: sha256, sha384, sha512, sha1.
        input_encoding: How text is turned into bytes: utf-8 (default), hex or base64. One of: utf-8, hex, base64.
        signature_encoding: Encoding of signature (base64 default). One of: base64, hex.

    Example: cipher.ecdsa_verify("message", signature_b64, public_pem)
    Needs: none (pure function).
    """

def ed25519_generate() -> Dict[str, str]:
    """Generates an Ed25519 key pair and returns a dict: private_key (32-byte seed, hex), public_key (hex), private_pem (PKCS#8) and public_pem (PKIX). Random every call.

    Example: cipher.ed25519_generate()
    Needs: none (pure function).
    """

def ed25519_public_key(private_key: str) -> str:
    """Derives the Ed25519 public key (hex) from a private key given as a 32-byte hex seed, a 64-byte hex key or a PEM.

        private_key: Private key: 32-byte seed in hex, 64-byte hex key or PEM.

    Example: cipher.ed25519_public_key("9d61b19deffd5a60ba844af492ec2cc44449c5697b326919703bac031cae7f60")
    Needs: none (pure function).
    """

def ed25519_sign(text: str, private_key: str, input_encoding: str = "utf-8", output: str = "hex") -> str:
    """Signs text with an Ed25519 private key (32-byte hex seed, 64-byte hex key or PEM) and returns the 64-byte signature, hex by default. Ed25519 signatures are deterministic.

        text: Message to sign.
        private_key: Private key: 32-byte seed in hex, 64-byte hex key or PEM.
        input_encoding: How text is turned into bytes: utf-8 (default), hex or base64. One of: utf-8, hex, base64.
        output: Signature format (hex default). One of: base64, base64url, hex, HEX, base32, utf-8.

    Example: cipher.ed25519_sign("message", "9d61b19deffd5a60ba844af492ec2cc44449c5697b326919703bac031cae7f60")
    Needs: none (pure function).
    """

def ed25519_verify(text: str, signature: str, public_key: str, input_encoding: str = "utf-8", signature_encoding: str = "hex") -> bool:
    """Verifies an Ed25519 signature (hex by default) of text against a public key (32-byte hex or PEM) and returns true or false. Malformed keys are errors; a wrong signature is false.

        text: Signed message.
        signature: Signature in signature_encoding (hex default).
        public_key: Public key: 32-byte hex or PEM.
        input_encoding: How text is turned into bytes: utf-8 (default), hex or base64. One of: utf-8, hex, base64.
        signature_encoding: Encoding of signature (hex default). One of: hex, base64.

    Example: cipher.ed25519_verify("message", signature_hex, public_key_hex)
    Needs: none (pure function).
    """

def pem_encode(type: str, hex: str) -> str:
    """Wraps hex bytes in a PEM block of the given type (e.g. 'CERTIFICATE', 'PUBLIC KEY', 'PRIVATE KEY') with 64-column base64 lines. Odd or non-hex input and an empty type are errors.

        type: PEM label, e.g. CERTIFICATE or PUBLIC KEY.
        hex: DER bytes as hex.

    Example: cipher.pem_encode("PUBLIC KEY", "3059301306...")
    Needs: none (pure function).
    """

def rsa_generate(bits: int = 2048) -> Dict[str, str]:
    """Generates an RSA key pair and returns a dict with public_pem (PKIX), private_pem (PKCS#8) and bits. bits must be 2048 (default), 3072 or 4096. Slow: 2048 takes about a second, 4096 can take several seconds to a minute; the private key is random every call.

        bits: Key size: 2048 (default), 3072 or 4096. One of: 2048, 3072, 4096.

    Example: cipher.rsa_generate(2048)
    Needs: none (pure function).
    """

def rsa_sign(text: str, private_pem: str, padding: str = "pss", hash: str = "sha256", input_encoding: str = "utf-8", output: str = "base64") -> str:
    """Signs text with an RSA private key (PEM), using RSASSA-PSS (default, salt length = hash length) or PKCS#1 v1.5, over the SHA-256 (or hash) digest. Output is base64 by default.

        text: Message to sign.
        private_pem: RSA private key in PEM.
        padding: Signature scheme: pss (RSASSA-PSS, default) or pkcs1v15 (RSASSA-PKCS1-v1_5). One of: pss, pkcs1v15.
        hash: Hash function (default sha256). One of: sha256, sha384, sha512, sha1.
        input_encoding: How text is turned into bytes: utf-8 (default), hex or base64. One of: utf-8, hex, base64.
        output: Signature format (default base64). One of: base64, base64url, hex, HEX, base32, utf-8.

    Example: cipher.rsa_sign("message", private_pem)
    Needs: none (pure function).
    """

def rsa_verify(text: str, signature: str, public_pem: str, padding: str = "pss", hash: str = "sha256", input_encoding: str = "utf-8", signature_encoding: str = "base64") -> bool:
    """Verifies an RSA signature (PSS by default, or pkcs1v15) of text against a public key PEM and returns true or false. A malformed key or signature encoding is an error; a bad signature is false.

        text: Signed message.
        signature: Signature in signature_encoding (base64 default).
        public_pem: RSA public key in PEM.
        padding: Signature scheme: pss (RSASSA-PSS, default) or pkcs1v15 (RSASSA-PKCS1-v1_5). One of: pss, pkcs1v15.
        hash: Hash function (default sha256). One of: sha256, sha384, sha512, sha1.
        input_encoding: How text is turned into bytes: utf-8 (default), hex or base64. One of: utf-8, hex, base64.
        signature_encoding: Encoding of signature (default base64). One of: base64, hex.

    Example: cipher.rsa_verify("message", signature_b64, public_pem)
    Needs: none (pure function).
    """

def x25519_generate() -> Dict[str, str]:
    """Generates an X25519 (Curve25519 Diffie-Hellman) key pair and returns a dict with private_key and public_key as 32-byte hex. Random every call.

    Example: cipher.x25519_generate()
    Needs: none (pure function).
    """

def x25519_shared_secret(private_key: str, peer_public_key: str) -> str:
    """Computes the X25519 Diffie-Hellman shared secret (32 bytes, hex) from your private key and the peer's public key, both 32-byte hex. Both sides derive the same value. All-zero (low-order) peer keys are rejected.

        private_key: Your private key, 32 bytes hex.
        peer_public_key: The peer's public key, 32 bytes hex.

    Example: cipher.x25519_shared_secret(my_private_hex, peer_public_hex)
    Needs: none (pure function).
    """

def aes_gcm_decrypt(data: str, key: str, key_encoding: str = "passphrase", bits: int = 0, iv: str = "", iv_encoding: str = "hex", aad: str = "", data_encoding: str = "base64", output: str = "utf-8") -> str:
    """Decrypts data produced by aes_gcm_encrypt (AES-GCM). Accepts the zg1 envelope (with key_encoding=passphrase) or raw iv|ciphertext|tag data in data_encoding (default base64) with a raw key. Authenticated: tampering is always detected. A wrong key, a wrong passphrase or damaged data returns an error; the plaintext is returned as UTF-8 text unless output asks for hex or base64.

        data: Ciphertext: a zg1 envelope, or raw iv|ciphertext|tag in data_encoding.
        key: Passphrase (key_encoding=passphrase, the default) or the raw key (AES key must be 16, 24 or 32 bytes.).
        key_encoding: passphrase (default): key is a password stretched by scrypt with a random salt, output is a zg1 envelope. utf-8, hex or base64: key is the raw key bytes in that encoding. One of: passphrase, utf-8, hex, base64.
        bits: AES key size: 128, 192 or 256. 0 (default) means 256 for passphrases and the key's own length for raw keys. Ignored by non-AES ciphers.
        iv: Optional IV/nonce, for deterministic output or interop (random by default). Encoded per iv_encoding. On decrypt with a raw key it means the data has no IV prefix.
        iv_encoding: How iv is turned into bytes: hex (default), base64 or utf-8. One of: hex, base64, utf-8.
        aad: Optional additional authenticated data (UTF-8), used by the AEAD ciphers (GCM, ChaCha20-Poly1305) and ignored by the others; it must be the same on decrypt.
        data_encoding: Encoding of data in raw-key mode (default base64). Ignored for zg1 envelopes. One of: base64, base64url, hex, base32.
        output: How the plaintext is returned: utf-8 text (default), hex, HEX, base64, base64url or base32. One of: base64, base64url, hex, HEX, base32, utf-8.

    Example: cipher.aes_gcm_decrypt("zg1.aes256gcm....", "correct horse battery staple")
    Needs: none (pure function).
    """

def xor_cipher(text: str, key: str, input_encoding: str = "utf-8", key_encoding: str = "utf-8", output: str = "hex") -> str:
    """XORs text with a repeating key and returns hex by default (output hex, base64, utf-8, ...). XOR is its own inverse: to decrypt, pass the hex result back with input_encoding=hex and output=utf-8, or use xor_decrypt. An empty key is an error. Not secure encryption.

        text: Data to XOR (UTF-8 unless input_encoding says hex or base64).
        key: Repeating key (UTF-8 unless key_encoding says hex or base64); must not be empty.
        input_encoding: How text is turned into bytes: utf-8 (default), hex or base64. One of: utf-8, hex, base64.
        key_encoding: How the key is turned into bytes: utf-8 (default), hex or base64. One of: utf-8, hex, base64.
        output: Output format: hex (default), HEX, base64, base64url, base32 or utf-8 (raw bytes as text). One of: base64, base64url, hex, HEX, base32, utf-8.

    Example: cipher.xor_cipher("hello", "key")
    Needs: none (pure function).
    """

def xor_decrypt(data: str, key: str, data_encoding: str = "hex", key_encoding: str = "utf-8", output: str = "utf-8") -> str:
    """Reverses xor_cipher: decodes data (hex by default), XORs it with the repeating key and returns UTF-8 text by default. An empty key is an error; a wrong key just gives different bytes since XOR has no integrity check.

        data: XOR-ed data in data_encoding (hex by default).
        key: Repeating key (UTF-8 unless key_encoding says hex or base64).
        data_encoding: Encoding of data (default hex). One of: hex, base64, utf-8.
        key_encoding: How the key is turned into bytes: utf-8 (default), hex or base64. One of: utf-8, hex, base64.
        output: Output format: utf-8 text (default), hex, HEX, base64, base64url or base32. One of: base64, base64url, hex, HEX, base32, utf-8.

    Example: cipher.xor_decrypt("03000a0c0e", "key")
    Needs: none (pure function).
    """

def xor_brute_single_byte(hex: str, limit: int = 5) -> List[str]:
    """Tries all 256 single-byte XOR keys on hex data and returns the best `limit` candidates (default 5) ranked by English letter-frequency score, each as '0xKK (c) score=S | plaintext'. Non-printable results are penalised. Works best on English text of 10+ characters.

        hex: Ciphertext as hex (spaces, colons and a 0x prefix are ignored).
        limit: How many candidates to return, 1 to 256 (default 5).

    Example: cipher.xor_brute_single_byte("1b37373331363f78151b7f2b783431333d78397828372d363c78373e783a393b3736")
    Needs: none (pure function).
    """

def caesar(text: str, shift: int = 3) -> str:
    """Shifts each letter forward by shift places (default 3, any integer; negative shifts go backwards, the shift wraps modulo 26). Letters keep their case; digits, spaces and punctuation pass through unchanged. Use caesar_decrypt to undo or caesar_brute to try all shifts.

        text: Text to process.
        shift: Places to shift (default 3); negative shifts backwards.

    Example: cipher.caesar("Hello, World!", 3)
    Needs: none (pure function).
    """

def rot13(text: str) -> str:
    """Rotates every letter by 13 places (A to N). Letters keep their case; digits, spaces and punctuation pass through unchanged. It is its own inverse: applying it twice returns the original text.

        text: Text to process.

    Example: cipher.rot13("Hello")
    Needs: none (pure function).
    """

def classic_all(text: str) -> Dict[str, str]:
    """Applies the parameter-free classic ciphers to the text at once and returns a dict: rot13, rot5, rot18, rot47, atbash, caesar_1, caesar_2, caesar_3 (shifts forward), reverse (characters reversed). Handy for quickly eyeballing a CTF string; use caesar_brute for all shifts.

        text: Text to process.

    Example: cipher.classic_all("Hello, World!")
    Needs: none (pure function).
    """

def encrypt(algo: str, text: str, key: str, key_encoding: str = "passphrase", bits: int = 0, iv: str = "", iv_encoding: str = "hex", aad: str = "", input_encoding: str = "utf-8", output: str = "base64") -> str:
    """Encrypts text with the named symmetric cipher (aes_gcm, aes_cbc, aes_ctr, aes_cfb, aes_ofb, aes_ecb, chacha20_poly1305, xchacha20_poly1305, chacha20, des3_cbc, blowfish_cbc, rc4; see algorithms()). Passphrase mode (default) returns the text envelope zg1.<algo>.<base64url(salt|nonce|ciphertext|tag)> using an scrypt-derived key; raw-key mode (key_encoding utf-8, hex or base64) returns base64(iv|ciphertext|tag).

        algo: Cipher name, e.g. aes_gcm or chacha20_poly1305.
        text: Plaintext (UTF-8 unless input_encoding says hex or base64).
        key: Passphrase (key_encoding=passphrase, the default) or the raw key (key length depends on algo).
        key_encoding: passphrase (default): key is a password stretched by scrypt with a random salt, output is a zg1 envelope. utf-8, hex or base64: key is the raw key bytes in that encoding. One of: passphrase, utf-8, hex, base64.
        bits: AES key size: 128, 192 or 256. 0 (default) means 256 for passphrases and the key's own length for raw keys. Ignored by non-AES ciphers.
        iv: Optional IV/nonce, for deterministic output or interop (random by default). Encoded per iv_encoding. On decrypt with a raw key it means the data has no IV prefix.
        iv_encoding: How iv is turned into bytes: hex (default), base64 or utf-8. One of: hex, base64, utf-8.
        aad: Optional additional authenticated data (UTF-8), used by the AEAD ciphers (GCM, ChaCha20-Poly1305) and ignored by the others; it must be the same on decrypt.
        input_encoding: How text is turned into bytes: utf-8 (default), hex or base64. One of: utf-8, hex, base64.
        output: Raw-key mode only: rendering of iv|ciphertext|tag (base64 default). One of: base64, base64url, hex, HEX, base32, utf-8.

    Example: cipher.encrypt("aes_gcm", "secret message", "correct horse battery staple")
    Needs: none (pure function).
    """

def decrypt(data: str, key: str, key_encoding: str = "passphrase", algo: str = "", bits: int = 0, iv: str = "", iv_encoding: str = "hex", aad: str = "", data_encoding: str = "base64", output: str = "utf-8") -> str:
    """Decrypts data from encrypt() or any <cipher>_encrypt. For a zg1 envelope the algorithm is read from the envelope, so algo can be omitted; raw-key data needs algo and a key_encoding of utf-8, hex or base64. A wrong key or damaged data is an error (authenticated ciphers and all passphrase envelopes), never silent garbage.

        data: Ciphertext: a zg1 envelope, or raw iv|ciphertext|tag in data_encoding.
        key: Passphrase (key_encoding=passphrase, the default) or the raw key (key length depends on algo).
        key_encoding: passphrase (default): key is a password stretched by scrypt with a random salt, output is a zg1 envelope. utf-8, hex or base64: key is the raw key bytes in that encoding. One of: passphrase, utf-8, hex, base64.
        algo: Cipher name; optional for zg1 envelopes.
        bits: AES key size: 128, 192 or 256. 0 (default) means 256 for passphrases and the key's own length for raw keys. Ignored by non-AES ciphers.
        iv: Optional IV/nonce, for deterministic output or interop (random by default). Encoded per iv_encoding. On decrypt with a raw key it means the data has no IV prefix.
        iv_encoding: How iv is turned into bytes: hex (default), base64 or utf-8. One of: hex, base64, utf-8.
        aad: Optional additional authenticated data (UTF-8), used by the AEAD ciphers (GCM, ChaCha20-Poly1305) and ignored by the others; it must be the same on decrypt.
        data_encoding: Encoding of data in raw-key mode (default base64). Ignored for zg1 envelopes. One of: base64, base64url, hex, base32.
        output: How the plaintext is returned: utf-8 text (default), hex, HEX, base64, base64url or base32. One of: base64, base64url, hex, HEX, base32, utf-8.

    Example: cipher.decrypt("zg1.aes256gcm....", "correct horse battery staple")
    Needs: none (pure function).
    """

def aes_cbc_encrypt(text: str, key: str, key_encoding: str = "passphrase", bits: int = 0, iv: str = "", iv_encoding: str = "hex", input_encoding: str = "utf-8", output: str = "base64") -> str:
    """Encrypts text with AES-CBC. PKCS#7 padding, 16-byte IV. Unauthenticated in raw-key mode (tampering is not detected; a wrong key usually fails with a padding error); passphrase envelopes add an HMAC-SHA256 tag. With key_encoding=passphrase (default) the key is a passphrase stretched by scrypt (N=32768) with a random 16-byte salt and the result is the text envelope zg1.<algo>.<base64url(salt|nonce|ciphertext|tag)>, which aes_cbc_decrypt reads back; with key_encoding utf-8, hex or base64 the key is raw and the result is base64(iv|ciphertext|tag) rendered per output (a random IV unless iv is given).

        text: Plaintext (UTF-8 unless input_encoding says hex or base64).
        key: Passphrase (key_encoding=passphrase, the default) or the raw key (AES key must be 16, 24 or 32 bytes.).
        key_encoding: passphrase (default): key is a password stretched by scrypt with a random salt, output is a zg1 envelope. utf-8, hex or base64: key is the raw key bytes in that encoding. One of: passphrase, utf-8, hex, base64.
        bits: AES key size: 128, 192 or 256. 0 (default) means 256 for passphrases and the key's own length for raw keys. Ignored by non-AES ciphers.
        iv: Optional IV/nonce, for deterministic output or interop (random by default). Encoded per iv_encoding. On decrypt with a raw key it means the data has no IV prefix.
        iv_encoding: How iv is turned into bytes: hex (default), base64 or utf-8. One of: hex, base64, utf-8.
        input_encoding: How text is turned into bytes: utf-8 (default), hex or base64. One of: utf-8, hex, base64.
        output: Raw-key mode only: how iv|ciphertext|tag is rendered (base64 default). Passphrase mode always returns the zg1 envelope. One of: base64, base64url, hex, HEX, base32, utf-8.

    Example: cipher.aes_cbc_encrypt("secret message", "correct horse battery staple")
    Needs: none (pure function).
    """

def aes_cbc_decrypt(data: str, key: str, key_encoding: str = "passphrase", bits: int = 0, iv: str = "", iv_encoding: str = "hex", data_encoding: str = "base64", output: str = "utf-8") -> str:
    """Decrypts data produced by aes_cbc_encrypt (AES-CBC). Accepts the zg1 envelope (with key_encoding=passphrase) or raw iv|ciphertext|tag data in data_encoding (default base64) with a raw key. Raw-key mode is unauthenticated (a wrong key may yield padding errors or garbage); passphrase envelopes carry an HMAC-SHA256 tag and always fail cleanly on a wrong passphrase. A wrong key, a wrong passphrase or damaged data returns an error; the plaintext is returned as UTF-8 text unless output asks for hex or base64.

        data: Ciphertext: a zg1 envelope, or raw iv|ciphertext|tag in data_encoding.
        key: Passphrase (key_encoding=passphrase, the default) or the raw key (AES key must be 16, 24 or 32 bytes.).
        key_encoding: passphrase (default): key is a password stretched by scrypt with a random salt, output is a zg1 envelope. utf-8, hex or base64: key is the raw key bytes in that encoding. One of: passphrase, utf-8, hex, base64.
        bits: AES key size: 128, 192 or 256. 0 (default) means 256 for passphrases and the key's own length for raw keys. Ignored by non-AES ciphers.
        iv: Optional IV/nonce, for deterministic output or interop (random by default). Encoded per iv_encoding. On decrypt with a raw key it means the data has no IV prefix.
        iv_encoding: How iv is turned into bytes: hex (default), base64 or utf-8. One of: hex, base64, utf-8.
        data_encoding: Encoding of data in raw-key mode (default base64). Ignored for zg1 envelopes. One of: base64, base64url, hex, base32.
        output: How the plaintext is returned: utf-8 text (default), hex, HEX, base64, base64url or base32. One of: base64, base64url, hex, HEX, base32, utf-8.

    Example: cipher.aes_cbc_decrypt("zg1.aes256gcm....", "correct horse battery staple")
    Needs: none (pure function).
    """

def aes_ctr_encrypt(text: str, key: str, key_encoding: str = "passphrase", bits: int = 0, iv: str = "", iv_encoding: str = "hex", input_encoding: str = "utf-8", output: str = "base64") -> str:
    """Encrypts text with AES-CTR. Stream mode, 16-byte initial counter block, no padding. Unauthenticated in raw-key mode; never reuse an IV with the same key; passphrase envelopes add an HMAC-SHA256 tag. With key_encoding=passphrase (default) the key is a passphrase stretched by scrypt (N=32768) with a random 16-byte salt and the result is the text envelope zg1.<algo>.<base64url(salt|nonce|ciphertext|tag)>, which aes_ctr_decrypt reads back; with key_encoding utf-8, hex or base64 the key is raw and the result is base64(iv|ciphertext|tag) rendered per output (a random IV unless iv is given).

        text: Plaintext (UTF-8 unless input_encoding says hex or base64).
        key: Passphrase (key_encoding=passphrase, the default) or the raw key (AES key must be 16, 24 or 32 bytes.).
        key_encoding: passphrase (default): key is a password stretched by scrypt with a random salt, output is a zg1 envelope. utf-8, hex or base64: key is the raw key bytes in that encoding. One of: passphrase, utf-8, hex, base64.
        bits: AES key size: 128, 192 or 256. 0 (default) means 256 for passphrases and the key's own length for raw keys. Ignored by non-AES ciphers.
        iv: Optional IV/nonce, for deterministic output or interop (random by default). Encoded per iv_encoding. On decrypt with a raw key it means the data has no IV prefix.
        iv_encoding: How iv is turned into bytes: hex (default), base64 or utf-8. One of: hex, base64, utf-8.
        input_encoding: How text is turned into bytes: utf-8 (default), hex or base64. One of: utf-8, hex, base64.
        output: Raw-key mode only: how iv|ciphertext|tag is rendered (base64 default). Passphrase mode always returns the zg1 envelope. One of: base64, base64url, hex, HEX, base32, utf-8.

    Example: cipher.aes_ctr_encrypt("secret message", "correct horse battery staple")
    Needs: none (pure function).
    """

def caesar_decrypt(text: str, shift: int = 3) -> str:
    """Undoes a Caesar cipher with a known shift (default 3) by shifting letters backwards. Letters keep their case; digits, spaces and punctuation pass through unchanged.

        text: Text to process.
        shift: The shift that was used to encrypt (default 3).

    Example: cipher.caesar_decrypt("Khoor, Zruog!", 3)
    Needs: none (pure function).
    """

def aes_ctr_decrypt(data: str, key: str, key_encoding: str = "passphrase", bits: int = 0, iv: str = "", iv_encoding: str = "hex", data_encoding: str = "base64", output: str = "utf-8") -> str:
    """Decrypts data produced by aes_ctr_encrypt (AES-CTR). Accepts the zg1 envelope (with key_encoding=passphrase) or raw iv|ciphertext|tag data in data_encoding (default base64) with a raw key. Raw-key mode is unauthenticated (a wrong key may yield padding errors or garbage); passphrase envelopes carry an HMAC-SHA256 tag and always fail cleanly on a wrong passphrase. A wrong key, a wrong passphrase or damaged data returns an error; the plaintext is returned as UTF-8 text unless output asks for hex or base64.

        data: Ciphertext: a zg1 envelope, or raw iv|ciphertext|tag in data_encoding.
        key: Passphrase (key_encoding=passphrase, the default) or the raw key (AES key must be 16, 24 or 32 bytes.).
        key_encoding: passphrase (default): key is a password stretched by scrypt with a random salt, output is a zg1 envelope. utf-8, hex or base64: key is the raw key bytes in that encoding. One of: passphrase, utf-8, hex, base64.
        bits: AES key size: 128, 192 or 256. 0 (default) means 256 for passphrases and the key's own length for raw keys. Ignored by non-AES ciphers.
        iv: Optional IV/nonce, for deterministic output or interop (random by default). Encoded per iv_encoding. On decrypt with a raw key it means the data has no IV prefix.
        iv_encoding: How iv is turned into bytes: hex (default), base64 or utf-8. One of: hex, base64, utf-8.
        data_encoding: Encoding of data in raw-key mode (default base64). Ignored for zg1 envelopes. One of: base64, base64url, hex, base32.
        output: How the plaintext is returned: utf-8 text (default), hex, HEX, base64, base64url or base32. One of: base64, base64url, hex, HEX, base32, utf-8.

    Example: cipher.aes_ctr_decrypt("zg1.aes256gcm....", "correct horse battery staple")
    Needs: none (pure function).
    """

def caesar_brute(text: str) -> List[str]:
    """Tries every Caesar shift from 1 to 25 on the text and returns 25 lines of the form 'N: candidate', where N is the forward shift applied to the text (so if the text was encrypted with shift S, the line with N = 26-S is the plaintext). Letters keep their case; other characters are unchanged.

        text: Text to process.

    Example: cipher.caesar_brute("Khoor")
    Needs: none (pure function).
    """

def aes_cfb_encrypt(text: str, key: str, key_encoding: str = "passphrase", bits: int = 0, iv: str = "", iv_encoding: str = "hex", input_encoding: str = "utf-8", output: str = "base64") -> str:
    """Encrypts text with AES-CFB. CFB-128 stream mode, 16-byte IV, no padding. Unauthenticated in raw-key mode; passphrase envelopes add an HMAC-SHA256 tag. With key_encoding=passphrase (default) the key is a passphrase stretched by scrypt (N=32768) with a random 16-byte salt and the result is the text envelope zg1.<algo>.<base64url(salt|nonce|ciphertext|tag)>, which aes_cfb_decrypt reads back; with key_encoding utf-8, hex or base64 the key is raw and the result is base64(iv|ciphertext|tag) rendered per output (a random IV unless iv is given).

        text: Plaintext (UTF-8 unless input_encoding says hex or base64).
        key: Passphrase (key_encoding=passphrase, the default) or the raw key (AES key must be 16, 24 or 32 bytes.).
        key_encoding: passphrase (default): key is a password stretched by scrypt with a random salt, output is a zg1 envelope. utf-8, hex or base64: key is the raw key bytes in that encoding. One of: passphrase, utf-8, hex, base64.
        bits: AES key size: 128, 192 or 256. 0 (default) means 256 for passphrases and the key's own length for raw keys. Ignored by non-AES ciphers.
        iv: Optional IV/nonce, for deterministic output or interop (random by default). Encoded per iv_encoding. On decrypt with a raw key it means the data has no IV prefix.
        iv_encoding: How iv is turned into bytes: hex (default), base64 or utf-8. One of: hex, base64, utf-8.
        input_encoding: How text is turned into bytes: utf-8 (default), hex or base64. One of: utf-8, hex, base64.
        output: Raw-key mode only: how iv|ciphertext|tag is rendered (base64 default). Passphrase mode always returns the zg1 envelope. One of: base64, base64url, hex, HEX, base32, utf-8.

    Example: cipher.aes_cfb_encrypt("secret message", "correct horse battery staple")
    Needs: none (pure function).
    """

def rot5(text: str) -> str:
    """Rotates every digit by 5 places (0 to 5); letters and other characters are unchanged. It is its own inverse: applying it twice returns the original text.

        text: Text to process.

    Example: cipher.rot5("2024")
    Needs: none (pure function).
    """

def aes_cfb_decrypt(data: str, key: str, key_encoding: str = "passphrase", bits: int = 0, iv: str = "", iv_encoding: str = "hex", data_encoding: str = "base64", output: str = "utf-8") -> str:
    """Decrypts data produced by aes_cfb_encrypt (AES-CFB). Accepts the zg1 envelope (with key_encoding=passphrase) or raw iv|ciphertext|tag data in data_encoding (default base64) with a raw key. Raw-key mode is unauthenticated (a wrong key may yield padding errors or garbage); passphrase envelopes carry an HMAC-SHA256 tag and always fail cleanly on a wrong passphrase. A wrong key, a wrong passphrase or damaged data returns an error; the plaintext is returned as UTF-8 text unless output asks for hex or base64.

        data: Ciphertext: a zg1 envelope, or raw iv|ciphertext|tag in data_encoding.
        key: Passphrase (key_encoding=passphrase, the default) or the raw key (AES key must be 16, 24 or 32 bytes.).
        key_encoding: passphrase (default): key is a password stretched by scrypt with a random salt, output is a zg1 envelope. utf-8, hex or base64: key is the raw key bytes in that encoding. One of: passphrase, utf-8, hex, base64.
        bits: AES key size: 128, 192 or 256. 0 (default) means 256 for passphrases and the key's own length for raw keys. Ignored by non-AES ciphers.
        iv: Optional IV/nonce, for deterministic output or interop (random by default). Encoded per iv_encoding. On decrypt with a raw key it means the data has no IV prefix.
        iv_encoding: How iv is turned into bytes: hex (default), base64 or utf-8. One of: hex, base64, utf-8.
        data_encoding: Encoding of data in raw-key mode (default base64). Ignored for zg1 envelopes. One of: base64, base64url, hex, base32.
        output: How the plaintext is returned: utf-8 text (default), hex, HEX, base64, base64url or base32. One of: base64, base64url, hex, HEX, base32, utf-8.

    Example: cipher.aes_cfb_decrypt("zg1.aes256gcm....", "correct horse battery staple")
    Needs: none (pure function).
    """

def rot18(text: str) -> str:
    """ROT13 on letters combined with ROT5 on digits. It is its own inverse: applying it twice returns the original text.

        text: Text to process.

    Example: cipher.rot18("Hello 2024")
    Needs: none (pure function).
    """

def aes_ofb_encrypt(text: str, key: str, key_encoding: str = "passphrase", bits: int = 0, iv: str = "", iv_encoding: str = "hex", input_encoding: str = "utf-8", output: str = "base64") -> str:
    """Encrypts text with AES-OFB. OFB stream mode, 16-byte IV, no padding. Unauthenticated in raw-key mode; passphrase envelopes add an HMAC-SHA256 tag. With key_encoding=passphrase (default) the key is a passphrase stretched by scrypt (N=32768) with a random 16-byte salt and the result is the text envelope zg1.<algo>.<base64url(salt|nonce|ciphertext|tag)>, which aes_ofb_decrypt reads back; with key_encoding utf-8, hex or base64 the key is raw and the result is base64(iv|ciphertext|tag) rendered per output (a random IV unless iv is given).

        text: Plaintext (UTF-8 unless input_encoding says hex or base64).
        key: Passphrase (key_encoding=passphrase, the default) or the raw key (AES key must be 16, 24 or 32 bytes.).
        key_encoding: passphrase (default): key is a password stretched by scrypt with a random salt, output is a zg1 envelope. utf-8, hex or base64: key is the raw key bytes in that encoding. One of: passphrase, utf-8, hex, base64.
        bits: AES key size: 128, 192 or 256. 0 (default) means 256 for passphrases and the key's own length for raw keys. Ignored by non-AES ciphers.
        iv: Optional IV/nonce, for deterministic output or interop (random by default). Encoded per iv_encoding. On decrypt with a raw key it means the data has no IV prefix.
        iv_encoding: How iv is turned into bytes: hex (default), base64 or utf-8. One of: hex, base64, utf-8.
        input_encoding: How text is turned into bytes: utf-8 (default), hex or base64. One of: utf-8, hex, base64.
        output: Raw-key mode only: how iv|ciphertext|tag is rendered (base64 default). Passphrase mode always returns the zg1 envelope. One of: base64, base64url, hex, HEX, base32, utf-8.

    Example: cipher.aes_ofb_encrypt("secret message", "correct horse battery staple")
    Needs: none (pure function).
    """

def rot47(text: str) -> str:
    """Rotates every printable ASCII character from '!' to '~' by 47 places, so letters, digits and punctuation all change; space and non-ASCII are unchanged. It is its own inverse: applying it twice returns the original text.

        text: Text to process.

    Example: cipher.rot47("Hello")
    Needs: none (pure function).
    """

def aes_ofb_decrypt(data: str, key: str, key_encoding: str = "passphrase", bits: int = 0, iv: str = "", iv_encoding: str = "hex", data_encoding: str = "base64", output: str = "utf-8") -> str:
    """Decrypts data produced by aes_ofb_encrypt (AES-OFB). Accepts the zg1 envelope (with key_encoding=passphrase) or raw iv|ciphertext|tag data in data_encoding (default base64) with a raw key. Raw-key mode is unauthenticated (a wrong key may yield padding errors or garbage); passphrase envelopes carry an HMAC-SHA256 tag and always fail cleanly on a wrong passphrase. A wrong key, a wrong passphrase or damaged data returns an error; the plaintext is returned as UTF-8 text unless output asks for hex or base64.

        data: Ciphertext: a zg1 envelope, or raw iv|ciphertext|tag in data_encoding.
        key: Passphrase (key_encoding=passphrase, the default) or the raw key (AES key must be 16, 24 or 32 bytes.).
        key_encoding: passphrase (default): key is a password stretched by scrypt with a random salt, output is a zg1 envelope. utf-8, hex or base64: key is the raw key bytes in that encoding. One of: passphrase, utf-8, hex, base64.
        bits: AES key size: 128, 192 or 256. 0 (default) means 256 for passphrases and the key's own length for raw keys. Ignored by non-AES ciphers.
        iv: Optional IV/nonce, for deterministic output or interop (random by default). Encoded per iv_encoding. On decrypt with a raw key it means the data has no IV prefix.
        iv_encoding: How iv is turned into bytes: hex (default), base64 or utf-8. One of: hex, base64, utf-8.
        data_encoding: Encoding of data in raw-key mode (default base64). Ignored for zg1 envelopes. One of: base64, base64url, hex, base32.
        output: How the plaintext is returned: utf-8 text (default), hex, HEX, base64, base64url or base32. One of: base64, base64url, hex, HEX, base32, utf-8.

    Example: cipher.aes_ofb_decrypt("zg1.aes256gcm....", "correct horse battery staple")
    Needs: none (pure function).
    """

def atbash(text: str) -> str:
    """Mirrors the alphabet (A to Z, B to Y, ...). Letters keep their case; digits, spaces and punctuation pass through unchanged. It is its own inverse: applying it twice returns the original text.

        text: Text to process.

    Example: cipher.atbash("Hello")
    Needs: none (pure function).
    """

def aes_ecb_encrypt(text: str, key: str, key_encoding: str = "passphrase", bits: int = 0, iv: str = "", iv_encoding: str = "hex", input_encoding: str = "utf-8", output: str = "base64") -> str:
    """Encrypts text with AES-ECB (insecure). INSECURE: identical plaintext blocks give identical ciphertext blocks and there is no IV. Supported only for compatibility and puzzles; PKCS#7 padding. Unauthenticated in raw-key mode. With key_encoding=passphrase (default) the key is a passphrase stretched by scrypt (N=32768) with a random 16-byte salt and the result is the text envelope zg1.<algo>.<base64url(salt|nonce|ciphertext|tag)>, which aes_ecb_decrypt reads back; with key_encoding utf-8, hex or base64 the key is raw and the result is base64(iv|ciphertext|tag) rendered per output (a random IV unless iv is given).

        text: Plaintext (UTF-8 unless input_encoding says hex or base64).
        key: Passphrase (key_encoding=passphrase, the default) or the raw key (AES key must be 16, 24 or 32 bytes.).
        key_encoding: passphrase (default): key is a password stretched by scrypt with a random salt, output is a zg1 envelope. utf-8, hex or base64: key is the raw key bytes in that encoding. One of: passphrase, utf-8, hex, base64.
        bits: AES key size: 128, 192 or 256. 0 (default) means 256 for passphrases and the key's own length for raw keys. Ignored by non-AES ciphers.
        iv: Optional IV/nonce, for deterministic output or interop (random by default). Encoded per iv_encoding. On decrypt with a raw key it means the data has no IV prefix.
        iv_encoding: How iv is turned into bytes: hex (default), base64 or utf-8. One of: hex, base64, utf-8.
        input_encoding: How text is turned into bytes: utf-8 (default), hex or base64. One of: utf-8, hex, base64.
        output: Raw-key mode only: how iv|ciphertext|tag is rendered (base64 default). Passphrase mode always returns the zg1 envelope. One of: base64, base64url, hex, HEX, base32, utf-8.

    Example: cipher.aes_ecb_encrypt("secret message", "correct horse battery staple")
    Needs: none (pure function).
    """

def vigenere_encrypt(text: str, key: str) -> str:
    """Encrypts text with the Vigenere cipher: each letter is shifted by the matching letter of the repeating key (A=0). Letters keep their case, other characters pass through and do not advance the key. The key must contain at least one letter; non-letters in the key are ignored.

        text: Text to process.
        key: Key word (letters).

    Example: cipher.vigenere_encrypt("Attack at dawn", "LEMON")
    Needs: none (pure function).
    """

def aes_ecb_decrypt(data: str, key: str, key_encoding: str = "passphrase", bits: int = 0, iv: str = "", iv_encoding: str = "hex", data_encoding: str = "base64", output: str = "utf-8") -> str:
    """Decrypts data produced by aes_ecb_encrypt (AES-ECB (insecure)). Accepts the zg1 envelope (with key_encoding=passphrase) or raw iv|ciphertext|tag data in data_encoding (default base64) with a raw key. Raw-key mode is unauthenticated (a wrong key may yield padding errors or garbage); passphrase envelopes carry an HMAC-SHA256 tag and always fail cleanly on a wrong passphrase. A wrong key, a wrong passphrase or damaged data returns an error; the plaintext is returned as UTF-8 text unless output asks for hex or base64.

        data: Ciphertext: a zg1 envelope, or raw iv|ciphertext|tag in data_encoding.
        key: Passphrase (key_encoding=passphrase, the default) or the raw key (AES key must be 16, 24 or 32 bytes.).
        key_encoding: passphrase (default): key is a password stretched by scrypt with a random salt, output is a zg1 envelope. utf-8, hex or base64: key is the raw key bytes in that encoding. One of: passphrase, utf-8, hex, base64.
        bits: AES key size: 128, 192 or 256. 0 (default) means 256 for passphrases and the key's own length for raw keys. Ignored by non-AES ciphers.
        iv: Optional IV/nonce, for deterministic output or interop (random by default). Encoded per iv_encoding. On decrypt with a raw key it means the data has no IV prefix.
        iv_encoding: How iv is turned into bytes: hex (default), base64 or utf-8. One of: hex, base64, utf-8.
        data_encoding: Encoding of data in raw-key mode (default base64). Ignored for zg1 envelopes. One of: base64, base64url, hex, base32.
        output: How the plaintext is returned: utf-8 text (default), hex, HEX, base64, base64url or base32. One of: base64, base64url, hex, HEX, base32, utf-8.

    Example: cipher.aes_ecb_decrypt("zg1.aes256gcm....", "correct horse battery staple")
    Needs: none (pure function).
    """

def vigenere_decrypt(text: str, key: str) -> str:
    """Decrypts text with the Vigenere cipher: each letter is shifted by the matching letter of the repeating key (A=0). Letters keep their case, other characters pass through and do not advance the key. The key must contain at least one letter; non-letters in the key are ignored.

        text: Text to process.
        key: Key word (letters).

    Example: cipher.vigenere_decrypt("Lxfopv ef rnhr", "LEMON")
    Needs: none (pure function).
    """

def autokey_encrypt(text: str, key: str) -> str:
    """Encrypts text with the Autokey cipher: like Vigenere, but after the key the plaintext itself becomes the key stream. Letters keep their case; other characters pass through. The key must contain at least one letter; non-letters in the key are ignored.

        text: Text to process.
        key: Key word (letters).

    Example: cipher.autokey_encrypt("Attack at dawn", "LEMON")
    Needs: none (pure function).
    """

def chacha20_poly1305_encrypt(text: str, key: str, key_encoding: str = "passphrase", iv: str = "", iv_encoding: str = "hex", aad: str = "", input_encoding: str = "utf-8", output: str = "base64") -> str:
    """Encrypts text with ChaCha20-Poly1305. Authenticated (AEAD, RFC 8439): a wrong key or tampered data fails with an error. 12-byte nonce; never reuse a nonce with the same key. With key_encoding=passphrase (default) the key is a passphrase stretched by scrypt (N=32768) with a random 16-byte salt and the result is the text envelope zg1.<algo>.<base64url(salt|nonce|ciphertext|tag)>, which chacha20_poly1305_decrypt reads back; with key_encoding utf-8, hex or base64 the key is raw and the result is base64(iv|ciphertext|tag) rendered per output (a random IV unless iv is given).

        text: Plaintext (UTF-8 unless input_encoding says hex or base64).
        key: Passphrase (key_encoding=passphrase, the default) or the raw key (ChaCha20-Poly1305 key must be 32 bytes.).
        key_encoding: passphrase (default): key is a password stretched by scrypt with a random salt, output is a zg1 envelope. utf-8, hex or base64: key is the raw key bytes in that encoding. One of: passphrase, utf-8, hex, base64.
        iv: Optional IV/nonce, for deterministic output or interop (random by default). Encoded per iv_encoding. On decrypt with a raw key it means the data has no IV prefix.
        iv_encoding: How iv is turned into bytes: hex (default), base64 or utf-8. One of: hex, base64, utf-8.
        aad: Optional additional authenticated data (UTF-8), used by the AEAD ciphers (GCM, ChaCha20-Poly1305) and ignored by the others; it must be the same on decrypt.
        input_encoding: How text is turned into bytes: utf-8 (default), hex or base64. One of: utf-8, hex, base64.
        output: Raw-key mode only: how iv|ciphertext|tag is rendered (base64 default). Passphrase mode always returns the zg1 envelope. One of: base64, base64url, hex, HEX, base32, utf-8.

    Example: cipher.chacha20_poly1305_encrypt("secret message", "correct horse battery staple")
    Needs: none (pure function).
    """

def autokey_decrypt(text: str, key: str) -> str:
    """Decrypts text with the Autokey cipher: like Vigenere, but after the key the plaintext itself becomes the key stream. Letters keep their case; other characters pass through. The key must contain at least one letter; non-letters in the key are ignored.

        text: Text to process.
        key: Key word (letters).

    Example: cipher.autokey_decrypt("Lxfopv ef rnhr", "LEMON")
    Needs: none (pure function).
    """

def chacha20_poly1305_decrypt(data: str, key: str, key_encoding: str = "passphrase", iv: str = "", iv_encoding: str = "hex", aad: str = "", data_encoding: str = "base64", output: str = "utf-8") -> str:
    """Decrypts data produced by chacha20_poly1305_encrypt (ChaCha20-Poly1305). Accepts the zg1 envelope (with key_encoding=passphrase) or raw iv|ciphertext|tag data in data_encoding (default base64) with a raw key. Authenticated: tampering is always detected. A wrong key, a wrong passphrase or damaged data returns an error; the plaintext is returned as UTF-8 text unless output asks for hex or base64.

        data: Ciphertext: a zg1 envelope, or raw iv|ciphertext|tag in data_encoding.
        key: Passphrase (key_encoding=passphrase, the default) or the raw key (ChaCha20-Poly1305 key must be 32 bytes.).
        key_encoding: passphrase (default): key is a password stretched by scrypt with a random salt, output is a zg1 envelope. utf-8, hex or base64: key is the raw key bytes in that encoding. One of: passphrase, utf-8, hex, base64.
        iv: Optional IV/nonce, for deterministic output or interop (random by default). Encoded per iv_encoding. On decrypt with a raw key it means the data has no IV prefix.
        iv_encoding: How iv is turned into bytes: hex (default), base64 or utf-8. One of: hex, base64, utf-8.
        aad: Optional additional authenticated data (UTF-8), used by the AEAD ciphers (GCM, ChaCha20-Poly1305) and ignored by the others; it must be the same on decrypt.
        data_encoding: Encoding of data in raw-key mode (default base64). Ignored for zg1 envelopes. One of: base64, base64url, hex, base32.
        output: How the plaintext is returned: utf-8 text (default), hex, HEX, base64, base64url or base32. One of: base64, base64url, hex, HEX, base32, utf-8.

    Example: cipher.chacha20_poly1305_decrypt("zg1.aes256gcm....", "correct horse battery staple")
    Needs: none (pure function).
    """

def beaufort(text: str, key: str) -> str:
    """Beaufort cipher: each letter becomes (key letter - letter) mod 26 with a repeating key. It is its own inverse, so the same call decrypts. Letters keep their case; other characters pass through.

        text: Text to process.
        key: Key word (letters).

    Example: cipher.beaufort("Attack at dawn", "LEMON")
    Needs: none (pure function).
    """

def xchacha20_poly1305_encrypt(text: str, key: str, key_encoding: str = "passphrase", iv: str = "", iv_encoding: str = "hex", aad: str = "", input_encoding: str = "utf-8", output: str = "base64") -> str:
    """Encrypts text with XChaCha20-Poly1305. Authenticated (AEAD) with a 24-byte nonce that is safe to pick at random. A wrong key or tampered data fails with an error. With key_encoding=passphrase (default) the key is a passphrase stretched by scrypt (N=32768) with a random 16-byte salt and the result is the text envelope zg1.<algo>.<base64url(salt|nonce|ciphertext|tag)>, which xchacha20_poly1305_decrypt reads back; with key_encoding utf-8, hex or base64 the key is raw and the result is base64(iv|ciphertext|tag) rendered per output (a random IV unless iv is given).

        text: Plaintext (UTF-8 unless input_encoding says hex or base64).
        key: Passphrase (key_encoding=passphrase, the default) or the raw key (XChaCha20-Poly1305 key must be 32 bytes.).
        key_encoding: passphrase (default): key is a password stretched by scrypt with a random salt, output is a zg1 envelope. utf-8, hex or base64: key is the raw key bytes in that encoding. One of: passphrase, utf-8, hex, base64.
        iv: Optional IV/nonce, for deterministic output or interop (random by default). Encoded per iv_encoding. On decrypt with a raw key it means the data has no IV prefix.
        iv_encoding: How iv is turned into bytes: hex (default), base64 or utf-8. One of: hex, base64, utf-8.
        aad: Optional additional authenticated data (UTF-8), used by the AEAD ciphers (GCM, ChaCha20-Poly1305) and ignored by the others; it must be the same on decrypt.
        input_encoding: How text is turned into bytes: utf-8 (default), hex or base64. One of: utf-8, hex, base64.
        output: Raw-key mode only: how iv|ciphertext|tag is rendered (base64 default). Passphrase mode always returns the zg1 envelope. One of: base64, base64url, hex, HEX, base32, utf-8.

    Example: cipher.xchacha20_poly1305_encrypt("secret message", "correct horse battery staple")
    Needs: none (pure function).
    """

def xchacha20_poly1305_decrypt(data: str, key: str, key_encoding: str = "passphrase", iv: str = "", iv_encoding: str = "hex", aad: str = "", data_encoding: str = "base64", output: str = "utf-8") -> str:
    """Decrypts data produced by xchacha20_poly1305_encrypt (XChaCha20-Poly1305). Accepts the zg1 envelope (with key_encoding=passphrase) or raw iv|ciphertext|tag data in data_encoding (default base64) with a raw key. Authenticated: tampering is always detected. A wrong key, a wrong passphrase or damaged data returns an error; the plaintext is returned as UTF-8 text unless output asks for hex or base64.

        data: Ciphertext: a zg1 envelope, or raw iv|ciphertext|tag in data_encoding.
        key: Passphrase (key_encoding=passphrase, the default) or the raw key (XChaCha20-Poly1305 key must be 32 bytes.).
        key_encoding: passphrase (default): key is a password stretched by scrypt with a random salt, output is a zg1 envelope. utf-8, hex or base64: key is the raw key bytes in that encoding. One of: passphrase, utf-8, hex, base64.
        iv: Optional IV/nonce, for deterministic output or interop (random by default). Encoded per iv_encoding. On decrypt with a raw key it means the data has no IV prefix.
        iv_encoding: How iv is turned into bytes: hex (default), base64 or utf-8. One of: hex, base64, utf-8.
        aad: Optional additional authenticated data (UTF-8), used by the AEAD ciphers (GCM, ChaCha20-Poly1305) and ignored by the others; it must be the same on decrypt.
        data_encoding: Encoding of data in raw-key mode (default base64). Ignored for zg1 envelopes. One of: base64, base64url, hex, base32.
        output: How the plaintext is returned: utf-8 text (default), hex, HEX, base64, base64url or base32. One of: base64, base64url, hex, HEX, base32, utf-8.

    Example: cipher.xchacha20_poly1305_decrypt("zg1.aes256gcm....", "correct horse battery staple")
    Needs: none (pure function).
    """

def affine_encrypt(text: str, a: int = 5, b: int = 8) -> str:
    """Affine cipher: maps letter x to (a*x+b) mod 26 (decrypt reverses it). a must be coprime to 26 (1, 3, 5, 7, 9, 11, 15, 17, 19, 21, 23, 25, or equivalents), otherwise it is an error. Letters keep their case; digits, spaces and punctuation pass through unchanged.

        text: Text to process.
        a: Multiplier, coprime to 26 (default 5).
        b: Shift (default 8).

    Example: cipher.affine_encrypt("AFFINE CIPHER", 5, 8)
    Needs: none (pure function).
    """

def chacha20_encrypt(text: str, key: str, key_encoding: str = "passphrase", iv: str = "", iv_encoding: str = "hex", input_encoding: str = "utf-8", output: str = "base64") -> str:
    """Encrypts text with ChaCha20 stream. Raw ChaCha20 keystream XOR (12-byte nonce, counter starts at 0), no authentication in raw-key mode; passphrase envelopes add an HMAC-SHA256 tag. With key_encoding=passphrase (default) the key is a passphrase stretched by scrypt (N=32768) with a random 16-byte salt and the result is the text envelope zg1.<algo>.<base64url(salt|nonce|ciphertext|tag)>, which chacha20_decrypt reads back; with key_encoding utf-8, hex or base64 the key is raw and the result is base64(iv|ciphertext|tag) rendered per output (a random IV unless iv is given).

        text: Plaintext (UTF-8 unless input_encoding says hex or base64).
        key: Passphrase (key_encoding=passphrase, the default) or the raw key (ChaCha20 key must be 32 bytes.).
        key_encoding: passphrase (default): key is a password stretched by scrypt with a random salt, output is a zg1 envelope. utf-8, hex or base64: key is the raw key bytes in that encoding. One of: passphrase, utf-8, hex, base64.
        iv: Optional IV/nonce, for deterministic output or interop (random by default). Encoded per iv_encoding. On decrypt with a raw key it means the data has no IV prefix.
        iv_encoding: How iv is turned into bytes: hex (default), base64 or utf-8. One of: hex, base64, utf-8.
        input_encoding: How text is turned into bytes: utf-8 (default), hex or base64. One of: utf-8, hex, base64.
        output: Raw-key mode only: how iv|ciphertext|tag is rendered (base64 default). Passphrase mode always returns the zg1 envelope. One of: base64, base64url, hex, HEX, base32, utf-8.

    Example: cipher.chacha20_encrypt("secret message", "correct horse battery staple")
    Needs: none (pure function).
    """

def affine_decrypt(text: str, a: int = 5, b: int = 8) -> str:
    """Affine cipher: maps letter x to (a*x+b) mod 26 (decrypt reverses it). a must be coprime to 26 (1, 3, 5, 7, 9, 11, 15, 17, 19, 21, 23, 25, or equivalents), otherwise it is an error. Letters keep their case; digits, spaces and punctuation pass through unchanged.

        text: Text to process.
        a: Multiplier, coprime to 26 (default 5).
        b: Shift (default 8).

    Example: cipher.affine_decrypt("IHHWVC SWFRCP", 5, 8)
    Needs: none (pure function).
    """

def chacha20_decrypt(data: str, key: str, key_encoding: str = "passphrase", iv: str = "", iv_encoding: str = "hex", data_encoding: str = "base64", output: str = "utf-8") -> str:
    """Decrypts data produced by chacha20_encrypt (ChaCha20 stream). Accepts the zg1 envelope (with key_encoding=passphrase) or raw iv|ciphertext|tag data in data_encoding (default base64) with a raw key. Raw-key mode is unauthenticated (a wrong key may yield padding errors or garbage); passphrase envelopes carry an HMAC-SHA256 tag and always fail cleanly on a wrong passphrase. A wrong key, a wrong passphrase or damaged data returns an error; the plaintext is returned as UTF-8 text unless output asks for hex or base64.

        data: Ciphertext: a zg1 envelope, or raw iv|ciphertext|tag in data_encoding.
        key: Passphrase (key_encoding=passphrase, the default) or the raw key (ChaCha20 key must be 32 bytes.).
        key_encoding: passphrase (default): key is a password stretched by scrypt with a random salt, output is a zg1 envelope. utf-8, hex or base64: key is the raw key bytes in that encoding. One of: passphrase, utf-8, hex, base64.
        iv: Optional IV/nonce, for deterministic output or interop (random by default). Encoded per iv_encoding. On decrypt with a raw key it means the data has no IV prefix.
        iv_encoding: How iv is turned into bytes: hex (default), base64 or utf-8. One of: hex, base64, utf-8.
        data_encoding: Encoding of data in raw-key mode (default base64). Ignored for zg1 envelopes. One of: base64, base64url, hex, base32.
        output: How the plaintext is returned: utf-8 text (default), hex, HEX, base64, base64url or base32. One of: base64, base64url, hex, HEX, base32, utf-8.

    Example: cipher.chacha20_decrypt("zg1.aes256gcm....", "correct horse battery staple")
    Needs: none (pure function).
    """

def des3_cbc_encrypt(text: str, key: str, key_encoding: str = "passphrase", iv: str = "", iv_encoding: str = "hex", input_encoding: str = "utf-8", output: str = "base64") -> str:
    """Encrypts text with 3DES-CBC (legacy). Legacy Triple-DES in CBC mode (PKCS#7 padding, 8-byte IV), kept for old systems; 16-byte keys mean K1,K2,K1. Unauthenticated in raw-key mode; passphrase envelopes add an HMAC-SHA256 tag. With key_encoding=passphrase (default) the key is a passphrase stretched by scrypt (N=32768) with a random 16-byte salt and the result is the text envelope zg1.<algo>.<base64url(salt|nonce|ciphertext|tag)>, which des3_cbc_decrypt reads back; with key_encoding utf-8, hex or base64 the key is raw and the result is base64(iv|ciphertext|tag) rendered per output (a random IV unless iv is given).

        text: Plaintext (UTF-8 unless input_encoding says hex or base64).
        key: Passphrase (key_encoding=passphrase, the default) or the raw key (Triple-DES key must be 16 or 24 bytes.).
        key_encoding: passphrase (default): key is a password stretched by scrypt with a random salt, output is a zg1 envelope. utf-8, hex or base64: key is the raw key bytes in that encoding. One of: passphrase, utf-8, hex, base64.
        iv: Optional IV/nonce, for deterministic output or interop (random by default). Encoded per iv_encoding. On decrypt with a raw key it means the data has no IV prefix.
        iv_encoding: How iv is turned into bytes: hex (default), base64 or utf-8. One of: hex, base64, utf-8.
        input_encoding: How text is turned into bytes: utf-8 (default), hex or base64. One of: utf-8, hex, base64.
        output: Raw-key mode only: how iv|ciphertext|tag is rendered (base64 default). Passphrase mode always returns the zg1 envelope. One of: base64, base64url, hex, HEX, base32, utf-8.

    Example: cipher.des3_cbc_encrypt("secret message", "correct horse battery staple")
    Needs: none (pure function).
    """

def substitution_encrypt(text: str, key: str) -> str:
    """Monoalphabetic substitution: the 26-letter key is the cipher alphabet for A..Z (each letter exactly once; anything else is an error). Letters keep their case; digits, spaces and punctuation pass through unchanged.

        text: Text to process.
        key: 26 distinct letters, the cipher alphabet in order A..Z.

    Example: cipher.substitution_encrypt("Hello", "QWERTYUIOPASDFGHJKLZXCVBNM")
    Needs: none (pure function).
    """

def des3_cbc_decrypt(data: str, key: str, key_encoding: str = "passphrase", iv: str = "", iv_encoding: str = "hex", data_encoding: str = "base64", output: str = "utf-8") -> str:
    """Decrypts data produced by des3_cbc_encrypt (3DES-CBC (legacy)). Accepts the zg1 envelope (with key_encoding=passphrase) or raw iv|ciphertext|tag data in data_encoding (default base64) with a raw key. Raw-key mode is unauthenticated (a wrong key may yield padding errors or garbage); passphrase envelopes carry an HMAC-SHA256 tag and always fail cleanly on a wrong passphrase. A wrong key, a wrong passphrase or damaged data returns an error; the plaintext is returned as UTF-8 text unless output asks for hex or base64.

        data: Ciphertext: a zg1 envelope, or raw iv|ciphertext|tag in data_encoding.
        key: Passphrase (key_encoding=passphrase, the default) or the raw key (Triple-DES key must be 16 or 24 bytes.).
        key_encoding: passphrase (default): key is a password stretched by scrypt with a random salt, output is a zg1 envelope. utf-8, hex or base64: key is the raw key bytes in that encoding. One of: passphrase, utf-8, hex, base64.
        iv: Optional IV/nonce, for deterministic output or interop (random by default). Encoded per iv_encoding. On decrypt with a raw key it means the data has no IV prefix.
        iv_encoding: How iv is turned into bytes: hex (default), base64 or utf-8. One of: hex, base64, utf-8.
        data_encoding: Encoding of data in raw-key mode (default base64). Ignored for zg1 envelopes. One of: base64, base64url, hex, base32.
        output: How the plaintext is returned: utf-8 text (default), hex, HEX, base64, base64url or base32. One of: base64, base64url, hex, HEX, base32, utf-8.

    Example: cipher.des3_cbc_decrypt("zg1.aes256gcm....", "correct horse battery staple")
    Needs: none (pure function).
    """

def substitution_decrypt(text: str, key: str) -> str:
    """Monoalphabetic substitution: the 26-letter key is the cipher alphabet for A..Z (each letter exactly once; anything else is an error). Letters keep their case; digits, spaces and punctuation pass through unchanged.

        text: Text to process.
        key: 26 distinct letters, the cipher alphabet in order A..Z.

    Example: cipher.substitution_decrypt("Itssg", "QWERTYUIOPASDFGHJKLZXCVBNM")
    Needs: none (pure function).
    """

def blowfish_cbc_encrypt(text: str, key: str, key_encoding: str = "passphrase", iv: str = "", iv_encoding: str = "hex", input_encoding: str = "utf-8", output: str = "base64") -> str:
    """Encrypts text with Blowfish-CBC (legacy). Legacy Blowfish in CBC mode (PKCS#7 padding, 8-byte IV). Unauthenticated in raw-key mode; passphrase envelopes add an HMAC-SHA256 tag. With key_encoding=passphrase (default) the key is a passphrase stretched by scrypt (N=32768) with a random 16-byte salt and the result is the text envelope zg1.<algo>.<base64url(salt|nonce|ciphertext|tag)>, which blowfish_cbc_decrypt reads back; with key_encoding utf-8, hex or base64 the key is raw and the result is base64(iv|ciphertext|tag) rendered per output (a random IV unless iv is given).

        text: Plaintext (UTF-8 unless input_encoding says hex or base64).
        key: Passphrase (key_encoding=passphrase, the default) or the raw key (Blowfish key must be 4 to 56 bytes.).
        key_encoding: passphrase (default): key is a password stretched by scrypt with a random salt, output is a zg1 envelope. utf-8, hex or base64: key is the raw key bytes in that encoding. One of: passphrase, utf-8, hex, base64.
        iv: Optional IV/nonce, for deterministic output or interop (random by default). Encoded per iv_encoding. On decrypt with a raw key it means the data has no IV prefix.
        iv_encoding: How iv is turned into bytes: hex (default), base64 or utf-8. One of: hex, base64, utf-8.
        input_encoding: How text is turned into bytes: utf-8 (default), hex or base64. One of: utf-8, hex, base64.
        output: Raw-key mode only: how iv|ciphertext|tag is rendered (base64 default). Passphrase mode always returns the zg1 envelope. One of: base64, base64url, hex, HEX, base32, utf-8.

    Example: cipher.blowfish_cbc_encrypt("secret message", "correct horse battery staple")
    Needs: none (pure function).
    """

def rail_fence_encrypt(text: str, rails: int = 3) -> str:
    """Rail fence transposition: writes the text in a zigzag over the given number of rails (2 or more, up to 1000) and reads the rails in order. Every character, including spaces, takes part. Rails at or above the text length leave it unchanged.

        text: Text to process.
        rails: Number of rails, 2 to 1000 (default 3).

    Example: cipher.rail_fence_encrypt("WEAREDISCOVEREDFLEEATONCE", 3)
    Needs: none (pure function).
    """

def blowfish_cbc_decrypt(data: str, key: str, key_encoding: str = "passphrase", iv: str = "", iv_encoding: str = "hex", data_encoding: str = "base64", output: str = "utf-8") -> str:
    """Decrypts data produced by blowfish_cbc_encrypt (Blowfish-CBC (legacy)). Accepts the zg1 envelope (with key_encoding=passphrase) or raw iv|ciphertext|tag data in data_encoding (default base64) with a raw key. Raw-key mode is unauthenticated (a wrong key may yield padding errors or garbage); passphrase envelopes carry an HMAC-SHA256 tag and always fail cleanly on a wrong passphrase. A wrong key, a wrong passphrase or damaged data returns an error; the plaintext is returned as UTF-8 text unless output asks for hex or base64.

        data: Ciphertext: a zg1 envelope, or raw iv|ciphertext|tag in data_encoding.
        key: Passphrase (key_encoding=passphrase, the default) or the raw key (Blowfish key must be 4 to 56 bytes.).
        key_encoding: passphrase (default): key is a password stretched by scrypt with a random salt, output is a zg1 envelope. utf-8, hex or base64: key is the raw key bytes in that encoding. One of: passphrase, utf-8, hex, base64.
        iv: Optional IV/nonce, for deterministic output or interop (random by default). Encoded per iv_encoding. On decrypt with a raw key it means the data has no IV prefix.
        iv_encoding: How iv is turned into bytes: hex (default), base64 or utf-8. One of: hex, base64, utf-8.
        data_encoding: Encoding of data in raw-key mode (default base64). Ignored for zg1 envelopes. One of: base64, base64url, hex, base32.
        output: How the plaintext is returned: utf-8 text (default), hex, HEX, base64, base64url or base32. One of: base64, base64url, hex, HEX, base32, utf-8.

    Example: cipher.blowfish_cbc_decrypt("zg1.aes256gcm....", "correct horse battery staple")
    Needs: none (pure function).
    """

def rail_fence_decrypt(text: str, rails: int = 3) -> str:
    """Rail fence transposition: writes the text in a zigzag over the given number of rails (2 or more, up to 1000) and reads the rails in order. Every character, including spaces, takes part. Rails at or above the text length leave it unchanged.

        text: Text to process.
        rails: Number of rails, 2 to 1000 (default 3).

    Example: cipher.rail_fence_decrypt("WECRLTEERDSOEEFEAOCAIVDEN", 3)
    Needs: none (pure function).
    """

def columnar_encrypt(text: str, key: str) -> str:
    """Columnar transposition: writes the text in rows under the key word and reads the columns in alphabetical order of the key letters (ties left to right). Irregular last row, no padding; every character, including spaces, takes part. The key must not be empty.

        text: Text to process.
        key: Key word; its letters order the columns.

    Example: cipher.columnar_encrypt("WEAREDISCOVERED", "ZEBRAS")
    Needs: none (pure function).
    """

def rc4_encrypt(text: str, key: str, key_encoding: str = "passphrase", iv: str = "", iv_encoding: str = "hex", input_encoding: str = "utf-8", output: str = "base64") -> str:
    """Encrypts text with RC4 (broken). BROKEN stream cipher with no IV, kept only for compatibility. Unauthenticated in raw-key mode; passphrase envelopes add an HMAC-SHA256 tag. With key_encoding=passphrase (default) the key is a passphrase stretched by scrypt (N=32768) with a random 16-byte salt and the result is the text envelope zg1.<algo>.<base64url(salt|nonce|ciphertext|tag)>, which rc4_decrypt reads back; with key_encoding utf-8, hex or base64 the key is raw and the result is base64(iv|ciphertext|tag) rendered per output (a random IV unless iv is given).

        text: Plaintext (UTF-8 unless input_encoding says hex or base64).
        key: Passphrase (key_encoding=passphrase, the default) or the raw key (RC4 key must be 1 to 256 bytes.).
        key_encoding: passphrase (default): key is a password stretched by scrypt with a random salt, output is a zg1 envelope. utf-8, hex or base64: key is the raw key bytes in that encoding. One of: passphrase, utf-8, hex, base64.
        iv: Optional IV/nonce, for deterministic output or interop (random by default). Encoded per iv_encoding. On decrypt with a raw key it means the data has no IV prefix.
        iv_encoding: How iv is turned into bytes: hex (default), base64 or utf-8. One of: hex, base64, utf-8.
        input_encoding: How text is turned into bytes: utf-8 (default), hex or base64. One of: utf-8, hex, base64.
        output: Raw-key mode only: how iv|ciphertext|tag is rendered (base64 default). Passphrase mode always returns the zg1 envelope. One of: base64, base64url, hex, HEX, base32, utf-8.

    Example: cipher.rc4_encrypt("secret message", "correct horse battery staple")
    Needs: none (pure function).
    """

def columnar_decrypt(text: str, key: str) -> str:
    """Columnar transposition: writes the text in rows under the key word and reads the columns in alphabetical order of the key letters (ties left to right). Irregular last row, no padding; every character, including spaces, takes part. The key must not be empty.

        text: Text to process.
        key: Key word; its letters order the columns.

    Example: cipher.columnar_decrypt("EEWDRAIDOSRCEVE", "ZEBRAS")
    Needs: none (pure function).
    """

def rc4_decrypt(data: str, key: str, key_encoding: str = "passphrase", iv: str = "", iv_encoding: str = "hex", data_encoding: str = "base64", output: str = "utf-8") -> str:
    """Decrypts data produced by rc4_encrypt (RC4 (broken)). Accepts the zg1 envelope (with key_encoding=passphrase) or raw iv|ciphertext|tag data in data_encoding (default base64) with a raw key. Raw-key mode is unauthenticated (a wrong key may yield padding errors or garbage); passphrase envelopes carry an HMAC-SHA256 tag and always fail cleanly on a wrong passphrase. A wrong key, a wrong passphrase or damaged data returns an error; the plaintext is returned as UTF-8 text unless output asks for hex or base64.

        data: Ciphertext: a zg1 envelope, or raw iv|ciphertext|tag in data_encoding.
        key: Passphrase (key_encoding=passphrase, the default) or the raw key (RC4 key must be 1 to 256 bytes.).
        key_encoding: passphrase (default): key is a password stretched by scrypt with a random salt, output is a zg1 envelope. utf-8, hex or base64: key is the raw key bytes in that encoding. One of: passphrase, utf-8, hex, base64.
        iv: Optional IV/nonce, for deterministic output or interop (random by default). Encoded per iv_encoding. On decrypt with a raw key it means the data has no IV prefix.
        iv_encoding: How iv is turned into bytes: hex (default), base64 or utf-8. One of: hex, base64, utf-8.
        data_encoding: Encoding of data in raw-key mode (default base64). Ignored for zg1 envelopes. One of: base64, base64url, hex, base32.
        output: How the plaintext is returned: utf-8 text (default), hex, HEX, base64, base64url or base32. One of: base64, base64url, hex, HEX, base32, utf-8.

    Example: cipher.rc4_decrypt("zg1.aes256gcm....", "correct horse battery staple")
    Needs: none (pure function).
    """

def scytale_encrypt(text: str, diameter: int = 4) -> str:
    """Scytale transposition: text wrapped around a rod of the given diameter (letters visible per turn); reading off the strip takes every diameter-th character in turn. Every character takes part; no padding.

        text: Text to process.
        diameter: Rod diameter, 1 to 1000 (default 4).

    Example: cipher.scytale_encrypt("HELPMEIAMTRAPPED", 4)
    Needs: none (pure function).
    """

def scytale_decrypt(text: str, diameter: int = 4) -> str:
    """Scytale transposition: text wrapped around a rod of the given diameter (letters visible per turn); reading off the strip takes every diameter-th character in turn. Every character takes part; no padding.

        text: Text to process.
        diameter: Rod diameter, 1 to 1000 (default 4).

    Example: cipher.scytale_decrypt("HMAEEPLITPAPMRED", 4)
    Needs: none (pure function).
    """

def playfair_encrypt(text: str, key: str) -> str:
    """Playfair digraph cipher on a 5x5 key square (J merged into I). Only letters are used: output is uppercase without spaces or punctuation; encryption inserts X between doubled letters and pads an odd end, and decryption leaves those padding letters in place. Decrypt needs an even letter count.

        text: Text to process.
        key: Key word or phrase for the square.

    Example: cipher.playfair_encrypt("Hide the gold in the tree stump", "Playfair example")
    Needs: none (pure function).
    """

def playfair_decrypt(text: str, key: str) -> str:
    """Playfair digraph cipher on a 5x5 key square (J merged into I). Only letters are used: output is uppercase without spaces or punctuation; encryption inserts X between doubled letters and pads an odd end, and decryption leaves those padding letters in place. Decrypt needs an even letter count.

        text: Text to process.
        key: Key word or phrase for the square.

    Example: cipher.playfair_decrypt("BMODZBXDNABEKUDMUIXMMOUVIF", "Playfair example")
    Needs: none (pure function).
    """

def polybius_encrypt(text: str, key: str = "") -> str:
    """Polybius square on a 5x5 grid (J merged into I): each letter becomes row and column digits 1-5, groups separated by spaces. Only letters are encoded (other characters are dropped); decryption reads digits 1-5 in pairs and returns uppercase letters. An optional key word scrambles the square.

        text: Text to process.
        key: Optional key word for the square (empty = plain alphabet).

    Example: cipher.polybius_encrypt("Hello")
    Needs: none (pure function).
    """

def polybius_decrypt(text: str, key: str = "") -> str:
    """Polybius square on a 5x5 grid (J merged into I): each letter becomes row and column digits 1-5, groups separated by spaces. Only letters are encoded (other characters are dropped); decryption reads digits 1-5 in pairs and returns uppercase letters. An optional key word scrambles the square.

        text: Text to process.
        key: Optional key word for the square (empty = plain alphabet).

    Example: cipher.polybius_decrypt("23 15 31 31 34")
    Needs: none (pure function).
    """

def bacon_encrypt(text: str, variant: str = "distinct") -> str:
    """Bacon's cipher: each letter becomes five A/B symbols, groups separated by spaces (variant 'distinct' uses 26 codes, 'classic' merges I/J and U/V into 24). Only letters are encoded; decryption reads A/B (or 0/1) in groups of five and returns uppercase letters.

        text: Text to process.
        variant: distinct (26 letters, default) or classic (24 letters). One of: distinct, classic.

    Example: cipher.bacon_encrypt("Hello")
    Needs: none (pure function).
    """

def bacon_decrypt(text: str, variant: str = "distinct") -> str:
    """Bacon's cipher: each letter becomes five A/B symbols, groups separated by spaces (variant 'distinct' uses 26 codes, 'classic' merges I/J and U/V into 24). Only letters are encoded; decryption reads A/B (or 0/1) in groups of five and returns uppercase letters.

        text: Text to process.
        variant: distinct (26 letters, default) or classic (24 letters). One of: distinct, classic.

    Example: cipher.bacon_decrypt("AABBB AABAA ABABB ABABB ABBAB")
    Needs: none (pure function).
    """

def rsa_encrypt(text: str, public_pem: str, input_encoding: str = "utf-8", hash: str = "sha256", output: str = "base64") -> str:
    """Encrypts text with an RSA public key (PEM) using RSA-OAEP with SHA-256 (or hash). The message must be short (key bytes minus 2*hash size minus 2, i.e. 190 bytes for RSA-2048/SHA-256); longer input is an error. Output is base64 by default.

        text: Message (UTF-8 unless input_encoding says hex or base64).
        public_pem: RSA public key in PEM (PKIX, PKCS#1 or a certificate).
        input_encoding: How text is turned into bytes: utf-8 (default), hex or base64. One of: utf-8, hex, base64.
        hash: Hash function (default sha256). One of: sha256, sha384, sha512, sha1.
        output: Output format (default base64). One of: base64, base64url, hex, HEX, base32, utf-8.

    Example: cipher.rsa_encrypt("hi", pem)
    Needs: none (pure function).
    """

def rsa_decrypt(data: str, private_pem: str, data_encoding: str = "base64", hash: str = "sha256", output: str = "utf-8") -> str:
    """Decrypts RSA-OAEP data with an RSA private key (PEM, PKCS#8 or PKCS#1). hash must match the one used to encrypt (default sha256). A wrong key or damaged data is an error.

        data: Ciphertext in data_encoding (base64 default).
        private_pem: RSA private key in PEM.
        data_encoding: Encoding of data (default base64). One of: base64, hex.
        hash: Hash function (default sha256). One of: sha256, sha384, sha512, sha1.
        output: How the plaintext is returned (utf-8 default). One of: base64, base64url, hex, HEX, base32, utf-8.

    Example: cipher.rsa_decrypt(ciphertext_b64, private_pem)
    Needs: none (pure function).
    """

def pem_decode(pem: str) -> Dict[str, str]:
    """Parses the first PEM block of the text and returns a dict with type (e.g. 'PRIVATE KEY', 'CERTIFICATE'), bytes_hex (the DER body in hex), length (bytes), count (number of PEM blocks found) and any headers as header_<name>. Text with no PEM block is an error.

        pem: PEM text including the BEGIN/END lines.

    Example: cipher.pem_decode(pem_text)
    Needs: none (pure function).
    """
