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
