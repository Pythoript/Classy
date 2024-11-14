# Classy

Classy is a powerful command-line tool written in Go that recursively scans directories for HTML, CSS, JavaScript, and PHP files, extracts class names, and renames them using a unique and efficient naming scheme. This project is designed to help developers manage class names in their web projects, ensuring consistency and reducing the size of class names.

## Table of Contents

- [Features](#features)
- [Installation](#installation)
- [Usage](#usage)
- [Options](#options)
- [Example](#example)
- [How to Contribute](#how-to-contribute)
- [License](#license)

## Features

- **Recursive File Scanning**: Scans all subdirectories for supported file types.
- **Class Extraction**: Identifies and counts class names used in HTML, CSS, JavaScript, and PHP files.
- **Unique Class Renaming**: Converts long class names into a shorter, more manageable format, starting with 'a', 'b', ..., 'z', and then continuing with 'aa', 'ab', ..., 'az', 'a0', ..., 'a9', 'aaa', etc.
- **HTML Class Handling**: Renames classes in standard HTML attributes and non-standard class declarations.
- **CSS Selector Support**: Processes complex CSS selectors including pseudo-classes and pseudo-elements, such as `:hover`, `:not()`, `::before`, and `::after`.
- **JavaScript Class Manipulation**: Supports various JavaScript class manipulations, including `classList.add()`, `classList.remove()`, and `querySelector()`.
- **Duplicate Class Handling**: Optionally allows duplicate classes in HTML attributes.
- **Preview Mode**: Preview class renaming without modifying the files.

## Installation

To get started with Classy, ensure you have [Go](https://golang.org/dl/) installed on your machine. Then, follow these steps:

1. Clone the repository:

   ```bash
   git clone https://github.com/Pythoript/Classy.git
   cd Classy
   ```

2. Install dependencies:

   ```bash
   go mod tidy
   ```

3. Build the application:

   ```bash
   go build -o classy
   ```

## Usage

Run the application from the command line, providing the directory to scan:

```bash
./classy -dir <directory_path>
```

## Options

- `-dir`: The directory to recursively scan for HTML, CSS, JS, and PHP files. Defaults to the current directory.
- `-preview`: Show class renaming without modifying files. Useful for verifying changes before applying them.
- `-allow-duplicates`: Allow duplicate classes in HTML attributes. By default, duplicates are combined.

## Example

To rename classes in the current directory while combining duplicates:

```bash
./classy -dir .
```

To preview class renaming without modifying files:

```bash
./classy -dir . -preview
```

To allow duplicate classes in HTML attributes:

```bash
./classy -dir . -allow-duplicates
```

## How to Contribute

We welcome contributions to Classy! If you'd like to contribute, please follow these steps:

1. Fork the repository.
2. Create a new branch (`git checkout -b feature/YourFeature`).
3. Make your changes and commit them (`git commit -m 'Add some feature'`).
4. Push to the branch (`git push origin feature/YourFeature`).
5. Open a pull request explaining your changes.

Please ensure that your code adheres to the project's standards.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.
