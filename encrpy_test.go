package main

import (
	"log"
	"testing"
)

func TestEncrypt(t *testing.T) {
	secretMessage := "shh! It's a secret!"
	key := "QJ$j!U1mFuWKmZV*Nuj$5Mikq2WhYEXg"
	crypted := encrypt([]byte(key), secretMessage)
	log.Println(crypted)
	decrypted := decrypt([]byte(key), crypted)
	if decrypted != secretMessage {
		t.Fail()
	}

}
