# Contribution Guidelines

## Contributing to Idl
To contribute to the IDL, please follow these steps:
1. Fork the repository.
2. Make your changes in a new branch.
3. Run `make buf` to generate idl changes.
4. Commit (and Sign) your changes with a clear message.
5. Push your changes to your fork.
6. Create a pull request against the `main` branch of the original repository.

### Releasing
1. Ensure that your changes are merged into the `main` branch.
2. Create a new release by strictly following the SemVer 2.0.0 guidelines.
   This guarantees Python, Rust, NPM, and Go versions are in sync.
