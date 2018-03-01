package middleware

//实体池
type Pool interface {
	Tack() (Entity, error)
	Return(pl Entity) error
	Total() uint32
	Used() uint32
}

type Entity interface {
	Id() uint32
}
