// Copyright 2012-2015 Oliver Eilhard. All rights reserved.
// Use of this source code is governed by a MIT-license.
// See http://olivere.mit-license.org/license.txt for details.

package metronome

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"regexp"
	"strings"
)

var kBasicAuthPattern = regexp.MustCompile(`^Basic ([a-zA-Z0-9\+/=]+)`)

// BasicAuth parses the Authorization header of the request.
// If absent or invalid, an error is returned.
func BasicAuth(r *http.Request) (username, password string, err error) {
	auth := r.Header.Get("Authorization")
	if auth == "" {
		err = fmt.Errorf("Missing \"Authorization\" in header")
		return
	}
	matches := kBasicAuthPattern.FindStringSubmatch(auth)
	if len(matches) != 2 {
		err = fmt.Errorf("Bogus Authorization header")
		return
	}
	encoded := matches[1]
	enc := base64.StdEncoding
	decBuf := make([]byte, enc.DecodedLen(len(encoded)))
	n, err := enc.Decode(decBuf, []byte(encoded))
	if err != nil {
		return
	}
	pieces := strings.SplitN(string(decBuf[0:n]), ":", 2)
	if len(pieces) != 2 {
		err = fmt.Errorf("didn't get two pieces")
		return
	}
	return pieces[0], pieces[1], nil
}
