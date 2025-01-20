---
title: replace-type
---

## Description

The `#!yaml replace-type:` parameter allows you to replace a type in the generated mocks with another type. Take for example the following interface:


```go title="interface.go"
package replace_type

import (
    "github.com/vektra/mockery/v3/internal/fixtures/example_project/replace_type/rti/rt1"
    "github.com/vektra/mockery/v3/internal/fixtures/example_project/replace_type/rti/rt2"
)

type RType interface {
    Replace1(f rt1.RType1)
}
```

You can selectively replace the `rt1.RType1` with a new type if so desired. For example:

```yaml title=".mockery.yml"
replace-type:
  github.com/vektra/mockery/v3/internal/fixtures/example_project/replace_type/rti/rt1:
    RType1:
      pkg-path: github.com/vektra/mockery/v3/internal/fixtures/example_project/replace_type/rti/rt2
      type-name: RType2
```

The mock will now replace all instances of `rt1.RType1` with `rt2.RType2`. You can see the before and after of `mockery`-style mocks:

=== "before"

    ```go
    // Replace2 provides a mock function for the type RTypeReplaced1
    func (_mock *RTypeReplaced1) Replace1(f rt1.RType1) {
        _mock.Called(f)
        return
    }
    ```

=== "after"

    ```go 
    // Replace2 provides a mock function for the type RTypeReplaced1
    func (_mock *RTypeReplaced1) Replace1(f rt2.RType2) {
        _mock.Called(f)
        return
    }
    ```

## Background

This parameter is useful if you need to need to work around packages that use internal types. Take for example the situation found [here](https://github.com/vektra/mockery/issues/864#issuecomment-2567788637), noted by [RangelReale](https://github.com/RangelReale).
