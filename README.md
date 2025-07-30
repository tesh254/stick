# stick

`stick` is a command-line tool for working with Git repositories. It provides useful information about commits, branches, and versions.

## Outline

- [Getting Started](#getting-started)
  - [Prerequisites](#prerequisites)
  - [Installation](#installation)
- [Usage](#usage)
- [Contributing](#contributing)
- [License](#license)

## Getting Started

### Prerequisites

- Go 1.24 or higher

### Installation

1.  Clone the repository:
    ```sh
    git clone https://github.com/tesh254/stick.git
    ```
2.  Navigate to the project directory:
    ```sh
    cd stick
    ```
3.  Build the application:
    ```sh
    go build .
    ```

## Usage

```sh
./stick [command]
```

### Available Commands

- `show -c <commit-hash>`: Show the diff for a specific commit.
- `version`: Print the version information.

## Contributing

Contributions are welcome! Please read the [contributing guidelines](CONTRIBUTING.md) before getting started.

## License

This project is licensed under the MIT License.