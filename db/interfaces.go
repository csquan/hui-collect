package db

type IReader interface {
}

type IWriter interface {
}

type IDB interface {
	IReader
	IWriter
}
