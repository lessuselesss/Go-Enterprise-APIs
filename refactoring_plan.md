# Go-Enterprise-APIs Refactoring Plan

**Goal:** Refactor the `Go-Enterprise-APIs` project into a more idiomatic Go API client library structure, keeping the existing struct names `CCertificate` and `CEPAccount`, consolidating crypto utilities into the utils package, and moving test definitions to a root-level test directory.

**Current State:**
The project structure has been partially refactored. The core logic files have been moved to the `models`, `services`, `internal`, and `examples` directories.
*   [`models/certificate.go`](models/certificate.go) contains the `CCertificate` struct and methods.
*   [`services/account.go`](services/account.go) contains the `CEPAccount` struct and methods, including network interaction logic.
*   [`internal/crypto.go`](internal/crypto.go) contains cryptographic helper functions.
*   [`internal/utils.go`](internal/utils.go) contains general utility functions and constants.
*   [`examples/main.go`](examples/main.go) exists and attempts to use the moved packages (though imports need updating).
*   [`lib/testdefs/test_types.go`](lib/testdefs/test_types.go) remains in its original location.

**Remaining Tasks:**

1.  Create top-level client files (`client.go`, `config.go`, `errors.go`).
2.  Define the main client struct (`CircularClient`) and its constructor.
3.  Define and implement client configuration.
4.  Define custom error types.
5.  Extract and centralize HTTP request logic into an internal package.
6.  Consolidate crypto utilities into the utils package.
7.  Update the `CEPAccount` service to use the new HTTP helper and the consolidated `internal/utils` package.
8.  Integrate the `CEPAccount` service into the main client struct.
9.  Update the example code to use the new client structure.
10. Move test definitions to a root-level test directory.
11. Update imports and references throughout the codebase.

**Detailed Plan:**

1.  **Create New Top-Level Files:**
    *   Create an empty file named [`client.go`](client.go) at the root of the project.
    *   Create an empty file named [`config.go`](config.go) at the root of the project.
    *   Create an empty file named [`errors.go`](errors.go) at the root of the project.
    *   Create an empty file named [`internal/http.go`](internal/http.go).

2.  **Define Client Configuration (`config.go`):**
    *   In [`config.go`](config.go), define a struct (e.g., `Config`) to hold client configuration options like `NagURL`, `NetworkNode`, `Blockchain`, `IntervalSec`, etc.
    *   Define a constructor function for the `Config` struct, potentially with default values.

3.  **Define Custom Errors (`errors.go`):**
    *   In [`errors.go`](errors.go), define custom error types (e.g., `NetworkError`, `APIError`, `InvalidResponseError`) that implement the `error` interface. This will provide more structured error handling.

4.  **Define Main Client (`client.go`):**
    *   In [`client.go`](client.go), define the main client struct (e.g., `CircularClient`).
    *   Include the `Config` struct as a field in `CircularClient`.
    *   Include an instance of the `CEPAccount` service (or a pointer to it) as a field in `CircularClient`.
    *   Define a constructor function `NewCircularClient(config Config)` that initializes the `CircularClient` struct and its embedded services.

5.  **Extract HTTP Logic (`internal/http.go`):**
    *   Move the HTTP request logic (using `net/http`, `io/ioutil`, `bytes`, `encoding/json`) from [`services/account.go`](services/account.go) to [`internal/http.go`](internal/http.go).
    *   Create helper functions in [`internal/http.go`](internal/http.go) for common HTTP operations, such as `PostJSON(url string, payload interface{}) ([]byte, error)` and `GetJSON(url string) ([]byte, error)`. These functions should handle the request execution, response reading, and basic error checking (like status codes).
    *   Update imports in [`internal/http.go`](internal/http.go) as needed.

6.  **Consolidate Crypto into Utils (`internal/utils.go`):**
    *   Move the `Sign` and `GetPublicKey` functions from [`internal/crypto.go`](internal/crypto.go) into [`internal/utils.go`](internal/utils.go).
    *   Delete the [`internal/crypto.go`](internal/crypto.go) file.

7.  **Update `CEPAccount` Service (`services/account.go`):**
    *   Remove the HTTP request logic that was moved to [`internal/http.go`](internal/http.go).
    *   Modify the methods in `CEPAccount` (e.g., `SetNetwork`, `UpdateAccount`, `SubmitCertificate`, `GetTransaction`, `GetTransactionOutcome`, `WaitForTransactionOutcome`) to use the new HTTP helper functions from [`internal/http.go`](internal/http.go) and the consolidated utility functions from [`internal/utils.go`](internal/utils.go).
    *   Update the error handling in `CEPAccount` methods to return the custom error types defined in [`errors.go`](errors.go) instead of just string messages or generic errors.
    *   Remove configuration fields (`nagUrl`, `networkNode`, `blockchain`, `intervalSec`) from the `CEPAccount` struct and update methods to accept or access configuration via the main client struct (this will be done in step 8).
    *   Update imports in [`services/account.go`](services/account.go) to include `circular-api/internal/http` and `circular-api/errors`.

8.  **Integrate Service into Client (`client.go`):**
    *   In the `NewCircularClient` constructor in [`client.go`](client.go), initialize the `CEPAccount` service and pass the relevant configuration values from the `Config` struct to it (or ensure the `CEPAccount` struct has access to the config).
    *   Add methods to `CircularClient` that delegate calls to the embedded `CEPAccount` service (e.g., `client.Account.Open(...)`, `client.Account.SubmitCertificate(...)`).

9.  **Update Example (`examples/main.go`):**
    *   Update the imports in [`examples/main.go`](examples/main.go) to use the new top-level client package (e.g., `circular-api`).
    *   Modify the example code to create a `CircularClient` instance using the `NewCircularClient` constructor and the `Config` struct.
    *   Update the calls to `CCertificate` and `CEPAccount` functionality to go through the `CircularClient` instance (e.g., `client.Account.Open(...)`).

10. **Move Test Definitions:**
    *   Create a new directory named `test/` at the root of the project.
    *   Move the file [`lib/testdefs/test_types.go`](lib/testdefs/test_types.go) to the new `test/` directory.

11. **Update Internal References:**
    *   Review all Go files in the project and update any remaining import paths or references to structs/functions that have been moved or renamed. Ensure consistency with the new package structure. This includes updating imports that previously referenced `circular-api/internal/crypto` to now reference `circular-api/internal/utils`, and imports that referenced `circular-api/lib/testdefs` to now reference `circular-api/test`.

**Final Proposed Structure Visualization (Mermaid Diagram):**

```mermaid
graph TD
    A[Go-Enterprise-APIs/] --> B(go.mod)
    A --> C(go.sum)
    A --> D(README.md)
    A --> E(client.go)
    A --> F(config.go)
    A --> G(errors.go)
    A --> H(models/)
    A --> I(services/)
    A --> J(internal/)
    A --> K(examples/)
    A --> L(test/)

    H --> H1(certificate.go)
    H --> H2(certificate_test.go)

    I --> I1(account.go)
    I --> I2(account_test.go)

    J --> J1(http.go)
    J --> J2(utils.go)

    K --> K1(main.go)

    L --> L1(test_types.go)

    E --> F
    E --> I1
    E --> H1
    I1 --> J1
    I1 --> J2
    I1 --> H1
    I1 --> G
    K1 --> E
    H1 --> J2