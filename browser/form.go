package browser

import (
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/haruyama/surf/errors"
)

// Submittable represents an element that may be submitted, such as a form.
type Submittable interface {
	Method() string
	Action() string
	SetAction(string)
	Field(name string) (string, bool)
	Input(name, value string) error
	DeleteField(name string) error
	InputSlice(name string, values []string) error
	CheckBox(name string, values []string) error
	Click(button string) error
	Submit() error
	Dom() *goquery.Selection
}

// Form is the default form element.
type Form struct {
	bow           Browsable
	selection     *goquery.Selection
	method        string
	action        string
	definedFields map[string]bool
	fields        url.Values
	buttons       url.Values
}

// NewForm creates and returns a *Form type.
func NewForm(bow Browsable, s *goquery.Selection) *Form {
	definedFields, fields, buttons := serializeForm(s)
	method, action := formAttributes(bow, s)

	return &Form{
		bow:           bow,
		selection:     s,
		method:        method,
		action:        action,
		definedFields: definedFields,
		fields:        fields,
		buttons:       buttons,
	}
}

// Method returns the form method, eg "GET" or "POST".
func (f *Form) Method() string {
	return f.method
}

// Action returns the form action URL.
// The URL will always be absolute.
func (f *Form) Action() string {
	return f.action
}

// SetAction set Action URL.
// The URL will always be absolute.
func (f *Form) SetAction(aurl string) {
	f.action = aurl
}

// Field returns the value of a form field.
func (f *Form) Field(name string) (string, bool) {
	if f.definedFields[name] {
		return f.fields.Get(name), true
	}
	return "", false
}

// Input sets the value of a form field.
func (f *Form) Input(name, value string) error {
	if f.definedFields[name] {
		f.fields.Set(name, value)
		return nil
	}
	return errors.NewElementNotFound(
		"No input found with name '%s'.", name)
}

// DeleteField deletes a form field
func (f *Form) DeleteField(name string) error {
	if f.definedFields[name] {
		f.fields.Del(name)
		return nil
	}
	return errors.NewElementNotFound(
		"No input found with name '%s'.", name)
}

// InputSlice sets the values of a form field.
func (f *Form) InputSlice(name string, values []string) error {
	if f.definedFields[name] {
		f.fields.Del(name)
		for _, v := range values {
			f.fields.Add(name, v)
		}
		return nil
	}
	return errors.NewElementNotFound(
		"No input found with name '%s'.", name)
}

// CheckBox sets the values of a form field.
func (f *Form) CheckBox(name string, values []string) error {
	return f.InputSlice(name, values)
}

// Submit submits the form.
// Clicks the first button in the form, or submits the form without using
// any button when the form does not contain any buttons.
func (f *Form) Submit() error {
	if len(f.buttons) > 0 {
		for name := range f.buttons {
			return f.Click(name)
		}
	}
	return f.send("", "")
}

// Click submits the form by clicking the button with the given name.
func (f *Form) Click(button string) error {
	if _, ok := f.buttons[button]; !ok {
		return errors.NewInvalidFormValue(
			"Form does not contain a button with the name '%s'.", button)
	}
	return f.send(button, f.buttons[button][0])
}

// Dom returns the inner *goquery.Selection.
func (f *Form) Dom() *goquery.Selection {
	return f.selection
}

// send submits the form.
func (f *Form) send(buttonName, buttonValue string) error {
	method, ok := f.selection.Attr("method")
	if !ok {
		method = "GET"
	}
	action := f.action
	if action == "" {
		action, ok = f.selection.Attr("action")
		if !ok {
			action = f.bow.Url().String()
		}
	}
	aurl, err := url.Parse(action)
	if err != nil {
		return err
	}
	aurl = f.bow.ResolveUrl(aurl)

	values := make(url.Values, len(f.fields)+1)
	for name, vals := range f.fields {
		values[name] = vals
	}
	if buttonName != "" {
		values.Set(buttonName, buttonValue)
	}

	if strings.ToUpper(method) == "GET" {
		return f.bow.OpenForm(aurl.String(), values)
	}
	enctype, _ := f.selection.Attr("enctype")
	if enctype == "multipart/form-data" {
		return f.bow.PostMultipart(aurl.String(), values)
	}
	return f.bow.PostForm(aurl.String(), values)
}

// Serialize converts the form fields into a url.Values type.
// Returns two url.Value types. The first is the form field values, and the
// second is the form button values.
func serializeForm(sel *goquery.Selection) (map[string]bool, url.Values, url.Values) {
	input := sel.Find("input,button")
	definedFields := map[string]bool{}
	fields := make(url.Values)
	buttons := make(url.Values)

	input.Each(func(_ int, s *goquery.Selection) {
		name, ok := s.Attr("name")
		if ok {
			typ, ok := s.Attr("type")
			if ok {
				if typ == "submit" {
					val, ok := s.Attr("value")
					if ok {
						buttons.Add(name, val)
					} else {
						buttons.Add(name, "")
					}
				} else if typ == "radio" || typ == "checkbox" {
					definedFields[name] = true
					_, ok := s.Attr("checked")
					if ok {
						val, ok := s.Attr("value")
						if ok {
							fields.Add(name, val)
						}
					}
				} else {
					definedFields[name] = true
					val, ok := s.Attr("value")
					if ok {
						fields.Add(name, val)
					}
				}
			}
		}
	})

	selec := sel.Find("select")

	selec.Each(func(_ int, s *goquery.Selection) {
		name, ok := s.Attr("name")
		if !ok {
			return
		}
		definedFields[name] = true
		s.Find("option[selected]").Each(func(_ int, so *goquery.Selection) {
			val, ok := so.Attr("value")
			if ok {
				fields.Add(name, val)
			}
		})
	})

	textarea := sel.Find("textarea")
	textarea.Each(func(_ int, s *goquery.Selection) {
		name, ok := s.Attr("name")
		if !ok {
			return
		}
		definedFields[name] = true
		fields.Add(name, s.Text())
	})

	return definedFields, fields, buttons
}

func formAttributes(bow Browsable, s *goquery.Selection) (string, string) {
	method, ok := s.Attr("method")
	if !ok {
		method = "GET"
	}
	action, ok := s.Attr("action")
	if !ok {
		action = bow.Url().String()
	}
	aurl, err := url.Parse(action)
	if err != nil {
		return "", ""
	}
	aurl = bow.ResolveUrl(aurl)

	return strings.ToUpper(method), aurl.String()
}
