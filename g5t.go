// Internationalization and localization support.

package g5t

// Copyright 2012 G.vd.Schoot. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// See README file for usage.

import (
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"path"
	"strings"
)

type Parser func(fp *os.File) error

// Override this method to support alternative .mo formats.
func GettextParser(fp *os.File) error {

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
		return errors.New(fmt.Sprint("Bad magic number. File: ", filename))
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
			return errors.New(fmt.Sprint("File is corrupt. File: ", filename))
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
	return nil
}

func Setup(domain, localedir, language string, parser Parser) error {

	// Select the language file
	mofile := path.Join(localedir, language, "LC_MESSAGES", fmt.Sprintf("%s.mo", domain))

	// Opening, reading, and parsing the .mo file.
	fp, err := os.Open(mofile)
	if err == nil {
		parser(fp)
		return nil
	}

	// else 
	return err
}

//
// catalog with translation strings
//
type CatalogType map[string]string

var Catalog CatalogType = make(CatalogType, 100)

func String(message string) string {
	tmsg := Catalog[message]
	if tmsg == "" {
		return message
	}
	return tmsg
}

func StringN(msgid1, msgid2 string, n int) (tmsg string) {
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
