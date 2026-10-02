package main

import (
 "encoding/json"
 "log"
 "net/http"
 "os"
 "strings"
 "sync"
 "time"
)

type Notification struct { ID string `json:"id"`; Tenant string `json:"tenant"`; Channel string `json:"channel"`; Recipient string `json:"recipient"`; Template string `json:"template"`; Status string `json:"status"`; CreatedAt time.Time `json:"created_at"` }
type API struct { mu sync.Mutex; requests map[string]Notification }
func (a *API) create(w http.ResponseWriter, r *http.Request) { var n Notification; if json.NewDecoder(http.MaxBytesReader(w,r,4096)).Decode(&n)!=nil || strings.TrimSpace(n.Tenant)=="" || n.Channel!="email" && n.Channel!="webhook" || strings.TrimSpace(n.Recipient)=="" || strings.TrimSpace(n.Template)=="" { http.Error(w,"invalid notification",400); return }; key:=r.Header.Get("Idempotency-Key"); if key==""||len(key)>128 { http.Error(w,"Idempotency-Key is required",400); return }; a.mu.Lock(); defer a.mu.Unlock(); if a.requests==nil {a.requests=map[string]Notification{}}; compound:=n.Tenant+"\x00"+key; if old,ok:=a.requests[compound];ok { if old.Recipient!=n.Recipient||old.Template!=n.Template||old.Channel!=n.Channel {http.Error(w,"idempotency key conflict",409);return}; w.Header().Set("Content-Type","application/json");w.WriteHeader(http.StatusOK);_ = json.NewEncoder(w).Encode(old);return }; n.ID="notification-"+time.Now().UTC().Format("20060102150405.000000000");n.Status="queued";n.CreatedAt=time.Now().UTC();a.requests[compound]=n;w.Header().Set("Content-Type","application/json");w.WriteHeader(http.StatusAccepted);_ = json.NewEncoder(w).Encode(n) }
func main(){ mux:=http.NewServeMux();api:=&API{};mux.HandleFunc("POST /notifications",api.create);mux.HandleFunc("GET /health",func(w http.ResponseWriter,r *http.Request){w.Header().Set("Content-Type","application/json");_,_=w.Write([]byte(`{"status":"ok"}`))});srv:=&http.Server{Addr:":"+env("PORT","8080"),Handler:mux,ReadHeaderTimeout:3*time.Second,WriteTimeout:10*time.Second,IdleTimeout:30*time.Second};log.Printf("notification API listening on %s",srv.Addr);log.Fatal(srv.ListenAndServe()) }
func env(k,d string)string{if v:=os.Getenv(k);v!=""{return v};return d}
