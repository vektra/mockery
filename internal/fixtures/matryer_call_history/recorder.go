package mock

type Entry struct{ ID int }

type Recorder interface {
	Record(entry Entry)
}
