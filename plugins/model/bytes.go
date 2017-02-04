package model

type Bytes []byte

func (b Bytes) Length() int {
	return len(b)
}

func (b Bytes) String() string {
	return string(b)
}
