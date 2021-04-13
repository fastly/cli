## Testing

To run the test suite:

```sh
make test
```

Note that by default the tests are run using `go test` with the following configuration:

```
-race ./{cmd,pkg}/...
```

To run a specific test use the `-run` flag (exposed by `go test`) and also provide the path to the directory where the test files reside (replace `...` and `<path>` with appropriate values):

```sh
make test TESTARGS="-run <...> <path>"
```

**Example**:

```sh
make test TESTARGS="-run TestBackendCreate ./pkg/backend/..."
```

Some integration tests aren't run outside of the CI environment, to enable these tests locally you'll need to set a specific environment variable relevant to the test.

The available environment variables are:

- `TEST_COMPUTE_INIT`: runs `TestInit`.
- `TEST_COMPUTE_BUILD`: runs `TestBuildRust` and `TestBuildAssemblyScript`.
- `TEST_COMPUTE_BUILD_RUST`: runs `TestBuildRust`.
- `TEST_COMPUTE_BUILD_ASSEMBLYSCRIPT`: runs `TestBuildAssemblyScript`.

**Example**:

```sh
TEST_COMPUTE_BUILD_RUST=1 make test TESTARGS="-run TestBuildRust/fastly_crate_prerelease ./pkg/compute/..." 
```

When running the tests locally, if you don't have the relevant language ecosystems set-up properly then the tests will fail to run and you'll need to review the code to see what the remediation steps are, as that output doesn't get shown when running the test suite.

> **NOTE**: you might notice a discrepancy between CI and your local environment which is caused by the difference in Rust toolchain versions as defined in .github/workflows/pr_test.yml which specifies the version required to be tested for in CI. Running `rustup toolchain install <version>` and `rustup target add wasm32-wasi --toolchain <version>` will resolve any failing integration tests you may be running locally.
