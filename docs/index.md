mockery
========

Mockery is a project that creates mock implementations of Golang interfaces. The mocks generated in this project are based off of the [github.com/stretchr/testify](https://github.com/stretchr/testify) suite of testing packages.

![](assets/images/demo.gif)
![](assets/images/MockScreenshot.png)

Why mockery?
-------------

When you have an interface like this:

```golang title="db.go"
type DB interface {
	Get(val string) string
}
```

and a function that takes this interface:

```golang title="db_getter.go"
func getFromDB(db DB) string {
	return db.Get("ice cream")
}
```

You can test `getFromDB` by either instantiating a testing database, or you can simply create a mock implementation of `DB` using mockery. Mockery can autogenerate a mock implementation that allows us to define assertions on how the mock was used, what to return, and other useful tidbits. We can add a `//go:generate` directive above our interface:

```golang title="db.go"
//go:generate mockery --name DB
type DB interface {
	Get(val string) string
}
```

```yaml title=".mockery.yaml"
inpackage: True # (1)!
with-expecter: True # (2)!
testonly: True # (3)!
```

1. Generate our mocks next to the original interface
2. Create [expecter methods](/mockery/features/#expecter-structs)
3. Append `_test.go` to the filename so the mock object is not packaged 

```bash
$ go generate  
05 Mar 23 21:49 CST INF Starting mockery dry-run=false version=v2.20.0
05 Mar 23 21:49 CST INF Using config: .mockery.yaml dry-run=false version=v2.20.0
05 Mar 23 21:49 CST INF Walking dry-run=false version=v2.20.0
05 Mar 23 21:49 CST INF Generating mock dry-run=false interface=DB qualified-name=github.com/vektra/mockery/v2/pkg/fixtures/example_project version=v2.20.0
```

We can then use the mock object in a test:

```go title="db_getter_test.go"
import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_getFromDB(t *testing.T) {
	mockDB := NewMockDB(t)
	mockDB.EXPECT().Get("ice cream").Return("chocolate").Once()
	flavor := getFromDB(mockDB)
	assert.Equal(t, "chocolate", flavor)
}
```

Why use mockery over gomock?
-----------------------------

1. mockery provides a much more user-friendly API and is less confusing to use
2. mockery utilizes `testify` which is a robust and highly feature-rich testing framework
3. mockery has rich configuration options that allow fine-grained control over how your mocks are generated
4. mockery's CLI is more robust, user-friendly, and provides many more options
5. mockery supports generics (this may no longer be an advantage if/when gomock supports generics)

Who uses mockery?
------------------

:simple-grafana: [grafana](https://github.com/grafana/grafana) · :simple-google: [Google Skia](https://github.com/google/skia) · [Hashicorp](https://github.com/search?q=org%3Ahashicorp%20mockery&type=code) · :simple-google: [Google Skyzkaller](https://github.com/google/syzkaller) · :fontawesome-brands-uber: [Uber Cadence](https://github.com/uber/cadence) · [Jaeger](https://github.com/jaegertracing/jaeger) · [Splunk](https://github.com/splunk/kafka-mq-go) · [Ignite CLI](https://github.com/ignite/cli) · [Tendermint](https://github.com/tendermint/tendermint) · [Datadog](https://github.com/DataDog/datadog-agent)


[Get Started](/mockery/installation/){ .md-button .md-button--primary .md-button--stretch }
