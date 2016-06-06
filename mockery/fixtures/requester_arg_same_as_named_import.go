package test

import json "encoding/json"

type RequesterArgSameAsNamedImport interface {
	Get(json string) *json.RawMessage
}
