---
name: golang-testify-suite
description: Use when organizing related Go tests into a suite with shared lifecycle hooks (Setup/Teardown) using stretchr/testify.
---

# Golang Testify Suite Setup

## Overview
The `testify/suite` package allows you to group related tests into a struct, enabling shared setup/teardown logic and state. This is especially useful for integration tests or complex unit tests with expensive resource initialization.

## When to Use
- **Shared Resources**: Database connections, mock servers, or heavy configuration shared across multiple tests.
- **Ordered Lifecycle**: When you need precise control over what happens before/after the whole suite vs. each individual test.
- **Resource Cleanup**: Ensuring temp files or DB records are deleted even if a test panics.

## Core Pattern
```go
import (
    "testing"
    "github.com/stretchr/testify/suite"
)

// 1. Define the suite struct
type MyServiceSuite struct {
    suite.Suite
    DB *MockDB
}

// 2. Implement lifecycle hooks
func (s *MyServiceSuite) SetupSuite() {
    // Runs once before all tests
    s.DB = NewMockDB()
}

func (s *MyServiceSuite) SetupTest() {
    // Runs before each test
    // Re-initialize mocks to ensure a clean slate
    s.DB = new(MockDB)
}

// 3. Write test methods (must start with "Test")
func (s *MyServiceSuite) TestFeatureA() {
    s.NoError(s.DB.Insert("data"))
    s.Equal(1, s.DB.Count())
}

// 4. Create the entry point
func TestMyServiceSuite(t *testing.T) {
    suite.Run(t, new(MyServiceSuite))
}
```

## Lifecycle Order
1. `SetupSuite()`: Once before any tests run.
2. `BeforeTest(suite, test)`: Before each test (optional, receives names).
3. `SetupTest()`: Before each test.
4. **`TestXxx()`**: The actual test method.
5. `TearDownTest()`: After each test.
6. `AfterTest(suite, test)`: After each test (optional).
7. `TearDownSuite()`: Once after all tests finish.

## Quick Reference
| Feature | Suite Method | Equivalent |
|---------|--------------|------------|
| Assert | `s.Equal(a, b)` | `assert.Equal(t, a, b)` |
| Require | `s.Require().Nil(err)` | `require.Nil(t, err)` |
| Testing T | `s.T()` | `t` |
| Fail Now | `s.FailNow("msg")` | `t.FailNow()` |

## Common Mistakes
- **Forgetting `suite.Run`**: The Go test runner will NOT find your suite methods without a standard `TestXxx(t *testing.T)` entry point.
- **Value Receiver**: Using `(s MySuite)` instead of `(s *MySuite)` prevents state sharing between hooks and tests.
- **Expensive `SetupTest`**: Putting DB migrations in `SetupTest` instead of `SetupSuite` will drastically slow down your tests.
- **Parallel Tests**: `testify/suite` does NOT natively support parallel execution of methods within the same suite struct. Using `t.Parallel()` inside a suite test method can lead to race conditions on the suite's fields.

## Advanced: Table-Driven Tests in Suites
Use `s.Run()` to define subtests within a suite method, similar to `t.Run()`:
```go
func (s *MyServiceSuite) TestMultipleCases() {
    cases := []struct{ name string; input string }{...}
    for _, tc := range cases {
        s.Run(tc.name, func() {
            // s.T() is automatically updated to the subtest context
            s.Equal(expected, actual)
        })
    }
}
```

## Using Context7 for Advanced Patterns
If you need more advanced patterns (e.g., parallel suites, nested suites, or specific mock integrations), use the `context7` tool:
```bash
# Resolve the library ID first
mcp__context7__resolve-library-id(libraryName="testify")

# Query specific suite features
mcp__context7__query-docs(libraryId="/stretchr/testify", query="how to run testify suites in parallel")
```
