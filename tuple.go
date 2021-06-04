package fp

type Tuple struct {
	E1, E2 interface{}
}

func TupleOf(e1, e2 interface{}) Tuple {
	return Tuple{E1: e1, E2: e2}
}

type TupleString struct{ E1, E2 string }

func TupleStringOf(e1, e2 string) TupleString { return TupleString{E1: e1, E2: e2} }

type TupleStringInt struct {
	E1 string
	E2 int
}

func TupleStringIntOf(e1 string, e2 int) TupleStringInt { return TupleStringInt{E1: e1, E2: e2} }

type TupleStringAny struct {
	E1 string
	E2 interface{}
}

func TupleStringAnyOf(e1 string, e2 interface{}) TupleStringAny {
	return TupleStringAny{E1: e1, E2: e2}
}

type TupleStringStrings struct {
	E1 string
	E2 []string
}

func TupleStringStringsOf(e1 string, e2 []string) TupleStringStrings {
	return TupleStringStrings{E1: e1, E2: e2}
}
