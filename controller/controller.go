package controller

import (
	"net/http"
	"fmt"
	"github.com/alfred-zhong/wserver"
	"bytes"
	"../session"
	"encoding/json"
	"html/template"
	"strings"
	"reflect"
	"os"
)



var pushURL = "http://127.0.0.1:3001/push"		// push msg




func (this *dispatchHandler)IndexAction(w http.ResponseWriter, r *http.Request) {
	ses := session.GetSession(w, r)
	t := template.Must(template.ParseFiles("index.html"))
	t.Execute(w, ses.Id())
}

func (this *dispatchHandler)PingAction(w http.ResponseWriter, r *http.Request) {
	ses := session.GetSession(w, r)
	fmt.Fprintf(w, "session id=%s, foo=%s", ses.Id(), ses.Get("foo"))
	val := r.FormValue("v")
	if val == "" {
		val = "default"
	}
	ses.Set("foo", val)
	sendMsg(val,ses.Id())
}

func (this *dispatchHandler)Error404Action(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "This is 404 page")
}

func (this *dispatchHandler)StaticAction(w http.ResponseWriter, r *http.Request) {
	file :=  r.URL.Path[len("/"):]
	f,err := os.Open(file)
	defer f.Close()

	if err != nil && os.IsNotExist(err){
		fmt.Fprintf(w, "This is 404 page")
	}else{
		http.ServeFile(w,r,file)
	}
}

// msg 需要传送的消息， topic 主题，也可以是sessionId
func sendMsg(msg string,topic string)  {
	contentType := "application/json"
	pm := wserver.PushMessage{
		UserID:  "jack",
		Event:   topic,
		Message: msg,
	}
	b, _ := json.Marshal(pm)
	http.DefaultClient.Post(pushURL, contentType, bytes.NewReader(b))
}

type dispatchHandler struct {}

func Dispatcher(w http.ResponseWriter, r *http.Request) {
	pathInfo := strings.Trim(r.URL.Path, "/")
	parts := strings.Split(pathInfo, "/")
	var action = ""
	var size = len(parts)
	if size > 0  && len(parts[0]) > 0{
		if strings.HasPrefix(parts[0],"static" ) {
			action = "StaticAction"
		}else{
			action = strings.Title(parts[0]) + "Action"
		}
	}else{
		action = "Error404Action"
	}
	subHandler := &dispatchHandler{}
	controller := reflect.ValueOf(subHandler)
	method := controller.MethodByName(action)
	if !method.IsValid() {
		method = controller.MethodByName("Error404Action")
	}
	requestValue := reflect.ValueOf(r)
	responseValue := reflect.ValueOf(w)
	method.Call([]reflect.Value{responseValue, requestValue})
}