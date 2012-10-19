package g5t

// Copyright 2012 G.vd.Schoot. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.


/* Internationalization and localization support.

This module provides internationalization (I18N) and localization (L10N)
support for your Go programs by providing an interface to the GNU gettext
message catalog library.

I18N refers to the operation by which a program is made aware of multiple
languages.  L10N refers to the adaptation of your program, once
internationalized, to the local language and cultural habits.

This module is a <<< subset >>> and rewrite of the Python gettext module 
in the Go language.
The main difference between this module and the Python gettext module is that
this one only uses unicode AND with only one language file at a time (which is 
good enough for most programs). This approach reduces the code size 
significantly.


This module contains only ---- FOUR ---- exported functions:

	Parse()		// This is the ".mo" file parser and is overridable
	Setup()		// This function sets the translation
	String()	// This function returns a translated string
	StringN()	// This function returns a translated string in plural form


Usage (short):

	1) To use this package in your application, write this line (it's 
	   go-installable): 

		import "i18n/gt"


	2) In your application define functions (for singualar and plural):

		func G(msg string) string { return gt.String(msg) }
		func GN(msgid1, msgid2 string, n int) string { return gt.StringN(msgid1, msgid2, n) }


	4) Code the rest of your program. When you want a string that needs to be 
	   translated, write G("string to translate") etc.

	5) Run for each ".go" file:

		xgettext -o messages.po -C -kG -kGN:1,2 yourprogramfile.go


	6) If you want to place the translation files to a sub directory of 
	   your application, make a directory called "translations" (for example)
	   To locate your translations place this function in your app:

		gt.Setup("your_app", "directory_to_transl_files", "language", Parser)


	7) Translate "messages.po" with a text editor to a specific language and
	   save the file to "yourdomain.po". Then run

	   msgfmt yourdomain.po


	   Msgfmt creates "yourdomain.mo". Copy this file to the language
	   translation directory (in Debian this is "/usr/share/locale/")

	8) That's it.


TODO:
- Bugfixing. 
- Make it work on different operating systems.
- Listen to input from others. 

*/

import (
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"io/ioutil"
	"path"
	"strings"
)

type Parser func(fp *os.File) (CatalogType, error)

//
// catalog with translation strings
//
type CatalogType map[string]string

// Override this method to support alternative .mo formats.
func GettextParser(fp *os.File) (CatalogType, error) {
	var Catalog CatalogType = make(CatalogType, 100)

	// Magic number of .mo files; 32 bits; Used for Little/Big Endian testing
	const LE_MAGIC = 0x950412de
	const BE_MAGIC = 0xde120495

	filename := fp.Name()
	// Parse the .mo file header, which consists of 5 little endian 32
	// bit words.
	fpstat, _ := fp.Stat()
	buflen := fpstat.Size()

	// Are we big endian or little endian?
	var magic uint32
	var ii binary.ByteOrder
	_ = binary.Read(fp, binary.LittleEndian, &magic)
	if magic == LE_MAGIC {
		ii = binary.LittleEndian
	} else if magic == BE_MAGIC {
		ii = binary.BigEndian
	} else {
		return nil, errors.New(fmt.Sprint("Bad magic number. File: ", filename))
	}

	var version, msgcount, masteridx, transidx uint32
	_ = binary.Read(fp, ii, &version)
	_ = binary.Read(fp, ii, &msgcount)
	_ = binary.Read(fp, ii, &masteridx)
	_ = binary.Read(fp, ii, &transidx)

	// Now put all messages from the .mo file buffer into the catalog
	// dictionary.
	for i := 0; i < int(msgcount); i++ {
		var mlen, moff, mend, tlen, toff, tend uint32
		fp.Seek(int64(masteridx), 0)
		binary.Read(fp, ii, &mlen)
		binary.Read(fp, ii, &moff)
		mend = moff + mlen
		fp.Seek(int64(transidx), 0)
		_ = binary.Read(fp, ii, &tlen)
		_ = binary.Read(fp, ii, &toff)
		tend = toff + tlen
		msg := make([]byte, mlen)
		tmsg := make([]byte, tlen)
		if int64(mend) < buflen && int64(tend) < buflen {
			_, _ = fp.ReadAt(msg, int64(moff))
			_, _ = fp.ReadAt(tmsg, int64(toff))
		} else {
			return nil, errors.New(fmt.Sprint("File is corrupt. File: ", filename))
		}

		if strings.Index(string(msg), "\x00") >= 0 {
			// Plural forms
			msgid12 := strings.Split(string(msg), "\x00")
			tmsg12 := strings.Split(string(tmsg), "\x00")
			for i := 0; i < len(tmsg); i++ {
				Catalog[msgid12[i]] = tmsg12[i]
			}
		} else {
			Catalog[string(msg)] = string(tmsg)
		}
		// advance to next entry in the seek tables
		masteridx += 8
		transidx += 8
	}
	return Catalog, nil
}

func LoadLang(domain, localedir, language string, parser Parser) (CatalogType, error) {

	// Select the language file
	mofile := path.Join(localedir, language, "LC_MESSAGES", fmt.Sprintf("%s.mo", domain))

	// Opening, reading, and parsing the .mo file.
	fp, err := os.Open(mofile)
	if err == nil {
		return parser(fp)
	}

	// else 
	return nil, err
}

func LoadAllLangs(domain, localedir string, parser Parser) (map[string]CatalogType)  {
	dirs, _ := ioutil.ReadDir(localedir)
	ret := make(map[string]CatalogType)
	for _, fileInfo := range dirs {
		catalog, err := LoadLang(domain, localedir, fileInfo.Name(), parser)
		if err == nil {
			ret[fileInfo.Name()] = catalog
		} else {
			fmt.Println("Error parsing locale", fileInfo.Name(), "-", err)
		}
	}
	return ret
}

//
// catalog with translation strings
//
func (Catalog CatalogType) String(message string) string {
	tmsg := Catalog[message]
	if tmsg == "" {
			return message
	}
	return tmsg
}

func (Catalog CatalogType) StringPartial() func(string) string {
	return func(message string) string { return Catalog.String(message) }
}

func (Catalog CatalogType) StringN(msgid1, msgid2 string, n int) (tmsg string) {
	//
	// if n != 1 then it is plural
	// 
	if n == 1 {
		tmsg = Catalog[msgid1]
	} else {
		tmsg = Catalog[msgid2]
	}
	if tmsg == "" {
		if n == 1 {
			tmsg = msgid1
		} else {
			tmsg = msgid2
		}
	}
	return tmsg
}

func (Catalog CatalogType) StringNPartial() func(string, string, int) string {
	return func(msg1 string, msg2 string, n int) string {
		return Catalog.StringN(msg1, msg2, n)
	}
}

