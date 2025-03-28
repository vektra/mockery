Template
--------

This package contains all of the data passed to mockery templates. The top-most variable provided
to the templates is [Data](https://pkg.go.dev/github.com/vektra/mockery/v3/template#Data). Attributes
of this struct can be accessed using syntax like:

```
package {{.PkgName}}

import (
{{- range .Imports}}
	{{. | importStatement}}
{{- end}}
    mock "github.com/stretchr/testify/mock"
)
```

Further examples of how to use the data provided to mockery templates can be found in the pre-curated mocks, such as:

- [matryer](https://github.com/vektra/mockery/blob/v3/internal/mock_matryer.templ)
- [testify](https://github.com/vektra/mockery/blob/v3/internal/mock_testify.templ)


Full documentation is provided at: https://vektra.github.io/mockery/v3/