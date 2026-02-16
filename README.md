# Vel

The project CLI library for [Velocity](https://github.com/velocitykode/velocity) Go web framework.

## What is Vel?

Vel provides development commands for Velocity projects. Unlike traditional CLIs, vel is **built from source** in each project, giving it access to your migrations, models, and bootstrap code.

## Architecture

Velocity uses two CLI tools:

| Tool | Install | Purpose |
|------|---------|---------|
| [`velocity`](https://github.com/velocitykode/velocity-installer) | Homebrew | Create projects, manage config |
| `vel` | Built from source | Dev server, migrations, generators |

## Usage

After creating a project with `velocity new`, use `./vel` for development:

```bash
./vel serve           # Start dev server with hot reload
./vel migrate         # Run database migrations
./vel migrate:fresh   # Drop and re-run all migrations
./vel make:controller # Generate a controller
./vel build           # Build for production
./vel key:generate    # Generate encryption key
```

### Shell Function (Optional)

Add to `~/.zshrc` to use `vel` instead of `./vel`:

```bash
vel() { [ -x ./vel ] && ./vel "$@" || echo "vel: not found"; }
```

## How It Works

Your project's `cmd/vel/main.go` imports this package:

```go
package main

import (
    "github.com/velocitykode/velocity-cli"
    _ "myapp/bootstrap"
    _ "myapp/database/migrations"
)

func main() {
    vel.Execute()
}
```

This gives vel access to your project's migrations and configuration.

## Documentation

Full documentation at **[velocitykode.com/docs](https://velocitykode.com/docs)**

## License

MIT License - See [LICENSE](LICENSE) for details.
