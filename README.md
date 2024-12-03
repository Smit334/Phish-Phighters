# Phish-Phighters

Phish-Phighters is a secure file storage and sharing system implemented in Go. It provides functionalities for storing, loading, appending, and sharing files securely among users.

## Features

- Secure file storage with encryption and integrity checks
- File sharing with access control
- Append operations to existing files
- User authentication and invitation system

## Setup

### Prerequisites

- Go 1.20 or later
- Git

### Installation

1. Clone the repository:
    ```sh
    git clone https://github.com/yourusername/Phish-Phighters.git
    cd Phish-Phighters
    ```

2. Install dependencies:
    ```sh
    go mod tidy
    ```

## Usage

### Running Tests

To run the tests, use the following command:
```sh
go test ./client_test
```

### File Operations

Phish-Phighters supports various file operations such as storing, loading, and appending to files. Below are some examples of how to use these functionalities:

#### Storing a File
```go
alice, err := client.InitUser("alice", "password")
if err != nil {
    log.Fatal(err)
}
err = alice.StoreFile("example.txt", []byte("Hello, World!"))
if err != nil {
    log.Fatal(err)
}
```

#### Loading a File
```go
data, err := alice.LoadFile("example.txt")
if err != nil {
    log.Fatal(err)
}
fmt.Println(string(data)) // Output: Hello, World!
```

#### Appending to a File
```go
err = alice.AppendToFile("example.txt", []byte(" More content."))
if err != nil {
    log.Fatal(err)
}
data, err = alice.LoadFile("example.txt")
if err != nil {
    log.Fatal(err)
}
fmt.Println(string(data)) // Output: Hello, World! More content.
```

### Sharing Files

Phish-Phighters allows users to share files with others securely. Below is an example of how to share a file and accept an invitation:

#### Sharing a File
```go
invite, err := alice.CreateInvitation("example.txt", "bob")
if err != nil {
    log.Fatal(err)
}
```

#### Accepting an Invitation
```go
bob, err := client.InitUser("bob", "password")
if err != nil {
    log.Fatal(err)
}
err = bob.AcceptInvitation("alice", invite, "bob_example.txt")
if err != nil {
    log.Fatal(err)
}
data, err := bob.LoadFile("bob_example.txt")
if err != nil {
    log.Fatal(err)
}
fmt.Println(string(data)) // Output: Hello, World! More content.
```

### Revoking Access

Users can also revoke access to shared files:

```go
err = alice.RevokeAccess("example.txt", "bob")
if err != nil {
    log.Fatal(err)
}
```

This project is a secure file storage and sharing system with user authentication and multiple device support.

## Data Structures
Our system organizes data in Google Datastore using structs for:
- **Users**: Contains user-specific information.
- **Files**: Stored as linked lists for efficient operations.
- **Intermediates**: Facilitates secure file sharing.
- **ShareHubs**: Manages shared file access.

### Key Features:
- **Confidentiality and Integrity**: Data is encrypted and MACed before storage.
- **Linked List Storage**: Files are split into blocks, stored as nodes in a linked list, with each node pointing to its ciphertext and the next node.

---

## Helper Functions
1. **EncryptThenSign**: Encrypts data and signs it.
2. **VerifyThenDecrypt**: Verifies a signature and decrypts data.
3. **GenerateUserUUID**: Creates a UUID for a user.
4. **GenerateUserKeys**: Generates encryption and verification keys.
5. **GenerateBodyKeys**: Creates keys for file body and node encryption.
6. **MacThenDecrypt**: Verifies HMAC and decrypts.
7. **EncryptThenMac**: Encrypts and applies HMAC.

---

## User Authentication
- **New Users**: 
  - UUID is generated deterministically using a hash of the username.
  - Keys are derived using `Argon2Key` with `username ⊕ constants`.
- **Existing Users**:
  - Authentication ensures decryption keys match the stored user struct.
  - Error handling occurs if the username or credentials are invalid.

---

## Multiple Devices
- All devices pull user structs from Datastore and sync changes.
- Frequent updates ensure no desynchronization between devices.

---

## File Storage and Retrieval
### File Upload:
1. Split plaintext into `n` blocks.
2. Encrypt each block using:
   - Keys: Derived using `HashKDF(α, i)`.
   - IVs: Randomly generated.
   - MAC: Calculated using `Hash(α ⊕ i)`.
3. Store blocks as nodes in a linked list, with pointers for navigation.
4. The head node (Filehead) contains:
   - List length (`n`).
   - Keys (`α` and `β`).
   - UUID pointers for efficient access.

### File Retrieval:
1. Start with the Intermediate structure:
   - Decrypt with RSA to access the Filehead and decryption keys.
2. Traverse the linked list:
   - Decrypt nodes using derived keys (`β`).
   - Recompute keys dynamically based on node index.

---

## Efficient Append
Appending to files is efficient due to constant-time access:
1. The Filehead points to the first and last nodes.
2. Only the first and last blocks are modified during append operations.

---

## File Sharing
### CreateInvitation:
1. **Ownership Verification**: Only file owners can create invitations.
2. **ShareHub Creation**:
   - Includes UUID of Filehead, keys (`α`, `β`), and recipient UUID.
   - Encrypted with the recipient's public key and signed by the sender.
3. Updates:
   - Owner’s user struct (`SharedByMe` list).
   - Existing invitations for non-owners.

### AcceptInvitation:
1. Verifies the invitation signature using the sender's public key.
2. Decrypts the invitation using the recipient's private key.
3. Creates an Intermediate struct with:
   - File pointers and keys.
   - Ownership flag (`IsOwner`).
4. Updates the recipient's user struct (`SharedToMe` list).

---

## File Revocation
1. **Re-encryption**: Alters file encryption to invalidate existing keys.
2. **Access Revocation**:
   - Removes revoked user’s ShareHub.
   - Updates `SharedByMe` and `ShareList`.
3. **Key Update**:
   - Generates new keys (`α`, `β`).
   - Updates ShareHub structures for legitimate users.

---

## Key Value Mapping
| **UUID**                                 | **Encrypted** | **Key Derivation**                  | **Stored Data**                                                                                   | **Description**                                                                                   |
|------------------------------------------|---------------|--------------------------------------|---------------------------------------------------------------------------------------------------|---------------------------------------------------------------------------------------------------|
| `UUID.frombytes(sha256(username)[:16])`  | Yes           | User login credentials              | RSA keys, file share info, ownership details.                                                    | Central user struct for file management.                                                         |
| `UUID.New()`                             | Yes           | Random key generation               | File UUID, α, β, shared details.                                                                 | ShareHub struct for managing shared file access.                                                 |
| `UUID.New()`                             | Yes           | Derived in File Storage process     | Linked list pointers, node metadata (e.g., length, UUIDs).                                       | File struct representing blocks as linked list nodes.                                             |
| `UUID.New()`                             | Yes           | Derived in File Storage process     | Encrypted ciphertext.                                                                            | Stores file content as encrypted data blocks.                                                    |

---

## Summary
This system provides secure and efficient mechanisms for:
- User authentication.
- File storage and retrieval.
- File sharing and access revocation.

Each component is designed to ensure data integrity, confidentiality, and usability across devices and users.

## Contributing

Contributions are welcome! Please fork the repository and submit a pull request.

## License

This project is licensed under the MIT License.

## Contact

For any questions or support, please contact:

- Name: Smit Malde
- GitHub: [Smit334](https://github.com/Smit334)

or


- Name: Andres Chaidez
- GitHub: [4ndyNMC](https://github.com/4ndyNMC)
Phish Phighters - Design Documentation