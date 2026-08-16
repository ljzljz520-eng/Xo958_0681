package web

import (
	"bytes"
	"embed"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"time"

	"handmade-soap-shop/internal/shop"
)

const sessionCookie = "soap_session"

//go:embed templates/*.html
var templateFiles embed.FS

type Handler struct {
	service   *shop.Service
	templates map[string]*template.Template
}

type pageData struct {
	Title      string
	Viewer     *shop.Account
	Soaps      []shop.Soap
	Soap       shop.Soap
	Member     shop.MemberPage
	Action     string
	Error      string
	Registered bool
	Values     url.Values
}

func NewHandler(service *shop.Service) *Handler {
	functions := template.FuncMap{"money": func(cents int) string { return fmt.Sprintf("¥%.2f", float64(cents)/100) }}
	pages := make(map[string]*template.Template)
	for _, name := range []string{"home", "soap", "auth", "member"} {
		pages[name] = template.Must(template.New("layout.html").Funcs(functions).ParseFS(templateFiles, "templates/layout.html", "templates/"+name+".html"))
	}
	h := &Handler{service: service, templates: pages}
	return h
}

func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", h.home)
	mux.HandleFunc("GET /soaps/{slug}", h.soap)
	mux.HandleFunc("GET /register", h.registerForm)
	mux.HandleFunc("POST /register", h.register)
	mux.HandleFunc("GET /login", h.loginForm)
	mux.HandleFunc("POST /login", h.login)
	mux.HandleFunc("GET /member", h.member)
	mux.HandleFunc("POST /logout", h.logout)
	return mux
}

func (h *Handler) home(w http.ResponseWriter, r *http.Request) {
	h.render(w, http.StatusOK, "home", pageData{Title: "手工皂", Viewer: h.viewer(r), Soaps: h.service.Soaps()})
}

func (h *Handler) soap(w http.ResponseWriter, r *http.Request) {
	item, err := h.service.Soap(r.PathValue("slug"))
	if errors.Is(err, shop.ErrNotFound) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		http.Error(w, "暂时无法读取商品", http.StatusInternalServerError)
		return
	}
	h.render(w, http.StatusOK, "soap", pageData{Title: item.Name, Viewer: h.viewer(r), Soap: item})
}

func (h *Handler) registerForm(w http.ResponseWriter, r *http.Request) {
	h.render(w, http.StatusOK, "auth", pageData{Title: "注册会员", Action: "/register", Values: make(url.Values)})
}

func (h *Handler) register(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "无法读取表单", http.StatusBadRequest)
		return
	}
	err := h.service.Register(r.FormValue("name"), r.FormValue("email"), r.FormValue("password"))
	if err == nil {
		http.Redirect(w, r, "/login?registered=1", http.StatusSeeOther)
		return
	}
	message := "请填写姓名、有效邮箱和至少 8 位密码"
	if errors.Is(err, shop.ErrEmailInUse) {
		message = "这个邮箱已经注册"
	}
	h.render(w, http.StatusUnprocessableEntity, "auth", pageData{Title: "注册会员", Action: "/register", Error: message, Values: r.Form})
}

func (h *Handler) loginForm(w http.ResponseWriter, r *http.Request) {
	h.render(w, http.StatusOK, "auth", pageData{Title: "会员登录", Action: "/login", Registered: r.URL.Query().Get("registered") == "1", Values: make(url.Values)})
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "无法读取表单", http.StatusBadRequest)
		return
	}
	token, err := h.service.Login(r.FormValue("email"), r.FormValue("password"))
	if err != nil {
		h.render(w, http.StatusUnprocessableEntity, "auth", pageData{Title: "会员登录", Action: "/login", Error: "邮箱或密码不正确", Values: r.Form})
		return
	}
	http.SetCookie(w, &http.Cookie{Name: sessionCookie, Value: token, Path: "/", HttpOnly: true, SameSite: http.SameSiteLaxMode})
	http.Redirect(w, r, "/member", http.StatusSeeOther)
}

func (h *Handler) member(w http.ResponseWriter, r *http.Request) {
	member, err := h.service.Member(h.sessionToken(r))
	if errors.Is(err, shop.ErrUnauthenticated) {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	if err != nil {
		http.Error(w, "暂时无法读取会员说明", http.StatusInternalServerError)
		return
	}
	h.render(w, http.StatusOK, "member", pageData{Title: "会员说明", Member: member})
}

func (h *Handler) logout(w http.ResponseWriter, r *http.Request) {
	token := h.sessionToken(r)
	if token != "" {
		h.service.Logout(token)
	}
	http.SetCookie(w, &http.Cookie{Name: sessionCookie, Value: "", Path: "/", HttpOnly: true, SameSite: http.SameSiteLaxMode, MaxAge: -1, Expires: time.Unix(1, 0)})
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *Handler) viewer(r *http.Request) *shop.Account {
	account, ok := h.service.Viewer(h.sessionToken(r))
	if !ok {
		return nil
	}
	return &account
}

func (h *Handler) sessionToken(r *http.Request) string {
	cookie, err := r.Cookie(sessionCookie)
	if err != nil {
		return ""
	}
	return cookie.Value
}

func (h *Handler) render(w http.ResponseWriter, status int, name string, data pageData) {
	var output bytes.Buffer
	if err := h.templates[name].ExecuteTemplate(&output, "layout", data); err != nil {
		http.Error(w, "页面渲染失败", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	_, _ = w.Write(output.Bytes())
}
