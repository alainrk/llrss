# LLRSS (Long Live RSS)

**[Under Development]**

[![Build Status](https://github.com/alainrk/llrss/actions/workflows/main.yml/badge.svg)](https://github.com/alainrk/llrss/actions/workflows/main.yml)

Keep RSS alive.

## Overview

LLRSS is an RSS feed reader that helps you stay on top of your favorite content sources without the noise of social media algorithms.

## Prerequisites

- Go 1.21 or later
- Make (for build automation)
- Git

## Installation

1. Clone the repository:

```bash
git clone https://github.com/alainrk/llrss.git
cd llrss
```

2. Install dependencies:

```bash
go mod download
```

3. Build the application:

```bash
make build
```

## Quick Start

1. Run the application:

```bash
make run
```

2. For development with hot reload:

```bash
make run/live
```

The server will start at `http://localhost:8080` (or your configured port).

## Development

### Testing

Run all tests:

```bash
make test
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
