package web_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"handmade-soap-shop/internal/fixture"
	shopweb "handmade-soap-shop/internal/web"
)

func TestVisitorBrowsesStoriesScentsAndPrices(t *testing.T) {
	handler := newHandler()
	home := request(handler, http.MethodGet, "/", nil, nil)
	assertStatus(t, home, http.StatusOK)
	assertContains(t, home.Body.String(), "林间清晨")
	assertContains(t, home.Body.String(), "雪松、松针与薄荷")
	assertContains(t, home.Body.String(), "¥48.00")

	detail := request(handler, http.MethodGet, "/soaps/citrus-tea", nil, nil)
	assertStatus(t, detail, http.StatusOK)
	assertContains(t, detail.Body.String(), "灵感来自午后茶席")
	assertContains(t, detail.Body.String(), "甜橙、佛手柑与红茶")
	assertContains(t, detail.Body.String(), "¥52.00")
}

func TestRegisteredMemberLogsInReadsGuideAndLogsOut(t *testing.T) {
	handler := newHandler()

	registration := request(handler, http.MethodPost, "/register", url.Values{
		"name": {"周岚"}, "email": {"zhou@example.com"}, "password": {"eight888"},
	}, nil)
	assertStatus(t, registration, http.StatusSeeOther)
	assertLocation(t, registration, "/login?registered=1")

	login := request(handler, http.MethodPost, "/login", url.Values{
		"email": {"member@example.com"}, "password": {"soap1234"},
	}, nil)
	assertStatus(t, login, http.StatusSeeOther)
	assertLocation(t, login, "/member")
	cookie := responseCookie(t, login)

	member := request(handler, http.MethodGet, "/member", nil, cookie)
	assertStatus(t, member, http.StatusOK)
	assertContains(t, member.Body.String(), "林青，欢迎回来")
	assertContains(t, member.Body.String(), "你的秋季用皂说明")

	logout := request(handler, http.MethodPost, "/logout", nil, cookie)
	assertStatus(t, logout, http.StatusSeeOther)
	assertLocation(t, logout, "/")

	home := request(handler, http.MethodGet, "/", nil, nil)
	assertStatus(t, home, http.StatusOK)
	assertContains(t, home.Body.String(), "一块皂，一段慢下来的故事")

	memberAfterLogout := request(handler, http.MethodGet, "/member", nil, cookie)
	assertStatus(t, memberAfterLogout, http.StatusSeeOther)
	assertLocation(t, memberAfterLogout, "/login")
}

func TestMemberWithoutPersonalGuideSeesEmptyState(t *testing.T) {
	handler := newHandler()
	registration := request(handler, http.MethodPost, "/register", url.Values{
		"name": {"周岚"}, "email": {"zhou@example.com"}, "password": {"eight888"},
	}, nil)
	assertStatus(t, registration, http.StatusSeeOther)

	login := request(handler, http.MethodPost, "/login", url.Values{
		"email": {"zhou@example.com"}, "password": {"eight888"},
	}, nil)
	cookie := responseCookie(t, login)

	var response *httptest.ResponseRecorder
	var panicValue any
	func() {
		defer func() { panicValue = recover() }()
		response = request(handler, http.MethodGet, "/member", nil, cookie)
	}()
	if panicValue != nil {
		t.Fatalf("member page panicked: %v", panicValue)
	}
	assertStatus(t, response, http.StatusOK)
	assertContains(t, response.Body.String(), "尚无专属说明")
}

func TestInvalidLoginAndDuplicateRegistrationStayOnForms(t *testing.T) {
	handler := newHandler()
	login := request(handler, http.MethodPost, "/login", url.Values{
		"email": {"member@example.com"}, "password": {"incorrect"},
	}, nil)
	assertStatus(t, login, http.StatusUnprocessableEntity)
	assertContains(t, login.Body.String(), "邮箱或密码不正确")

	registration := request(handler, http.MethodPost, "/register", url.Values{
		"name": {"林青"}, "email": {"member@example.com"}, "password": {"soap1234"},
	}, nil)
	assertStatus(t, registration, http.StatusUnprocessableEntity)
	assertContains(t, registration.Body.String(), "这个邮箱已经注册")
}

func newHandler() http.Handler {
	return shopweb.NewHandler(fixture.NewService()).Routes()
}

func request(handler http.Handler, method, target string, form url.Values, cookie *http.Cookie) *httptest.ResponseRecorder {
	var body io.Reader
	if form != nil {
		body = strings.NewReader(form.Encode())
	}
	req := httptest.NewRequest(method, target, body)
	if form != nil {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	if cookie != nil {
		req.AddCookie(cookie)
	}
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, req)
	return response
}

func responseCookie(t *testing.T, response *httptest.ResponseRecorder) *http.Cookie {
	t.Helper()
	cookies := response.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("response did not set a session cookie")
	}
	return cookies[0]
}

func assertStatus(t *testing.T, response *httptest.ResponseRecorder, want int) {
	t.Helper()
	if response.Code != want {
		t.Fatalf("status = %d, want %d; body = %s", response.Code, want, response.Body.String())
	}
}

func assertLocation(t *testing.T, response *httptest.ResponseRecorder, want string) {
	t.Helper()
	if got := response.Header().Get("Location"); got != want {
		t.Fatalf("location = %q, want %q", got, want)
	}
}

func assertContains(t *testing.T, body, want string) {
	t.Helper()
	if !strings.Contains(body, want) {
		t.Fatalf("response does not contain %q", want)
	}
}
