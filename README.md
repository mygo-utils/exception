# Exception

**Exception** is a lightweight, production-ready Go package that simulates exception handling using panic and recover. It provides a simple way to capture and report errors along with a full stack trace, similar to try/catch/finally constructs in other languages.

## Features

- Create and propagate exceptions with custom messages.
- Wrap existing errors to add stack traces.
- Retrieve full exception details including a formatted stack trace.
- Write exception details to a file.
- Simulate try/catch and try/catch/finally constructs.
- Conditional throw based on a boolean flag.

## Installation

To install the package, use:

```bash
go get -u github.com/mygo-utils/exception
```

To use the package in your code, import it as:

```go
package main

import (
    "fmt"
    "log"

    "github.com/mygo-utils/exception"
)

func main() {
    // Example 1: Using TryCatch
    result := exception.TryCatch(func() any {
        fmt.Println("Entering try block; about to throw an exception.")
        exception.Throw("Fatal error occurred")
        return nil // not reached
    }, func(e *exception.Exception) any {
        fmt.Println("Caught exception:")
        fmt.Println(e.FullDetails())
        // Optionally, write the details to a file.
        if err := e.DumpToFile("exception.log"); err != nil {
            log.Printf("Error writing to file: %v", err)
        }
        return "Recovered"
    })
    fmt.Println("Result from Example 1:", result)

    // Example 2: Using TryCatchFinally with error condition
    result = exception.TryCatchFinally(func() any {
        fmt.Println("Entering try block; testing error condition.")
        a, b := 100, 0
        exception.ThrowIf(b == 0, "Division by zero error")
        return a / b
    }, func(e *exception.Exception) any {
        fmt.Println("Caught exception:")
        fmt.Println(e.FullDetails())
        return nil
    }, func() {
        fmt.Println("Finally block executed: cleanup complete.")
    })
    fmt.Println("Result from Example 2:", result)
}
```
