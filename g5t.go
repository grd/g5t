// Internationalization and localization support.
// NOT intended for productional use! See README.
package g5t

// Copyright 2012 G.vd.Schoot. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
)

type Parser func(fp *os.File) error

// Override this method to support alternative .mo formats.
func GettextParser(fp *os.File) error {

	// Parse the .mo file header, which consists of 5 little endian 32
	// bit words.

	type Mo_header struct {
		Magic, Version, Msgcount, Masteridx, Transidx uint32
	}
	var header Mo_header

	err := binary.Read(fp, binary.LittleEndian, &header)
	if err != nil {
		return err
	}

	// Magic number of .mo files; 32 bits;
	if header.Magic != 0x950412de {
		return errors.New(fmt.Sprint("Bad magic number. File: ", fp.Name()))
	}

	// Master- and translation indexes
	type Index struct{ Len, Off uint32 }
	mIndex := make([]Index, header.Msgcount)
	tIndex := make([]Index, header.Msgcount)

	// Read master index
	fp.Seek(int64(header.Masteridx), 0)
	err = binary.Read(fp, binary.LittleEndian, &mIndex)
	if err != nil {
		return err
	}

	// Read translation index
	fp.Seek(int64(header.Transidx), 0)
	err = binary.Read(fp, binary.LittleEndian, &tIndex)
	if err != nil {
		return err
	}

	// Now put all messages from the .mo file buffer into the catalog
	// dictionary.
	for i := 0; i < int(header.Msgcount); i++ {

		// Read message string
		msg := make([]byte, mIndex[i].Len)
		fp.Seek(int64(mIndex[i].Off), 0)
		_, err = io.ReadFull(fp, msg)
		if err != nil {
			return err
		}

		// Read translation string
		tmsg := make([]byte, tIndex[i].Len)
		fp.Seek(int64(tIndex[i].Off), 0)
		_, err = io.ReadFull(fp, tmsg)
		if err != nil {
			return err
		}

		// Adding the messages to the catalog...
		// First check for plural forms
		if mir := bytes.IndexRune(msg, '\x00'); mir >= 0 {
			tir := bytes.IndexRune(tmsg, '\x00')
			Catalog[string(msg[0:mir])] = string(tmsg[0:tir])
			Catalog[string(msg[mir+1:])] = string(tmsg[tir+1:])
		} else {
			Catalog[string(msg)] = string(tmsg)
		}
	}
	return nil
}

func Setup(domain, localedir, language string, parser Parser) error {

	// Select the language file
	mofile := path.Join(localedir, language, "LC_MESSAGES", fmt.Sprintf("%s.mo", domain))

	// Opening, reading, and parsing the .mo file.
	fp, err := os.Open(mofile)
	if err != nil {
		return err
	}
	defer fp.Close()

	err = parser(fp)
	if err != nil {
		return fmt.Sprintf("Error parsing %s. Error: %s\n", fp.Name(), err)
	}
	return nil
}

//
// catalog with translation strings
//
type CatalogType map[string]string

var Catalog CatalogType = make(CatalogType)

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
