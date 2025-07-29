package privatetypes

type (
	private int

	privateInterface interface {
		A()
	}

	PrivateMethod interface {
		a()
	}

	PrivateParam interface {
		B(private)
	}

	PrivateReturn interface {
		C() private
	}

	Public interface {
		A()
	}

	PrivateVariadic interface {
		A(...private)
	}

	PrivateChannel interface {
		A(chan private)
	}

	PrivateMapValue interface {
		A(map[string]private)
	}

	PrivateMapKey interface {
		A(map[private]string)
	}

	PrivateSlice interface {
		A([]private)
	}

	PrivatePointer interface {
		A(*private)
	}
)
