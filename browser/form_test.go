package browser

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/haruyama/surf/jar"
	"github.com/headzoo/ut"
)

func TestBrowserForm(t *testing.T) {
	ut.Run(t)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			fmt.Fprint(w, htmlForm)
		} else {
			r.ParseForm()
			fmt.Fprint(w, r.Form.Encode())
		}
	}))
	defer ts.Close()

	bow := &Browser{}
	bow.headers = make(http.Header, 10)
	bow.history = jar.NewMemoryHistory()

	err := bow.Open(ts.URL)
	ut.AssertNil(err)

	f, err := bow.Form("[name='default']")
	ut.AssertNil(err)

	f.Input("age", "55")
	f.Input("gender", "male")
	err = f.Click("submit2")
	ut.AssertNil(err)
	ut.AssertContains("age=55", bow.Body())
	ut.AssertContains("gender=male", bow.Body())
	ut.AssertContains("submit2=submitted2", bow.Body())
}

var htmlForm = `<!doctype html>
<html>
	<head>
		<title>Echo Form</title>
	</head>
	<body>
		<form method="post" action="/" name="default">
			<input type="text" name="age" value="" />
			<input type="radio" name="gender" value="male" />
			<input type="radio" name="gender" value="female" />
			<input type="submit" name="submit1" value="submitted1" />
			<input type="submit" name="submit2" value="submitted2" />
		</form>
	</body>
</html>
`

func TestBrowserForm2(t *testing.T) {
	ut.Run(t)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			fmt.Fprint(w, htmlForm2)
		} else {
			r.ParseForm()
			fmt.Fprint(w, r.Form.Encode())
		}
	}))
	defer ts.Close()

	bow := &Browser{}
	bow.headers = make(http.Header, 10)
	bow.history = jar.NewMemoryHistory()

	err := bow.Open(ts.URL)
	ut.AssertNil(err)

	f, err := bow.Form("[name='default']")
	ut.AssertNil(err)

	err = f.Input("age", "54")
	ut.AssertNil(err)
	err = f.Input("agee", "54")
	ut.AssertNotNil(err)

	err = f.CheckBox("music", []string{"rock", "fusion"})
	ut.AssertNil(err)

	err = f.CheckBox("music2", []string{"rock", "fusion"})
	ut.AssertNotNil(err)

	err = f.Click("submit2")

	ut.AssertNil(err)
	ut.AssertContains("company=none", bow.Body())
	ut.AssertContains("age=54", bow.Body())
	ut.AssertContains("gender=male", bow.Body())

	ut.AssertFalse(strings.Contains(bow.Body(), "music=jazz"))
	ut.AssertContains("music=rock", bow.Body())
	ut.AssertContains("music=fusion", bow.Body())

	ut.AssertContains("hobby=Dance", bow.Body())
	ut.AssertContains("city=NY", bow.Body())
	ut.AssertContains("submit2=submitted2", bow.Body())
}

var htmlForm2 = `<!doctype html>
<html>
	<head>
		<title>Echo Form</title>
	</head>
	<body>
		<form method="post" action="/" name="default">
			<input type="text" name="will_be_deleted" value="aaa">
			<input type="text" name="company" value="none">
			<input type="text" name="age" value="55">
			<input type="radio" name="gender" value="male" checked>
			<input type="radio" name="gender" value="female">
			<input type="checkbox" name="music" value="jazz" checked="checked">
			<input type="checkbox" name="music" value="rock">
			<input type="checkbox" name="music" value="fusion" checked>
			<select name="city">
				<option value="NY" selected>
				<option value="Tokyo">
			</select>
			<textarea name="hobby">Dance</textarea>
			<input type="submit" name="submit2" value="submitted2">
		</form>
	</body>
</html>
`
