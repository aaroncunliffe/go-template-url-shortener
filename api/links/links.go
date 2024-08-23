package links

import "net/http"

// App layer for links
// Does validation before hand off to business layer

func LinkRedirect(w http.ResponseWriter, r *http.Request) {
	// time.Sleep(10 * time.Second)
	w.Write([]byte("OK"))
}
