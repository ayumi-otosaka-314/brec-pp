_default:
    @just --list --unsorted

# Build main executable for brec-pp
build:
    go build -o ./bin/

# Build tool specified by {{name}}
build-tool name:
    go build -o ./bin/ ./cmd/{{name}}
