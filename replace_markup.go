package main

import "strings"

func replaceMarkup(s string) string {
	replacer := strings.NewReplacer(
		"<b>", "*",
		"</b>", "*",
		"<strong>", "*",
		"</strong>", "*",
		"<em>", "_",
		"</em>", "_",
		"<i>", "_",
		"</i>", "_",
		"<u>", "_",
		"</u>", "_",
		"<br>", "\n",
		"<s>", "~",
		"</s>", "~",
		"<pre>", "`",
		"</pre>", "`",
		"<blockquote>", "```",
		"</blockquote>", "```",
		"<p>", "\n",
		"</p>", "\n",
		"</br>", "\n",
		"<br/>", "\n",
		"<br />", "\n")
	return replacer.Replace(s)
}

func removeMarkup(s string) string {
	replacer := strings.NewReplacer(
		"<b>", " ",
		"</b>", " ",
		"<strong>", " ",
		"</strong>", " ",
		"<em>", " ",
		"</em>", " ",
		"<i>", " ",
		"</i>", " ",
		"<u>", " ",
		"</u>", " ",
		"<br>", " ",
		"<s>", " ",
		"</s>", " ",
		"<pre>", " ",
		"</pre>", " ",
		"<blockquote>", "\n",
		"</blockquote>", "\n",
		"<p>", "\n",
		"</p>", "\n",
		"</br>", "\n",
		"<br/>", "\n",
		"<br />", "\n")
	return replacer.Replace(s)
}
