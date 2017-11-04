package main

import "io"

type EOLStat struct {
	CR, CRLF, LF int
	NUL          int
}

func (s EOLStat) IsMixed() bool {
	n := 0
	if s.CR != 0 {
		n++
	}
	if s.CRLF != 0 {
		n++
	}
	if s.LF != 0 {
		n++
	}
	return n > 1
}

func (s EOLStat) IsCR() bool   { return s.CRLF == 0 && s.LF == 0 }
func (s EOLStat) IsCRLF() bool { return s.CR == 0 && s.LF == 0 }
func (s EOLStat) IsLF() bool   { return s.CR == 0 && s.CRLF == 0 }

func (s EOLStat) String() string {
	if s.NUL != 0 {
		return "binary"
	}
	var eols string
	f := func(n int, eol string) {
		if n != 0 {
			var pfx string
			if eols != "" {
				pfx = " "
			}
			eols += pfx + eol
		}
	}
	f(s.CR, "CR")
	f(s.CRLF, "CRLF")
	f(s.LF, "LF")
	return eols
}

func CalcEOLStat(p []byte) EOLStat {
	var s EOLStat
	for len(p) != 0 {
		var part []byte
		part, p = SplitEOL(p)
		switch {
		case len(part) == 0:
			// pass
		case len(part) == 2 && part[0] == '\r' && part[1] == '\n':
			s.CRLF++
		case part[0] == '\n':
			s.LF++
		case part[0] == '\r':
			s.CR++
		case part[0] == 0:
			s.NUL++
		}
	}
	return s
}

func CalcEOLStatReader(r io.Reader) (EOLStat, error) {
	var s EOLStat
	buf := make([]byte, 1024*1024)
	keep := 0
	for {
		n, err := r.Read(buf[keep:])
		n += keep
		m := n
		if n > 0 && buf[n-1] == '\r' && err == nil {
			n--
		}
		sx := CalcEOLStat(buf[:n])
		s.CR += sx.CR
		s.CRLF += sx.CRLF
		s.LF += sx.LF
		s.NUL += sx.NUL
		keep = copy(buf, buf[n:m])
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			return s, err
		}
	}
	panic("not reached")
}

// SplitEOL splits p into part and rest.
// Part is either:
//  - a single EOL ('\r', '\n', '\r\n')
//  - a single NUL byte
//  - a sequence of bytes other than NUL or EOL
// Rest is the remaining bytes in p just after part.
func SplitEOL(p []byte) (part, rest []byte) {
	switch {
	case len(p) == 0:
		return p, p
	case p[0] == '\n' || p[0] == 0:
		return p[:1], p[1:]
	case p[0] == '\r':
		if len(p) > 1 && p[1] == '\n' {
			return p[:2], p[2:]
		}
		return p[:1], p[1:]
	}

	for i, b := range p {
		if b == '\r' || b == '\n' || b == 0 {
			return p[:i], p[i:]
		}
	}
	return p, nil
}
