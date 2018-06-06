package essen

import (
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
)

type GetBody struct {
	body *url.URL
}

type PostBody struct {
	body *http.Request
}

type MultiPartBody struct {
	body *http.Request
}

type Param interface {
	Params(name string) (string, EssenError)
}

func (b GetBody) Params(name string) (string, EssenError) {
	v := b.body.Query().Get(name)
	ee := EssenError{nilval: true}
	if v == "" {
		ee.nilval = false
		ee.errortype = "InvalidParam"
		ee.message = `No parameter with key "` + name + `"`
	}
	return v, ee
}

func (b PostBody) Params(name string) (string, EssenError) {
	ee := EssenError{nilval: true}
	v := b.body.PostFormValue(name)
	if v == "" {
		ee.nilval = false
		ee.errortype = "InvalidParam"
		ee.message = `No parameter with key "` + name + `"`
		return "", ee
	}
	return v, ee
}

func (b MultiPartBody) Params(name string) (string, EssenError) {
	ee := EssenError{nilval: true}
	file, fileHeader, err := b.body.FormFile(name)
	if err != nil && err.Error() != "http: no such file" {
		ee.nilval = false
		ee.errortype = "FormParseError"
		ee.message = err.Error()
		return "", ee
	} else if err == nil {
		UploadDir := MultiPartConfig["UploadDir"]
		path := UploadDir + "/" + fileHeader.Filename
		f, ee := CreateFileIfNotExist(path)
		if !ee.IsNil() {
			return "", ee
		}
		n, err := io.Copy(f, file)
		log.Println(n, err)
		return path, ee
	}
	v := b.body.FormValue(name)
	if v == "" {
		ee.nilval = false
		ee.errortype = "InvalidParam"
		ee.message = `No parameter with key "` + name + `"`
		return "", ee
	}
	return v, ee
}

func (r Request) Path() string {
	return r.Req.URL.Path
}

func (r Request) Host() string {
	return r.Req.URL.Host
}

func (r Request) Method() string {
	return r.Req.Method
}

func (r Request) HasHeader(key string) bool {
	v := r.Req.Header.Get(key)
	if v == "" {
		return false
	}
	return true
}

func (r Request) Header(key string) (string, EssenError) {
	if r.HasHeader(key) {
		hval := r.Req.Header.Get(key)
		ok := strings.HasPrefix(hval, "multipart")
		if ok {
			hval = strings.Split(hval, ";")[0]
		}
		return hval, EssenError{nilval: true}
	}
	return "", EssenError{message: "No Header Found", errortype: "NoHeader", nilval: false}
}

func (r *Request) requestBody() {
	contentType, ee := r.Header("Content-Type")
	if !ee.IsNil() {
		log.Panic(ee.Error())
	}
	if contentType == "multipart/form-data" {
		if !mConfigIsSet() {
			setDefaultConfig()
		}
		r.Body = MultiPartBody{body: r.Req}
		return
	}
	if r.Method() == "GET" || r.Method() == "HEAD" {
		r.Body = GetBody{body: r.Req.URL}
		return
	}
	if r.Method() == "POST" {
		err := r.Req.ParseForm()
		if err != nil {
			ee := EssenError{nilval: false, errortype: "FormParseError", message: err.Error()}
			log.Panic(ee.Error())
		}
		r.Body = PostBody{body: r.Req}
		return
	}
}
