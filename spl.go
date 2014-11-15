package spl

import (
	"io"
)

// TODO: Implement INTEGER and BLOB.

type SeqParser struct {
	_line    int
	_column  int
	_end     bool
	_current [1]byte
	_reader  io.Reader
}

func NewSeqParser(input io.Reader) *SeqParser {
	p := new(SeqParser)
	p._reader = input
	p.shift(1)
	return p
}

func (p *SeqParser) shift(count int) {
	for i := 0; !p._end && i < count; i++ {
		if p._current[0] == '\n' {
			p._line++
			p._column = 0
		}

		n, err := p._reader.Read(p._current[:])
		if n < 1 || err != nil {
			p._current[0] = 0
			p._end = true
		} else {
			p._column++
		}
	}
}

func (p *SeqParser) isEOF() bool {
	return p._end
}

func (p *SeqParser) current() byte {
	if p._end {
		return 0
	}
	return p._current[0]
}

func (p *SeqParser) skipSpace() {
	for p.current() == ' ' || p.current() == '\t' || p.current() == '\r' || p.current() == '\n' {
		p.shift(1)
	}
}

func (p *SeqParser) Line() int {
	return p._line
}

func (p *SeqParser) Column() int {
	return p._column
}

func (p *SeqParser) IsList() bool {
	return p.current() == '('
}

func (p *SeqParser) IsString() bool {
	return p.current() == '"'
}

func (p *SeqParser) IsEnd() bool {
	return p.isEOF() || p.current() == ')'
}

func (p *SeqParser) Down() {
	p.shift(1)
	p.skipSpace()
}

func (p *SeqParser) Up() {
	for !p.IsEnd() {
		p.Skip()
	}

	p.shift(1)
	p.skipSpace()
}

func (p *SeqParser) Skip() {
	switch {
	case p.IsList():
		p.Down()
		p.Up()

	case p.IsString():
		p.skipString()

	case p.IsEnd():
		// Nothing.

	default:
		// TODO: Remove panic() in favor of returning errors.
		panic("Bad format in SPL file.")
	}
}

func (p *SeqParser) skipString() {
	p.shift(1)

	for {
		if p.isEOF() {
			panic("End of file within a string.")
		}

		c := p.current()
		p.shift(1)

		switch c {
		case '"':
			p.skipSpace()
			return

		case '\\':
			switch p.current() {
			case '"', '\\', 'n', 'r':
				p.shift(1)
			case 'x':
				// TODO: validate escape sequences.
				p.shift(3)
			case 'u':
				p.shift(5)
			case 'U':
				p.shift(9)
			}
		}
	}
}

func unhex(h []byte) (result uint) {
	for _, d := range h {
		switch {
		case d >= '0' && d <= '9':
			result = result*16 + uint(d-'0')
		case d >= 'a' && d <= 'f':
			result = result*16 + 10 + uint(d-'a')
		case d >= 'A' && d <= 'F':
			result = result*16 + 10 + uint(d-'A')
		default:
			panic("not a hex digit")
		}
	}

	return result
}

func (p *SeqParser) String() string {
	if p.current() != '"' {
		panic("Not a string")
	}
	p.shift(1)

	str := make([]byte, 0, 8)

	for {
		if p.isEOF() {
			panic("End of file within a string.")
		}

		c := p.current()
		p.shift(1)

		switch c {
		case '"':
			p.skipSpace()
			return string(str)

		case '\\':
			switch p.current() {
			case '"', '\\':
				str = append(str, p.current())
				p.shift(1)
			case 'n':
				str = append(str, '\n')
				p.shift(1)
			case 'r':
				str = append(str, '\r')
				p.shift(1)
			case 'x':
				h := []byte{0, 0}
				p.shift(1)
				h[0] = p.current()
				p.shift(1)
				h[1] = p.current()
				p.shift(1)
				str = append(str, byte(unhex(h)))
			case 'u':
				// TODO
				panic("Not implemented yet")
			case 'U':
				// TODO
				panic("Not implemented yet")
			}
		default:
			str = append(str, c)
		}
	}
}
