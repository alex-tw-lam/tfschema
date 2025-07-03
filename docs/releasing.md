# Releasing

This document outlines the process for creating a new release of `tfschema`.

## Versioning

This project follows [Semantic Versioning](https://semver.org/). All release tags should be in the format `vX.Y.Z`.

## Creating a Release

1.  **Finalize Changes**: Ensure that all changes for the release are merged into the `main` branch.

2.  **Update Dependencies**: If any dependencies have changed, update the `vendor` directory.

    ```bash
    go mod vendor
    ```

3.  **Determine the Version Number**: Based on the changes since the last release, determine the next version number.

4.  **Create an Annotated Tag**: Create a new annotated git tag for the release.

    ```bash
    git tag -a vX.Y.Z -m "Release vX.Y.Z"
    ```

5.  **Push the Tag**: Push the new tag to the remote repository.

    ```bash
    git push origin vX.Y.Z
    ```

    This will trigger a new release on GitHub.

## Building with Version Information

The `tfschema` command-line tool includes a `-version` flag to display the current version. To build the binary with the correct version information, use the `-ldflags` option and the `-mod=vendor` flag to build from the `vendor` directory:

```bash
go build -ldflags="-X main.version=vX.Y.Z" -mod=vendor ./cmd/tfschema
```

When you run the built binary with the `-version` flag, it will display the version you specified:

```bash
./tfschema -version
# Output: vX.Y.Z
```

## Updating a Release

If you need to update a release after it has been tagged, follow these steps:

1.  **Commit Changes**: Commit any new changes to the `main` branch.

2.  **Delete the Local Tag**:

    ```bash
    git tag -d vX.Y.Z
    ```

3.  **Delete the Remote Tag**:

    ```bash
    git push --delete origin vX.Y.Z
    ```

4.  **Re-create the Tag**: Create the tag again on the latest commit.

    ```bash
    git tag -a vX.Y.Z -m "Release vX.Y.Z (updated)"
    ```

5.  **Push the New Tag**:
    ```bash
    git push origin vX.Y.Z
    ```
