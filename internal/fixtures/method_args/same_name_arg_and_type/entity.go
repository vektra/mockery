package same_name_arg_and_type

type (
	interfaceA interface {
		// SomeMethod - contains args with the same names of the type and arg
		DoB(interfaceB interfaceB) interfaceB
		DoB0(interfaceB interfaceB0) interfaceB0
		DoB0v2(interfaceB0 interfaceB0) interfaceB0
	}

	interfaceB interface {
		GetData() int
	}

	interfaceB0 interface {
		DoB0(interfaceB0 interfaceB0) interfaceB0
	}
)
