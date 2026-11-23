package mock_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/vektra/mockery/v3/internal/fixtures/matryer_call_history"
)

func TestRecordCalls(t *testing.T) {
	recorder := &MoqRecorder{}
	entry := mock.Entry{ID: 42}
	recorder.Record(entry)
	calls := recorder.RecordCalls()
	require.Len(t, calls, 1)
	require.Equal(t, entry, calls[0].Entry)
}
