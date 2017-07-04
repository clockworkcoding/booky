package main

import (
	"fmt"
	"testing"
)

func TestShortenUrl(t *testing.T) {
	config := readConfig()
	if config.BitlyKey == "{your key here}" {
		t.Skip()
	}
	url := "https://www.amazon.com/Heir-Empire-Timothy-Zahn/dp/B007CJZREQ?SubscriptionId=AKIAIOVA54MGWSKB7FTA&tag=booky07-20&linkCode=xm2&camp=2025&creative=165953&creativeASIN=B007CJZREQ"
	if shortURL := shortenURl(url); shortURL == url {
		t.Fail()
	} else {
		fmt.Println(shortURL)
	}

}
