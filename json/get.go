package json

import (
	"crypto/sha1"
)

const (
	expectKey int = iota
	expectKeyClose
	expectColon
	expectValue
	expectString
	expectListClose
	expectObjClose
	expectBoolClose
	expectNumClose
)

type Field struct {
	Key   []byte
	Value []byte
}

func Get(b []byte, keys [][]byte) []Field {
	s := 0
	state := expectKey

	kmap := make(map[[20]byte]struct{}, len(keys))

	for _, k := range keys {
		h := sha1.Sum(k)
		if _, ok := kmap[h]; !ok {
			kmap[h] = struct{}{}
		}
	}

	l := 10
	res := make([]Field, l)

	for i := 0; i < len(b); i++ {
		switch {
		case state == expectKey && b[i] == '"':
			state = expectKeyClose
			s = i + 1
			continue

		case state == expectKeyClose && b[i] == '"':
			state = expectColon
		}

		if state != expectColon {
			continue
		}

		k := b[s:i]
		h := sha1.Sum(k)
		_, kf := kmap[h]

		e := 0
		d := 0
		for ; i < len(b); i++ {
			if state == expectObjClose || state == expectListClose {
				switch b[i] {
				case '{', '[':
					d++
				case '}', ']':
					d--
				}
			}

			switch {
			case state == expectColon && b[i] == ':':
				state = expectValue

			case state == expectValue && b[i] == '"':
				state = expectString
				s = i

			case state == expectString && b[i] == '"':
				e = i

			case state == expectValue && b[i] == '[':
				state = expectListClose
				s = i
				d++

			case state == expectListClose && d == 0 && b[i] == ']':
				e = i
				i = s

			case state == expectValue && b[i] == '{':
				state = expectObjClose
				s = i
				d++

			case state == expectObjClose && d == 0 && b[i] == '}':
				e = i
				i = s

			case state == expectValue && (b[i] >= '0' && b[i] <= '9'):
				state = expectNumClose
				s = i

			case state == expectNumClose &&
				((b[i] < '0' || b[i] > '9') &&
					(b[i] != '.' && b[i] != 'e' && b[i] != 'E' && b[i] != '+' && b[i] != '-')):
				i--
				e = i

			case state == expectValue &&
				(b[i] == 'f' || b[i] == 'F' || b[i] == 't' || b[i] == 'T'):
				state = expectBoolClose
				s = i

			case state == expectBoolClose && (b[i] == 'e' || b[i] == 'E'):
				e = i
			}

			if e != 0 {
				e++

				if kf {
					if len(res) == cap(res) {
						r := make([]Field, 0, (len(res) + l))
						copy(r, res)
						res = r
					}

					res = append(res, Field{k, b[s:e]})
				}

				state = expectKey
				break
			}
		}
	}

	return res
}