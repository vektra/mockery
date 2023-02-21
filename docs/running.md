Running
========

Using `go generate` <small>recommended</small>
--------------------

`go generate` is often preferred as it give you more targeted generation of specific interfaces. Use `generate` as a directive above the interface you want to generate a mock for.

``` golang
package example_project

//go:generate mockery --name Root
type Root interface {
        Foobar(s string) error
}
```

Then simply:

``` bash
$ go generate      
09 Feb 23 22:55 CST INF Starting mockery dry-run=false version=v2.18.0
09 Feb 23 22:55 CST INF Using config: /Users/landonclipp/git/LandonTClipp/mockery/.mockery.yaml dry-run=false version=v2.18.0
09 Feb 23 22:55 CST INF Walking dry-run=false version=v2.18.0
09 Feb 23 22:55 CST INF Generating mock dry-run=false interface=Root qualified-name=github.com/vektra/mockery/v2/pkg/fixtures/example_project version=v2.18.0
```

For all interfaces in project
------------------------------

If you provide `all: True`, you can generate mocks for the entire project. This is not recommended for larger projects as it can take a large amount of time parsing packages to generate mocks that you might never use.

```bash
$ mockery
09 Feb 23 22:47 CST INF Starting mockery dry-run=false version=v2.18.0
09 Feb 23 22:47 CST INF Using config: /Users/landonclipp/git/LandonTClipp/mockery/.mockery.yaml dry-run=false version=v2.18.0
09 Feb 23 22:47 CST INF Walking dry-run=false version=v2.18.0
09 Feb 23 22:47 CST INF Generating mock dry-run=false interface=A qualified-name=github.com/vektra/mockery/v2/pkg/fixtures version=v2.18.0
```

!!! note

        Note that in some cases, using `//go:generate` may turn out to be slower than running for entire packages. `go:generate` calls mockery once for each `generate` directive, which means that mockery may need to parse the package multiple times, which is wasteful. Good judgement is recommended when determining the best option for your own project.

!!! note

        For mockery to correctly generate mocks, the command has to be run on a module (i.e. your project has to have a go.mod file)