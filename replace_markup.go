package main

import "strings"

func replaceMarkup(s string) string {
	s = strings.Replace(s, "<b>", "*", -1)
	s = strings.Replace(s, "</b>", "*", -1)
	s = strings.Replace(s, "<em>", "_", -1)
	s = strings.Replace(s, "</em>", "_", -1)
	s = strings.Replace(s, "<i>", "_", -1)
	s = strings.Replace(s, "</i>", "_", -1)
	s = strings.Replace(s, "<u>", "_", -1)
	s = strings.Replace(s, "</u>", "_", -1)
	s = strings.Replace(s, "<br>", "\n", -1)
	s = strings.Replace(s, "<s>", "~", -1)
	s = strings.Replace(s, "</s>", "~", -1)
	s = strings.Replace(s, "<pre>", "`", -1)
	s = strings.Replace(s, "</pre>", "`", -1)
	s = strings.Replace(s, "<blockquote>", "```", -1)
	s = strings.Replace(s, "</blockquote>", "```", -1)
	s = strings.Replace(s, "<p>", "\n", -1)
	s = strings.Replace(s, "</p>", "\n", -1)
	s = strings.Replace(s, "</br>", "\n", -1)
	s = strings.Replace(s, "<br/>", "\n", -1)
	s = strings.Replace(s, "<br />", "\n", -1)
	return s
}

func removeMarkup(s string) string {
	s = strings.Replace(s, "<b>", " ", -1)
	s = strings.Replace(s, "</b>", " ", -1)
	s = strings.Replace(s, "<em>", " ", -1)
	s = strings.Replace(s, "</em>", " ", -1)
	s = strings.Replace(s, "<i>", " ", -1)
	s = strings.Replace(s, "</i>", " ", -1)
	s = strings.Replace(s, "<u>", " ", -1)
	s = strings.Replace(s, "</u>", " ", -1)
	s = strings.Replace(s, "<br>", " ", -1)
	s = strings.Replace(s, "<s>", " ", -1)
	s = strings.Replace(s, "</s>", " ", -1)
	s = strings.Replace(s, "<pre>", " ", -1)
	s = strings.Replace(s, "</pre>", " ", -1)
	s = strings.Replace(s, "<blockquote>", " ", -1)
	s = strings.Replace(s, "</blockquote>", " ", -1)
	s = strings.Replace(s, "<p>", "\n", -1)
	s = strings.Replace(s, "</p>", "\n", -1)
	s = strings.Replace(s, "</br>", "\n", -1)
	s = strings.Replace(s, "<br/>", "\n", -1)
	s = strings.Replace(s, "<br />", "\n", -1)
	return s
}
