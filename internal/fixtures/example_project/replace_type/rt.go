package replace_type

import (
	"github.com/vektra/mockery/v3/internal/fixtures/example_project/replace_type/rti/rt1"
	"github.com/vektra/mockery/v3/internal/fixtures/example_project/replace_type/rti/rt2"
)

type RType interface {
	Replace1(f rt1.RType1)
	Replace2(f rt2.RType2)
}
