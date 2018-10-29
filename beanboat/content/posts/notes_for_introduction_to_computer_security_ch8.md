+++
title = "Notes for Introduction to Computer Security CH8"
description = "Notes for <Introduction to Computer Security> CH8"
tags = [
    "cryptography",
    "computer_security",
]
date = "2018-09-26"
categories = [
    "notes_for_introduction_to_computer_security",
    "book_notes",
]
menu = "main"
+++

# CH8 Cryptography

<!--more-->
## Symmetric Encryption
### Concept and Properties
* same key used for encryption and decryption
* Advantage: conceptual simplicity
* Disadvantage:
 * secure channel to set up key
 * Distribution:
   * A distinct keys needs to be set up for each pair of communicating users
   * Quadratic number of keys (n(n-1)/2) for pairwise communication

### Classic Symmetric Encryption
#### Alphabet Shift Cipher / Julius Caesar's Cipher
#### Substitution Cipher / Arbitrary permutation of the characters
can be cracked by frequency analysis
#### One-Time Pad
* Key: sequence of random bits, same length as plaintext
* Encryption: C = K ⊕ P
* Decryption: P = K ⊕ C
* Advantages:
  * Each bit of the ciphertext is random
  * Fully secure if key used only once
* Disadvantages:
  * Key as large as plaintext
  * Difficult to generate and share
  * Key cannot be reused

### Modern Symmetric Encryption
* Data Encryption Standard (DES)
 * Developed by IBM in collaboration with the NSA
 * Became US government standard in 1977
 * 56-bit keys
 * Exhaustive search attack feasible since late 90s
* Advanced Encryption Standard (AES)
 * Selected as US government standard in 2001 through open competition
 * 128-, 192-, or 256-bit keys
 * Exhaustive search attack not currently possible

#### Stream Cipher / Pseudo-Random Number Generators
 * Key stream
   * Pseudo-random bit sequence generated from a secret key K: SK = SK[0], SK[1], SK[2], …
   * Generated on-demand, one bit (or block) at the time
 * Stream cipher: XOR the plaintext with the key stream C[i] = SK[i] xor P[i]
 * Advantages
   * Fixed-length secret key
   * Plaintext can have arbitrary length (e.g., media stream)
   * Incremental encryption and decryption
   * Works for packets sent over an unreliable channel
 * Disadvantages: Key stream cannot be reused

#### Block Cipher
* A block cipher is a symmetric encryption scheme for messages (blocks) of a given fixed length. The length of the block is independent from the length of the key
* AES is a block cipher that operates on blocks of 128 bits (16 bytes)
 * AES supports keys of length 128, 192, and 256 bits
* Operation of Modes
    * Electronic Codebook (ECB) Mode
        * Partition plaintext P into sequence of m blocks P[0], …, P[m-1], where n <= b m < n + b
        * Block P[i] encrypted into ciphertext block C[i] = EK(P[i])
        * Documents and images are not suitable for ECB as patterns in the plaintext are repeated in the ciphertext
        * Simplicity, tolerance of loss of a block
    * Cipher-Block Chaining (CBC) Mode
        * Encryption: C[i] = EK(C[i -1] ⊕ P[i]), C[-1] = V is a random block (initialization vector) sent encrypted during setup
        * Decryption: P[i] = Dk(C[i]) xor C[i-1]
        * Encryption is sequential while decryption can be parallelized
    * Cipher Feedback (CFB) Mode
        * Encryption: C[i] = EK(C[i−1]) ⊕ P[i]; Decryption: P[i] = EK(C[i]) ⊕ C[i−1]
        * The decryption algorithm is actually never used and thus can be fast
    * Output Feedback (OFB) Mode
    * Counter (CTR) Mode
* Padding
 * Block cipher modes require the length n of the plaintext to be a multiple of the block size b
* Initialization Vector
 * Goal: Avoid sharing a new secret key for each stream encryption
 *  Solution: Use a two-part key (U, V). Part U is fixed, Part V is transmitted together
with the ciphertext. V is called initialization vector
 * Setup
   * Alice and Bob share secret U Encryption
   * Alice picks V and creates key K = (U, V)
   * Alice creates stream ciphertext C and sends (V, C)
   * Bob reconstructs key K = (U, V)
   * Bob decrypts the message


## Cryptographic Hash Functions
* A cryptographic hash function produces a compressed digest of a message, while also provides a mapping that is deterministic, one-way, and collision- resistant.

#### Properties
* One-way: Given a hash value x, it is hard to find a plaintext P such that h(P) = x
* Weak collision resistance: Given a plaintext P, it is hard to find a plaintext Q such that h(Q) = h(P)
* Strong collision resistance: It is hard to find a pair of plaintexts P and Q such that h(Q) = h(P)

#### Practical Hash Functions for Cryptographic Applications
* SHA-256, SHA-512 (Secure Hash Algorithm) <- Merkle-Damga ̊rd construction

#### Applications
* Password authentication
* Coin tossing online
* Alice and Bob wants to toss a
coin. They communicate via text messages.Neither trusts the other
* Solution: Alice and Bob agree on a crypto hash function, h. Alice picks a random value R and sends d = h(R) to Bob. Bob guesses whether R is odd or even. Alice reveals R to Bob

#### Birthday Attach
* Probability of repetitions in a sequence of random values. Assume each value is an integer between 0 and n − 1, about 50% probability for a sequence of length n
* Brute force attack on strong collision resistance needs 2^(b/2) tries to succeed with 50% probability
* A b-bit hash function has b/2	bits of security

## Public Key Cryptography
#### Concept
* Key pair
 * Public key: shared with everyone
 * Secret key: kept secret, hard to derive from the public key
* Protocol
 * Sender encrypts using recipient's public key
 * Recipient decrypts using its secret key

#### Properties
* Advantages
 * A single public-secret key pair allows receiving confidential messages from multiple parties
* Disadvantages
 * Conceptually complex
 * Slower performance than symmetric cryptography
 * Long key length

#### RSA
* Most widely used public key cryptosystem today
* RSA patent expired in 2000
* 2048-bit (or longer) keys recommended
* Much slower than AES
* Typically used to encrypt an AES symmetric key

## Digital Signature
#### Goals
* Authenticity
 * Binds an identity (signer) to a message
 * Provides assurance of the signer
* Framework: Alice should be able to use her private key with a signature algorithm to produce a digital signature, SAlice(M), for a message, M. In addition, given Alice’s public key, the message M, and Alice’s signature, SAlice(M), it should be possible for another party, Bob, to verify Alice’s signature on M, using just these elements.

#### Properties
* Unforgeability
 * An attacker cannot forge a signature for a different identity
* Nonrepudiation
 * Signer cannot deny having signed the message
* Integrity / Nonmutability
 *  An attacker cannot take a signature by Alice for a message and create a signature by Alice for a different message

#### Public-Key Encryption Method
* Alice “decrypts” plaintext message M with the secret key and obtains a
digital signature on M sign(M, S<sub>K</sub>) = D<sub>S<sub>K</sub></sub>(M)
* Bob “encrypts” signature S with Alice's public key P<sub>K</sub>, and checks if the result is message M: M == E<sub>P<sub>K</sub></sub>(S)

* Signing hash: shorten the message and fasten the encryption and decryption
 * Sign: S = D<sub>S<sub>K</sub></sub> (h(M))
 * Verify: h(M) == E<sub>P<sub>K</sub></sub> (S)
