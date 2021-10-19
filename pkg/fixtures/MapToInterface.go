package test

// SKIP MOCK
type MapToInt interface {
	// SKIP
	Foo(arg1 ...map[string]interface{}) // TAMBIEN
}

// SKIP MOCK2
type MapToInt2 interface {
	// SKIP2
	Foobar(arg1 ...map[string]interface{}) // TAMBIEN2
}
