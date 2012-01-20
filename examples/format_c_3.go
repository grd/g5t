package main

/* Test program, used by the format-c-3 test.
   Copyright (C) 2002, 2009 Free Software Foundation, Inc.

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU General Public License as published by
   the Free Software Foundation; either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU General Public License for more details.

   You should have received a copy of the GNU General Public License
   along with this program.  If not, see <http://www.gnu.org/licenses/>.  */


import (
	"os"
	"fmt"
	"bytes"
	"github.com/GerardNL/g5t"
)

func G(msgid string) string {
	return gt.String(msgid)
}

func GN(msgid1, msgid2 string, nr int) string {
	return gt.StringN(msgid1, msgid2, nr)
}

func main() {
	var n = 5

	gt.Setup("translations", "/usr/share/locale", "de", gt.GettextParser)

	s := G("father of %d children")
	s2 := GN("father of one kid", "father of %d children", n)
	fmt.Printf(s2, n)
	fmt.Println(" ")
	c1 := "Vater von %"
	c2 := " Kindern"

	if !(len(s) > len(c1)+len(c2) && s[0:len(c1)] == c1 && s[len(s)-len(c2):] == c2) {
		fmt.Println(s)
		fmt.Fprintf(os.Stderr, "String not translated.\n")
		os.Exit(1)
	}
	if bytes.IndexByte([]byte(s), '<') != -1 || bytes.IndexByte([]byte(s), '>') != -1 {
		fmt.Fprintf(os.Stderr, "Translation contains <...> markers.\n")
		os.Exit(1)
	}
	buf := []byte(fmt.Sprintf(s, n))
	strc := []byte("Vater von 5 Kindern")
	if bytes.Compare(buf, strc) != 0 {
		fmt.Fprintf(os.Stderr, "printf of translation wrong.\n")
		os.Exit(1)
	}
}

