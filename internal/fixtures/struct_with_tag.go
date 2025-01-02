package test

type StructWithTag interface {
	MethodA(v *struct {
		FieldA int `json:"field_a"`
		FieldB int `json:"field_b" xml:"field_b"`
	}) *struct {
		FieldC int `json:"field_c"`
		FieldD int `json:"field_d" xml:"field_d"`
	}
}
