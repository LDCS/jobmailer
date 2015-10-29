package main

import (
	"github.com/LDCS/sflag"
	"fmt"
	"os"
	"os/user"
	"os/exec"
	"strings"
	"io/ioutil"
	"encoding/base64"
)

func usage() {
	fmt.Println("                                                                                                                                                                   ")
        fmt.Println("./jobMailer --Subject <subject> --To <address> --Cc <address> --Bcc <address> --Contents <text file> --Attachment <attachment>                                     ")
	fmt.Println("            --Subject:     subject of the email                  ")
	fmt.Println("            --To:          receivers (comma separated list)      ")
	fmt.Println("            --Cc:          receivers (comma separated list)      ")
	fmt.Println("            --Bcc:         receivers (comma separated list)      ")
	fmt.Println("            --Contents:    a file contains the body of the email ")
	fmt.Println("            --Attachment:  email attachment                      ")
}

var opt = struct {
	Subject     string  "Email subject"
	To          string  "To address"
	Cc          string  "Cc address"
	Bcc         string  "Bcc address"
	Contents    string  "Contents of the email|/dev/null"
	Attachment  string  "Email attachment|/dev/null"
}{}

func main() {
	sflag.Parse(&opt)
	
	usr, err      := user.Current()
	if err != nil {
		fmt.Println("failed to get uid, exiting.....")
		return
	}
	
	hostname,err1 := os.Hostname()
	if err1 != nil {
		fmt.Println("failed to get hostname, exiting.....")
		return
	}
	
	if opt.To == "" && opt.Cc == "" && opt.Bcc == "" {
		usage()
		return
	}

	bodyStr    := ""
	uuid1,_    := exec.Command("uuidgen").Output()
	uuid       := strings.TrimSpace(string(uuid1))
	mpartClose := false  //true -> multipart message (i.e) msg, attachment
	
	bodyStr += "Subject:" + opt.Subject + "\n"
	if opt.To  != "" { bodyStr += "To:"  + opt.To  + "\n" }
	if opt.Cc  != "" { bodyStr += "Cc:"  + opt.Cc  + "\n" }
	if opt.Bcc != "" { bodyStr += "Bcc:" + opt.Bcc + "\n" }

	if opt.Contents != "/dev/null" || opt.Attachment != "/dev/null" {
		bodyStr += "Content-Type: multipart/mixed; boundary=\""+uuid+"\"\nMIME-Version: 1.0\n"
	}
	if opt.Contents != "/dev/null" {
		bodyStr += "\n--" + uuid + "\nContent-Type: text/plain\nContent-Disposition: inline\n\n"
		mpartClose = true
		cStr, err := ioutil.ReadFile(opt.Contents)
		if err != nil {
			panic("Unable to read contents file: " + err.Error())
		}
		bodyStr += string(cStr)
	}

	if opt.Attachment != "/dev/null" {
		bodyStr += ( "\n--" +
			uuid +
			"\nContent-Transfer-Encoding: base64\nContent-Type: application/octet-stream; name=" +
			opt.Attachment +
			"\nContent-Disposition: attachment; filename=" +
			opt.Attachment + "\n\n" )
		mpartClose = true
		cStr, err := ioutil.ReadFile(opt.Attachment)
		if err != nil {
			panic("Unable to read attachment file: " + err.Error())
		}
		bodyStr += base64.StdEncoding.EncodeToString(cStr) + "\n"
	}

	if mpartClose {
		bodyStr += "--" + uuid + "--"
	}
	cmd := exec.Command("/usr/sbin/sendmail", "-t", "-F", usr.Username+"@"+hostname)
	cmd.Stdin = strings.NewReader(bodyStr)
	cmd.Run()
}
