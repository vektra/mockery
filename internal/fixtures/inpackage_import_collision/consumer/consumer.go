package dep

import libdep "github.com/vektra/mockery/v3/internal/fixtures/inpackage_import_collision/dep"

type dep struct{}

type Iface interface {
	M() libdep.T
}
