package test

type StructWithTag interface {
	MethodA(v *struct {
		FieldA int `json:"field_a"`
		FieldB int `json:"field_b" xml:"field_b"`
	}) *struct {
		FieldA int `json:"field_a"`
		FieldB int `json:"field_b" xml:"field_b"`
	}
}
